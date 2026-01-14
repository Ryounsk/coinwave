package controllers

import (
	"coin-wave/database"
	"coin-wave/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DepositInput struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func Deposit(c *gin.Context) {
	var input DepositInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Update User Balance
		if err := tx.Model(&models.User{}).Where("id = ?", userID).UpdateColumn("balance", gorm.Expr("balance + ?", input.Amount)).Error; err != nil {
			return err
		}

		// Create Wallet Log
		log := models.WalletLog{
			UserID:      userID.(uint),
			Amount:      input.Amount,
			Type:        "deposit",
			Description: "User deposit",
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Deposit failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deposit successful"})
}

func GetBalance(c *gin.Context) {
	userID, _ := c.Get("userID")
	var user models.User
	if err := database.DB.Select("balance").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": user.Balance})
}

func PurchaseArticle(c *gin.Context) {
	articleID := c.Param("id")
	userID, _ := c.Get("userID")
	uid := userID.(uint)

	var article models.Article
	if err := database.DB.First(&article, articleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Article not found"})
		return
	}

	if !article.IsPaid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Article is free"})
		return
	}

	if article.AuthorID == uid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot purchase your own article"})
		return
	}

	// Check if already purchased
	var count int64
	database.DB.Model(&models.Purchase{}).Where("user_id = ? AND article_id = ?", uid, article.ID).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "Already purchased"})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.First(&user, uid).Error; err != nil {
			return err
		}

		if user.Balance < article.Price {
			return  gorm.ErrInvalidData // Not enough balance
		}

		// Deduct from buyer
		if err := tx.Model(&user).UpdateColumn("balance", gorm.Expr("balance - ?", article.Price)).Error; err != nil {
			return err
		}

		// Add to author
		if err := tx.Model(&models.User{}).Where("id = ?", article.AuthorID).UpdateColumn("balance", gorm.Expr("balance + ?", article.Price)).Error; err != nil {
			return err
		}

		// Create Purchase Record
		purchase := models.Purchase{
			UserID:    uid,
			ArticleID: article.ID,
		}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}

		// Log for Buyer
		buyerLog := models.WalletLog{
			UserID:      uid,
			Amount:      -article.Price,
			Type:        "purchase",
			Description: "Purchased article: " + article.Title,
		}
		if err := tx.Create(&buyerLog).Error; err != nil {
			return err
		}

		// Log for Seller
		sellerLog := models.WalletLog{
			UserID:      article.AuthorID,
			Amount:      article.Price,
			Type:        "sale",
			Description: "Sold article: " + article.Title,
		}
		if err := tx.Create(&sellerLog).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrInvalidData {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Purchase failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Purchase successful"})
}
