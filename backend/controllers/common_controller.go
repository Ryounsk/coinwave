package controllers

import (
	"coin-wave/database"
	"coin-wave/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetExchangeRate(c *gin.Context) {
	// Mock data
	rates := map[string]float64{
		"USD": 1.0,
		"CNY": 7.25,
		"EUR": 0.92,
		"BTC": 65000.0,
		"ETH": 3500.0,
	}
	c.JSON(http.StatusOK, gin.H{"data": rates})
}

func GetRankings(c *gin.Context) {
	period := c.Query("period") // daily, monthly, yearly
	if period == "" {
		period = "daily"
	}

	var redisKey string
	now := time.Now()

	switch period {
	case "daily":
		redisKey = "rankings:daily:" + now.Format("2006-01-02")
	case "monthly":
		redisKey = "rankings:monthly:" + now.Format("2006-01")
	case "yearly":
		redisKey = "rankings:yearly:" + now.Format("2006")
	default:
		redisKey = "rankings:daily:" + now.Format("2006-01-02")
	}

	// 1. Try to get Top 10 from Redis ZSET
	vals, err := database.RDB.ZRevRangeWithScores(database.Ctx, redisKey, 0, 9).Result()

	var articles []models.Article

	if err == nil && len(vals) > 0 {
		// Redis Hit
		ids := make([]string, len(vals))
		for i, v := range vals {
			ids[i] = v.Member.(string)
		}

		var dbArticles []models.Article
		database.DB.Preload("Author").Where("id IN ?", ids).Find(&dbArticles)

		// Re-sort in memory to match Redis order
		articleMap := make(map[string]models.Article)
		for _, a := range dbArticles {
			articleMap[strconv.Itoa(int(a.ID))] = a
		}

		for _, idStr := range ids {
			if a, ok := articleMap[idStr]; ok {
				articles = append(articles, a)
			}
		}
	} else {
		// Redis Miss: Fallback to DB (Slow Path) & Populate Redis
		var startTime time.Time
		switch period {
		case "daily":
			startTime = now.AddDate(0, 0, -1)
		case "monthly":
			startTime = now.AddDate(0, -1, 0)
		case "yearly":
			startTime = now.AddDate(-1, 0, 0)
		default:
			startTime = now.AddDate(0, 0, -1)
		}

		type Result struct {
			ArticleID uint
			Count     int
		}
		var results []Result

		err := database.DB.Table("bookmarks").
			Select("article_id, count(*) as count").
			Where("created_at >= ?", startTime).
			Group("article_id").
			Order("count desc").
			Limit(10).
			Scan(&results).Error

		if err == nil && len(results) > 0 {
			// Populate Redis
			pipe := database.RDB.Pipeline()
			for _, r := range results {
				pipe.ZAdd(database.Ctx, redisKey, redis.Z{
					Score:  float64(r.Count),
					Member: r.ArticleID,
				})
			}
			pipe.Expire(database.Ctx, redisKey, 24*time.Hour)
			pipe.Exec(database.Ctx)

			// Fetch Articles
			ids := make([]uint, len(results))
			for i, r := range results {
				ids[i] = r.ArticleID
			}
			
			// We need to preserve order here too, or just accept DB order for the first fetch.
			// Let's rely on DB Find, and since we just built `results` in order, we can map it back.
			var dbArticles []models.Article
			database.DB.Preload("Author").Where("id IN ?", ids).Find(&dbArticles)
			
			articleMap := make(map[uint]models.Article)
			for _, a := range dbArticles {
				articleMap[a.ID] = a
			}
			
			for _, r := range results {
				if a, ok := articleMap[r.ArticleID]; ok {
					articles = append(articles, a)
				}
			}
		}
	}

	if articles == nil {
		articles = []models.Article{}
	}

	c.JSON(http.StatusOK, gin.H{"data": articles})
}
