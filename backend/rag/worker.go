package rag

import (
	"context"
	"fmt"
	"log"
	"math"

	"coin-wave/models"

	"github.com/cloudwego/eino/compose"
)

// IngestionInput is the input for the ingestion workflow.
type IngestionInput struct {
	ArticleID uint
}

// IngestionOutput is the output for the ingestion workflow.
type IngestionOutput struct {
	ArticleID uint
	Status    string
	Error     error
}

type ArticleWithChunks struct {
	Article *models.Article
	Chunks  []string
}

type ArticleWithEmbeddings struct {
	Article *models.Article
	Chunks  []ChunkData
}

// IngestionWorker manages the Eino graph for article ingestion.
type IngestionWorker struct {
	runnable compose.Runnable[IngestionInput, IngestionOutput]
	service  *RagService
}

func NewIngestionWorker(service *RagService) (*IngestionWorker, error) {
	// 1. Create Graph
	g := compose.NewGraph[IngestionInput, IngestionOutput]()

	// 2. Define Nodes

	// Node 1: Fetch Article
	fetchNode := compose.InvokableLambda(func(ctx context.Context, input IngestionInput) (*models.Article, error) {
		log.Printf("[Worker] Fetching article %d", input.ArticleID)
		var article models.Article
		if err := service.db.First(&article, input.ArticleID).Error; err != nil {
			log.Printf("[Worker] Failed to fetch article %d: %v", input.ArticleID, err)
			return nil, err
		}
		service.db.Model(&article).Updates(map[string]interface{}{
			"vector_status":   "processing",
			"vector_progress": 10,
		})
		return &article, nil
	})

	// Node 2: Chunking
	chunkNode := compose.InvokableLambda(func(ctx context.Context, article *models.Article) (*ArticleWithChunks, error) {
		log.Printf("[Worker] Chunking article %d", article.ID)
		chunks := service.chunkText(article.Content, ChunkSize, ChunkOverlap)
		log.Printf("[Worker] Generated %d chunks for article %d", len(chunks), article.ID)
		service.db.Model(article).Update("vector_progress", 20)
		return &ArticleWithChunks{Article: article, Chunks: chunks}, nil
	})

	// Node 3: Embedding
	embedNode := compose.InvokableLambda(func(ctx context.Context, input *ArticleWithChunks) (*ArticleWithEmbeddings, error) {
		log.Printf("[Worker] Embedding %d chunks for article %d", len(input.Chunks), input.Article.ID)
		var textsToEmbed []string
		var chunkDataList []ChunkData

		for i, content := range input.Chunks {
			text := fmt.Sprintf("%s %s\n%s", input.Article.Title, input.Article.Tags, content)
			textsToEmbed = append(textsToEmbed, text)
			chunkDataList = append(chunkDataList, ChunkData{
				UserID:     input.Article.AuthorID,
				ArticleID:  input.Article.ID,
				ChunkIndex: i,
				Content:    content,
			})
		}

		// Batch Embedding
		batchSize := 10
		totalChunks := len(textsToEmbed)
		
		for i := 0; i < totalChunks; i += batchSize {
			end := i + batchSize
			if end > totalChunks {
				end = totalChunks
			}

			batchTexts := textsToEmbed[i:end]
			embeddings, err := service.volcClient.GetEmbeddings(batchTexts)
			if err != nil {
				log.Printf("Embedding failed for chunk batch %d-%d: %v", i, end, err)
				return nil, err
			}

			// If embeddings are fewer than requested (should not happen with robust client)
			if len(embeddings) < len(batchTexts) {
				log.Printf("Warning: received fewer embeddings (%d) than requested (%d)", len(embeddings), len(batchTexts))
			}

			for j, emb := range embeddings {
				if i+j < len(chunkDataList) {
					chunkDataList[i+j].Embedding = emb
				}
			}

			// Update Progress: 20 -> 90
			progress := 20 + int(math.Round(float64(end)/float64(totalChunks)*70))
			service.db.Model(input.Article).Update("vector_progress", progress)
		}

		return &ArticleWithEmbeddings{Article: input.Article, Chunks: chunkDataList}, nil
	})

	// Node 4: Storing
	storeNode := compose.InvokableLambda(func(ctx context.Context, input *ArticleWithEmbeddings) (IngestionOutput, error) {
		ids, err := service.milvusStore.InsertChunks(ctx, input.Chunks)
		if err != nil {
			service.db.Model(input.Article).Updates(map[string]interface{}{
				"vector_status":   "failed",
				"vector_progress": 0,
			})
			return IngestionOutput{ArticleID: input.Article.ID, Status: "failed", Error: err}, nil
		}

		// Save Chunks to DB
		// Clear old chunks first to avoid duplication on re-index
		service.db.Where("article_id = ?", input.Article.ID).Delete(&models.Chunk{})

		var dbChunks []models.Chunk
		for i, c := range input.Chunks {
			var vid int64
			if i < len(ids) {
				vid = ids[i]
			}
			dbChunks = append(dbChunks, models.Chunk{
				ArticleID:  c.ArticleID,
				Content:    c.Content,
				ChunkIndex: c.ChunkIndex,
				VectorID:   vid,
			})
		}
		if len(dbChunks) > 0 {
			service.db.Create(&dbChunks)
		}

		service.db.Model(input.Article).Updates(map[string]interface{}{
			"vector_status":   "completed",
			"vector_progress": 100,
		})
		return IngestionOutput{ArticleID: input.Article.ID, Status: "completed"}, nil
	})

	// 3. Add Nodes
	if err := g.AddLambdaNode("fetch", fetchNode); err != nil { return nil, err }
	if err := g.AddLambdaNode("chunk", chunkNode); err != nil { return nil, err }
	if err := g.AddLambdaNode("embed", embedNode); err != nil { return nil, err }
	if err := g.AddLambdaNode("store", storeNode); err != nil { return nil, err }

	// 4. Add Edges
	if err := g.AddEdge(compose.START, "fetch"); err != nil { return nil, err }
	if err := g.AddEdge("fetch", "chunk"); err != nil { return nil, err }
	if err := g.AddEdge("chunk", "embed"); err != nil { return nil, err }
	if err := g.AddEdge("embed", "store"); err != nil { return nil, err }
	if err := g.AddEdge("store", compose.END); err != nil { return nil, err }

	// 5. Compile
	r, err := g.Compile(context.Background())
	if err != nil {
		return nil, err
	}

	return &IngestionWorker{runnable: r, service: service}, nil
}

func (w *IngestionWorker) Run(ctx context.Context, articleID uint) {
	// Run in background
	go func() {
		out, err := w.runnable.Invoke(context.Background(), IngestionInput{ArticleID: articleID})
		if err != nil {
			log.Printf("Ingestion failed for article %d: %v", articleID, err)
			// Mark as failed if not handled inside
			var article models.Article
			if err := w.service.db.First(&article, articleID).Error; err == nil {
				w.service.db.Model(&article).Updates(map[string]interface{}{
					"vector_status":   "failed",
					"vector_progress": 0,
				})
			}
		} else if out.Error != nil {
			log.Printf("Ingestion failed for article %d: %v", articleID, out.Error)
			// Already handled inside storeNode but safe to double check or log
		} else {
			log.Printf("Ingestion completed for article %d", articleID)
		}
	}()
}
