package routes

import (
	"coin-wave/controllers"
	"coin-wave/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// API v1 Grouping
	v1 := r.Group("/api/v1")
	// Compatibility Routes for old frontend clients
	compat := r.Group("/api")
	{
		// Auth Compatibility
		compatAuth := compat.Group("/auth")
		{
			compatAuth.POST("/register", controllers.Register)
			compatAuth.POST("/login", controllers.Login)
		}

		// Articles Compatibility
		compatArticles := compat.Group("/articles")
		{
			compatArticles.GET("", middleware.OptionalAuthMiddleware(), controllers.GetArticles)
			compatArticles.GET("/:id", middleware.OptionalAuthMiddleware(), controllers.GetArticle)
			compatArticles.POST("", middleware.AuthMiddleware(), controllers.CreateArticle)
			compatArticles.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteArticle)
			compatArticles.POST("/:id/bookmark", middleware.AuthMiddleware(), controllers.BookmarkArticle)
			compatArticles.POST("/:id/purchase", middleware.AuthMiddleware(), controllers.PurchaseArticle)
			compatArticles.POST("/:id/reindex", middleware.AuthMiddleware(), controllers.ReIndexArticle)
		}

		// User Compatibility
		compatUser := compat.Group("/user")
		compatUser.Use(middleware.AuthMiddleware())
		{
			compatUser.GET("/articles", controllers.GetUserArticles)
			compatUser.GET("/bookmarks", controllers.GetUserBookmarks)
		}

		// Wallet Compatibility
		compatWallet := compat.Group("/wallet")
		compatWallet.Use(middleware.AuthMiddleware())
		{
			compatWallet.POST("/deposit", controllers.Deposit)
			compatWallet.GET("/balance", controllers.GetBalance)
		}

		// RAG Compatibility
		compatRag := compat.Group("/rag")
		compatRag.Use(middleware.AuthMiddleware())
		{
			compatRag.POST("/query", controllers.RagQuery)
		}

		compat.GET("/rankings", middleware.OptionalAuthMiddleware(), controllers.GetRankings)
		compat.GET("/rates", middleware.OptionalAuthMiddleware(), controllers.GetExchangeRate)
	}

	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Articles Routes (Public & Protected mixed, handling inside)
		articles := v1.Group("/articles")
		{
			// Public
			articles.GET("", middleware.OptionalAuthMiddleware(), controllers.GetArticles)
			articles.GET("/:id", middleware.OptionalAuthMiddleware(), controllers.GetArticle)
			
			// Protected
			articles.POST("", middleware.AuthMiddleware(), controllers.CreateArticle)
			articles.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteArticle)
			articles.POST("/:id/bookmark", middleware.AuthMiddleware(), controllers.BookmarkArticle)
			articles.POST("/:id/purchase", middleware.AuthMiddleware(), controllers.PurchaseArticle)
			articles.POST("/:id/reindex", middleware.AuthMiddleware(), controllers.ReIndexArticle)
		}

		// User Routes
		user := v1.Group("/user")
		user.Use(middleware.AuthMiddleware())
		{
			user.GET("/articles", controllers.GetUserArticles)
			user.GET("/bookmarks", controllers.GetUserBookmarks)
		}

		// Wallet Routes
		wallet := v1.Group("/wallet")
		wallet.Use(middleware.AuthMiddleware())
		{
			wallet.POST("/deposit", controllers.Deposit)
			wallet.GET("/balance", controllers.GetBalance)
		}
		
		// RAG Routes
		ragGroup := v1.Group("/rag")
		ragGroup.Use(middleware.AuthMiddleware())
		{
			ragGroup.POST("/query", controllers.RagQuery)
		}

		// Misc
		v1.GET("/rankings", middleware.OptionalAuthMiddleware(), controllers.GetRankings)
		v1.GET("/rates", middleware.OptionalAuthMiddleware(), controllers.GetExchangeRate)
	}
}
