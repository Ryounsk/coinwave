package main

import (
	"coin-wave/controllers"
	"coin-wave/database"
	"coin-wave/middleware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	r := gin.Default()

	// CORS Setup
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000", "http://localhost"}, // Added localhost:3000/80 for docker
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Routes
	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Public Routes
		public := api.Group("/")
		public.Use(middleware.OptionalAuthMiddleware())
		{
			public.GET("/articles", controllers.GetArticles)
			public.GET("/articles/:id", controllers.GetArticle)
			public.GET("/rankings", controllers.GetRankings)
			public.GET("/rates", controllers.GetExchangeRate)
		}

		// Protected Routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			protected.POST("/articles", controllers.CreateArticle)
			protected.DELETE("/articles/:id", controllers.DeleteArticle)
			protected.POST("/articles/:id/bookmark", controllers.BookmarkArticle)
			protected.POST("/articles/:id/purchase", controllers.PurchaseArticle)
			
			protected.POST("/wallet/deposit", controllers.Deposit)
			protected.GET("/wallet/balance", controllers.GetBalance)

			protected.GET("/user/articles", controllers.GetUserArticles)
			protected.GET("/user/bookmarks", controllers.GetUserBookmarks)
		}
	}

	r.Run(":8080")
}
