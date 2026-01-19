package main

import (
	"context"
	"log"
	"time"

	"coin-wave/controllers"
	"coin-wave/database"
	"coin-wave/rag"
	"coin-wave/routes"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	// Initialize RAG Components
	ctx := context.Background()
	milvusStore, err := rag.NewMilvusStore(ctx)
	if err != nil {
		log.Printf("Warning: Failed to connect to Milvus: %v", err)
	} else {
		// Initialize Collection
		if err := milvusStore.InitCollection(ctx); err != nil {
			log.Printf("Warning: Failed to init Milvus collection: %v", err)
		}
	}

	volcClient := rag.NewVolcClient()

	var ragService *rag.RagService
	var ingestionWorker *rag.IngestionWorker

	if milvusStore != nil {
		ragService = rag.NewRagService(database.DB, volcClient, milvusStore)
		ingestionWorker, err = rag.NewIngestionWorker(ragService)
		if err != nil {
			log.Printf("Warning: Failed to create ingestion worker: %v", err)
		}

		controllers.InitRag(ingestionWorker, ragService)
		log.Println("RAG System Initialized Successfully")
	}

	r := gin.Default()

	// CORS Setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://localhost"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Setup Routes
	routes.SetupRoutes(r)

	// Backward Compatibility Redirects (Optional, or just force frontend update)
	// For now, let's just use v1.

	r.Run(":8080")
}
