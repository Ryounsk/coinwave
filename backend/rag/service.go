package rag

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"coin-wave/models"

	"gorm.io/gorm"
)

type RagService struct {
	db          *gorm.DB
	volcClient  *VolcClient
	milvusStore *MilvusStore
}

func NewRagService(db *gorm.DB, volc *VolcClient, milvus *MilvusStore) *RagService {
	return &RagService{
		db:          db,
		volcClient:  volc,
		milvusStore: milvus,
	}
}

// ProcessArticle chunks the article, embeds it, and stores it in Milvus.
func (s *RagService) ProcessArticle(ctx context.Context, articleID uint) error {
	var article models.Article
	if err := s.db.First(&article, articleID).Error; err != nil {
		return err
	}

	// Update status to processing
	s.db.Model(&article).Update("vector_status", "processing")

	// 1. Chunking
	chunks := s.chunkText(article.Content, ChunkSize, ChunkOverlap)

	// 2. Prepare texts for embedding
	// Format: Title + Tags + Content
	var textsToEmbed []string
	var chunkDataList []ChunkData

	for i, content := range chunks {
		text := fmt.Sprintf("%s %s\n%s", article.Title, article.Tags, content)
		textsToEmbed = append(textsToEmbed, text)
		chunkDataList = append(chunkDataList, ChunkData{
			UserID:     article.AuthorID,
			ArticleID:  article.ID,
			ChunkIndex: i,
			Content:    content,
			// Embedding will be filled later
		})
	}

	// 3. Batch Embedding
	// Max batch size for Doubao API is usually limited (e.g. 20 or 50).
	// We should batch requests.
	batchSize := 10
	for i := 0; i < len(textsToEmbed); i += batchSize {
		end := i + batchSize
		if end > len(textsToEmbed) {
			end = len(textsToEmbed)
		}

		batchTexts := textsToEmbed[i:end]
		log.Printf("Embedding batch %d-%d for article %d", i, end, articleID) // Debug Log
		embeddings, err := s.volcClient.GetEmbeddings(batchTexts)
		if err != nil {
			log.Printf("Embedding failed: %v", err)
			s.db.Model(&article).Update("vector_status", "failed")
			return err
		}

		// Fill embeddings back to chunkDataList
		for j, emb := range embeddings {
			chunkDataList[i+j].Embedding = emb
		}
	}

	// 4. Insert to Milvus
	ids, err := s.milvusStore.InsertChunks(ctx, chunkDataList)
	if err != nil {
		s.db.Model(&article).Update("vector_status", "failed")
		return err
	}

	// 5. Save Chunks to DB (Optional, but requested)
	var dbChunks []models.Chunk
	for i, c := range chunkDataList {
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
	// Use batch insert
	if len(dbChunks) > 0 {
		if err := s.db.Create(&dbChunks).Error; err != nil {
			log.Printf("Failed to save chunks to DB: %v", err)
			// Don't fail the whole process if DB save fails, but good to know
		}
	}

	s.db.Model(&article).Update("vector_status", "completed")
	return nil
}

func (s *RagService) chunkText(text string, chunkSize, overlap int) []string {
	// Simple character-based chunking for now.
	// Production systems should use token-based chunking.
	// Assuming 1 char ~= 1 token for simplicity or use a library if available.
	// Since we don't have a tokenizer lib in dependencies, we'll use simple rune counting.
	
	runes := []rune(text)
	var chunks []string
	
	if len(runes) == 0 {
		return chunks
	}

	for i := 0; i < len(runes); i += (chunkSize - overlap) {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
		if end == len(runes) {
			break
		}
	}
	return chunks
}

func (s *RagService) Query(ctx context.Context, userID uint, question string) (string, []string, map[string]float64, error) {
	timings := make(map[string]float64)
	start := time.Now()

	// 1. Embed question
	embedStart := time.Now()
	embeddings, err := s.volcClient.GetEmbeddings([]string{question})
	if err != nil {
		return "", nil, nil, err
	}
	timings["embedding"] = time.Since(embedStart).Seconds()

	if len(embeddings) == 0 {
		return "", nil, nil, fmt.Errorf("failed to embed question")
	}

	// 2. Search Milvus
	searchStart := time.Now()
	results, err := s.milvusStore.Search(ctx, userID, embeddings[0], 5) // Top 5
	if err != nil {
		return "", nil, nil, err
	}
	timings["search"] = time.Since(searchStart).Seconds()

	// 3. Construct Prompt
	var retrievedChunks []string
	for _, res := range results {
		retrievedChunks = append(retrievedChunks, res.Content)
	}

	contextText := strings.Join(retrievedChunks, "\n\n")
	
	systemPrompt := fmt.Sprintf(`你是用户的私人知识助理。以下内容来自用户自己的文章：
%s

用户问题：
%s

请严格根据文章内容回答，不允许凭空生成。`, contextText, question)

	// 4. Call LLM
	llmStart := time.Now()
	messages := []Message{
		{Role: "user", Content: systemPrompt},
	}
	
	answer, err := s.volcClient.Chat(messages)
	if err != nil {
		return "", nil, nil, err
	}
	timings["llm"] = time.Since(llmStart).Seconds()

	timings["total_internal"] = time.Since(start).Seconds()

	return answer, retrievedChunks, timings, nil
}
