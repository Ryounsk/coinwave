package controllers

import (
	"coin-wave/database"
	"coin-wave/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type RagQueryInput struct {
	Question string `json:"question" binding:"required"`
}

func RagQuery(c *gin.Context) {
	var input RagQueryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	uid := userID.(uint)

	if RagService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "RAG Service not initialized"})
		return
	}

	start := time.Now()
	answer, sources, timings, err := RagService.Query(c.Request.Context(), uid, input.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "RAG Query failed: " + err.Error()})
		return
	}

	// Add total handler time
	if timings == nil {
		timings = make(map[string]float64)
	}
	timings["total_handler"] = time.Since(start).Seconds()

	c.JSON(http.StatusOK, gin.H{
		"answer":  answer,
		"sources": sources,
		"timings": timings,
	})
}

// ReIndexArticle triggers re-vectorization for an existing article
func ReIndexArticle(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if article.AuthorID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Reset progress
	article.VectorProgress = 0
	database.DB.Save(&article)

	if RagWorker != nil {
		RagWorker.Run(c.Request.Context(), article.ID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Re-indexing triggered"})
}
