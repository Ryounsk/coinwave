package rag

import (
	"os"
)

var (
	VolcEmbeddingEndpoint = "https://ark.cn-beijing.volces.com/api/v3/embeddings/multimodal"
	VolcResponseEndpoint  = "https://ark.cn-beijing.volces.com/api/v3/responses"
	VolcModelEmbedding    = "ep-20260119190844-zzrq7"
	VolcModelLLM          = "ep-20260119190028-tmjkb"
	VolcAuthToken         = "97443eb3-4d60-4738-bf9d-e7bf364566d2" // Default fallback

	MilvusAddress         = "localhost:19530"
	MilvusCollectionName  = "coinwave_articles"
	MilvusDim             = 2048
)

// Chunking config
const (
	ChunkSize    = 800
	ChunkOverlap = 100
)

func init() {
	if v := os.Getenv("VOLC_AUTH_TOKEN"); v != "" {
		VolcAuthToken = v
	}
	if v := os.Getenv("MILVUS_ADDRESS"); v != "" {
		MilvusAddress = v
	}
	// Allow overriding other configs if needed
	if v := os.Getenv("VOLC_MODEL_EMBEDDING"); v != "" {
		VolcModelEmbedding = v
	}
	if v := os.Getenv("VOLC_MODEL_LLM"); v != "" {
		VolcModelLLM = v
	}
}
