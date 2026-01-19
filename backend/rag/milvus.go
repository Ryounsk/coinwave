package rag

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
)

type MilvusStore struct {
	cli client.Client
}

func NewMilvusStore(ctx context.Context) (*MilvusStore, error) {
	c, err := client.NewClient(ctx, client.Config{
		Address: MilvusAddress,
	})
	if err != nil {
		return nil, err
	}
	return &MilvusStore{cli: c}, nil
}

func (m *MilvusStore) InitCollection(ctx context.Context) error {
	has, err := m.cli.HasCollection(ctx, MilvusCollectionName)
	if err != nil {
		return err
	}
	if has {
		return nil
	}

	schema := &entity.Schema{
		CollectionName: MilvusCollectionName,
		Description:    "Article chunks for RAG",
		Fields: []*entity.Field{
			{
				Name:       "vector_id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     true,
			},
			{
				Name:     "user_id",
				DataType: entity.FieldTypeInt64, // Use Int64 for filter
			},
			{
				Name:     "article_id",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:     "chunk_index",
				DataType: entity.FieldTypeInt64,
			},
			{
				Name:     "embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					"dim": fmt.Sprintf("%d", MilvusDim),
				},
			},
			// Metadata fields as dynamic fields or specific fields?
			// Milvus supports scalar fields.
			{
				Name:     "content",
				DataType: entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					"max_length": "8192", // Allow large content
				},
			},
		},
	}

	err = m.cli.CreateCollection(ctx, schema, entity.DefaultShardNumber)
	if err != nil {
		return err
	}

	// Create index
	idx, err := entity.NewIndexIvfFlat(entity.L2, 1024)
	if err != nil {
		return err
	}
	err = m.cli.CreateIndex(ctx, MilvusCollectionName, "embedding", idx, false)
	if err != nil {
		return err
	}

	// Create partition key index if needed, or just load
	err = m.cli.LoadCollection(ctx, MilvusCollectionName, false)
	return err
}

func (m *MilvusStore) InsertChunks(ctx context.Context, chunks []ChunkData) ([]int64, error) {
	if len(chunks) == 0 {
		return nil, nil
	}

	// Prepare columns
	userIds := make([]int64, len(chunks))
	articleIds := make([]int64, len(chunks))
	chunkIndices := make([]int64, len(chunks))
	contents := make([]string, len(chunks))
	embeddings := make([][]float32, len(chunks))

	for i, c := range chunks {
		userIds[i] = int64(c.UserID)
		articleIds[i] = int64(c.ArticleID)
		chunkIndices[i] = int64(c.ChunkIndex)
		contents[i] = c.Content
		embeddings[i] = c.Embedding
	}

	userCol := entity.NewColumnInt64("user_id", userIds)
	articleCol := entity.NewColumnInt64("article_id", articleIds)
	chunkIdxCol := entity.NewColumnInt64("chunk_index", chunkIndices)
	contentCol := entity.NewColumnVarChar("content", contents)
	embeddingCol := entity.NewColumnFloatVector("embedding", MilvusDim, embeddings)

	// vector_id is AutoID, so we don't provide it
	result, err := m.cli.Insert(ctx, MilvusCollectionName, "", userCol, articleCol, chunkIdxCol, contentCol, embeddingCol)
	if err != nil {
		return nil, err
	}

	// Retrieve generated IDs if needed.
	// Milvus Insert returns Column of IDs if AutoID is true.
	// The SDK returns Column based on PK.
	// We need to extract IDs.
	// However, result is Column, we need to cast it.
	// Typically result.IDs is NOT available directly on Column interface in all versions, 
    // but in v2 sdk, Insert returns (Column, error).
    
    // Let's assume we don't strictly need the IDs returned here for now, 
    // or we can implement extraction if required.
    // For now, let's just return empty IDs or try to extract.
    
    // In newer SDKs:
    // idCol, ok := result.(*entity.ColumnInt64)
    // if ok { return idCol.Data(), nil }
    
    // Just return nil IDs for now unless we need to update DB with them.
    // The requirement says "return generated IDs" is not strictly required but good to have.
    // We will update DB with vector_id later if needed.
    
    // Actually, we should try to get IDs to update the Chunk table.
    if result.Name() == "vector_id" {
        // It's likely the ID column
         if idCol, ok := result.(*entity.ColumnInt64); ok {
             return idCol.Data(), nil
         }
    }

	return nil, nil
}

type SearchResult struct {
	ID         int64
	Score      float32
	Content    string
	ArticleID  int64
	ChunkIndex int64
}

func (m *MilvusStore) Search(ctx context.Context, userID uint, queryVector []float32, topK int) ([]SearchResult, error) {
	// Filter by user_id
	expr := fmt.Sprintf("user_id == %d", userID)

	sp, _ := entity.NewIndexIvfFlatSearchParam(10) // nprobe

	// We want to return content and other metadata
	outputFields := []string{"content", "article_id", "chunk_index"}

	vectors := []entity.Vector{entity.FloatVector(queryVector)}

	results, err := m.cli.Search(ctx, MilvusCollectionName, []string{}, expr, outputFields, vectors, "embedding",
		entity.L2, topK, sp)
	if err != nil {
		return nil, err
	}

	var ret []SearchResult
	for _, res := range results {
		for i := 0; i < res.ResultCount; i++ {
			// Extract fields
			// SDK SearchResult access is a bit verbose
			// We need to use GetColumn to get output fields
			
			// This part can be tricky with different SDK versions.
			// Let's try standard way.
            
            // Wait, res is SearchResult which has Fields (Column).
            // But iteration is per result set (one query vector -> one SearchResult).
            // Inside SearchResult, we have IDs and Scores.
            // Output fields are in res.Fields.
            
            id := res.IDs.(*entity.ColumnInt64).Data()[i]
            score := res.Scores[i]
            
            // Helper to get field data
            var content string
            var articleID int64
            var chunkIndex int64
            
            for _, field := range res.Fields {
                if field.Name() == "content" {
                    content = field.(*entity.ColumnVarChar).Data()[i]
                } else if field.Name() == "article_id" {
                    articleID = field.(*entity.ColumnInt64).Data()[i]
                } else if field.Name() == "chunk_index" {
                    chunkIndex = field.(*entity.ColumnInt64).Data()[i]
                }
            }
            
            ret = append(ret, SearchResult{
                ID:         id,
                Score:      score,
                Content:    content,
                ArticleID:  articleID,
                ChunkIndex: chunkIndex,
            })
		}
	}
	return ret, nil
}

type ChunkData struct {
	UserID     uint
	ArticleID  uint
	ChunkIndex int
	Content    string
	Embedding  []float32
}
