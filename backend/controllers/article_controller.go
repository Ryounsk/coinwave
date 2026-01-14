package controllers

import (
	"coin-wave/database"
	"coin-wave/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"math"
	"sort"
	"strings"
	"time"
)

type CreateArticleInput struct {
	Title   string  `json:"title" binding:"required"`
	Content string  `json:"content" binding:"required"`
	Tags    string  `json:"tags" binding:"required"`
	IsPaid  bool    `json:"is_paid"`
	Price   float64 `json:"price"`
}

func CreateArticle(c *gin.Context) {
	var input CreateArticleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	user := models.User{}
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	article := models.Article{
		Title:    input.Title,
		Content:  input.Content,
		AuthorID: user.ID,
		Tags:     input.Tags,
		IsPaid:   input.IsPaid,
		Price:    input.Price,
	}

	if err := database.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create article"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": article})
}

func GetArticles(c *gin.Context) {
	// Common base query
	db := database.DB.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username") // Omit password
	})

	// Search Logic (Recall -> Rank -> Strategy)
	if search := c.Query("search"); search != "" {
		var candidates []models.Article
		keywords := strings.Fields(strings.ToLower(search))

		if len(keywords) == 0 {
			c.JSON(http.StatusOK, gin.H{"data": []models.Article{}})
			return
		}

		// 1. Recall: Broad match using SQL
		recallQuery := db

		// Apply Filters (Free/Paid) during Recall
		if typeParam := c.Query("type"); typeParam == "free" {
			recallQuery = recallQuery.Where("is_paid = ?", false)
		} else if typeParam == "paid" {
			recallQuery = recallQuery.Where("is_paid = ?", true)
		}

		// Build OR query for all keywords
		conditions := make([]string, 0)
		args := make([]interface{}, 0)
		for _, kw := range keywords {
			pattern := "%" + kw + "%"
			conditions = append(conditions, "title LIKE ? OR content LIKE ? OR tags LIKE ?")
			args = append(args, pattern, pattern, pattern)
		}

		if len(conditions) > 0 {
			recallQuery = recallQuery.Where(strings.Join(conditions, " OR "), args...)
		}

		if err := recallQuery.Find(&candidates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
			return
		}

		// 2. Rank: In-memory scoring
		type ScoredArticle struct {
			Article models.Article
			Score   float64
		}

		scored := make([]ScoredArticle, 0, len(candidates))

		for _, art := range candidates {
			score := 0.0
			titleLower := strings.ToLower(art.Title)
			contentLower := strings.ToLower(art.Content)
			tagsLower := strings.ToLower(art.Tags)

			for _, kw := range keywords {
				// Title Weight: 10
				if strings.Contains(titleLower, kw) {
					score += 10
					if titleLower == kw {
						score += 5
					} // Exact match bonus
				}
				// Tags Weight: 8
				if strings.Contains(tagsLower, kw) {
					score += 8
				}
				// Content Weight: 1 per occurrence (capped)
				count := strings.Count(contentLower, kw)
				score += math.Min(float64(count), 10.0)
			}

			// Popularity Weight
			score += float64(art.BookmarkCount) * 2.0
			score += float64(art.ViewCount) * 0.1

			scored = append(scored, ScoredArticle{art, score})
		}

		// Sort by Score DESC
		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})

		// Extract results
		results := make([]models.Article, len(scored))
		for i, s := range scored {
			results[i] = s.Article
		}

		c.JSON(http.StatusOK, gin.H{"data": results})
		return
	}

	// Default Logic (No Search)
	var articles []models.Article
	query := db

	// Filter by free/paid
	if typeParam := c.Query("type"); typeParam == "free" {
		query = query.Where("is_paid = ?", false)
	} else if typeParam == "paid" {
		query = query.Where("is_paid = ?", true)
	}

	// Sort
	if sort := c.Query("sort"); sort == "rank" {
		query = query.Order("bookmark_count desc")
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

func updateRankings(articleID int, scoreDelta float64) {
	now := time.Now()
	keys := []string{
		"rankings:daily:" + now.Format("2006-01-02"),
		"rankings:monthly:" + now.Format("2006-01"),
		"rankings:yearly:" + now.Format("2006"),
	}

	pipe := database.RDB.Pipeline()
	for _, key := range keys {
		pipe.ZIncrBy(database.Ctx, key, scoreDelta, strconv.Itoa(articleID))
		// Refresh TTL to ensure active keys don't expire immediately if we set them before
		// We can set a long TTL safely (e.g. 30 days for daily, 365 for others)
		// Or just rely on the populate logic setting it.
		// Let's set it to be safe.
		if len(key) > 15 { // crude check for daily/monthly vs yearly
			pipe.Expire(database.Ctx, key, 7*24*time.Hour) 
		}
	}
	pipe.Exec(database.Ctx)
}

func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article
	if err := database.DB.Preload("Author", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, username")
	}).First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	// Increment View Count
	database.DB.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1))

	// Check access permissions
	userID, exists := c.Get("userID")
	hasAccess := !article.IsPaid
	
	if exists {
		uid := userID.(uint)
		if uid == article.AuthorID {
			hasAccess = true
		} else if article.IsPaid {
			var purchase models.Purchase
			if err := database.DB.Where("user_id = ? AND article_id = ?", uid, article.ID).First(&purchase).Error; err == nil {
				hasAccess = true
			}
		}
	}

	// If not paid and no access, hide content
	if !hasAccess {
		article.Content = "This content is paid. Please purchase to view."
	}

	c.JSON(http.StatusOK, gin.H{"data": article, "has_access": hasAccess})
}

func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("userID")

	var article models.Article
	if err := database.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if article.AuthorID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this article"})
		return
	}

	database.DB.Delete(&article)
	c.JSON(http.StatusOK, gin.H{"message": "Article deleted"})
}

func BookmarkArticle(c *gin.Context) {
	articleID, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("userID")

	var bookmark models.Bookmark
	err := database.DB.Where("user_id = ? AND article_id = ?", userID, articleID).First(&bookmark).Error

	if err == nil {
		// Already bookmarked, remove it
		database.DB.Delete(&bookmark)
		database.DB.Model(&models.Article{Model: gorm.Model{ID: uint(articleID)}}).UpdateColumn("bookmark_count", gorm.Expr("bookmark_count - ?", 1))
		
		// Update Redis Rankings (Decrement)
		go updateRankings(articleID, -1)

		c.JSON(http.StatusOK, gin.H{"message": "Bookmark removed", "bookmarked": false})
	} else {
		// Not bookmarked, add it
		newBookmark := models.Bookmark{
			UserID:    userID.(uint),
			ArticleID: uint(articleID),
		}
		database.DB.Create(&newBookmark)
		database.DB.Model(&models.Article{Model: gorm.Model{ID: uint(articleID)}}).UpdateColumn("bookmark_count", gorm.Expr("bookmark_count + ?", 1))
		
		// Update Redis Rankings (Increment)
		go updateRankings(articleID, 1)

		c.JSON(http.StatusOK, gin.H{"message": "Article bookmarked", "bookmarked": true})
	}
}

func GetUserArticles(c *gin.Context) {
	userID, _ := c.Get("userID")

	var articles []models.Article
	if err := database.DB.Where("author_id = ?", userID).Order("created_at desc").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user articles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}

func GetUserBookmarks(c *gin.Context) {
	userID, _ := c.Get("userID")

	var bookmarks []models.Bookmark
	if err := database.DB.Preload("Article.Author").Where("user_id = ?", userID).Order("created_at desc").Find(&bookmarks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bookmarks"})
		return
	}

	// Extract articles from bookmarks
	articles := make([]models.Article, len(bookmarks))
	for i, b := range bookmarks {
		articles[i] = b.Article
		// Ensure Author info is minimal if needed, but Preload("Article.Author") fetches full user. 
		// We might want to sanitize it (remove password) but GORM struct usually handles JSON tag "-" for password.
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}
