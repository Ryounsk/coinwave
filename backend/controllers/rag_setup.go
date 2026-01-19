package controllers

import (
	"coin-wave/rag"
)

var RagWorker *rag.IngestionWorker
var RagService *rag.RagService

func InitRag(worker *rag.IngestionWorker, service *rag.RagService) {
	RagWorker = worker
	RagService = service
}
