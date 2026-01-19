package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null;size:191" json:"username"`
	Password string `json:"-"` // Don't return password in JSON
	Balance  float64 `gorm:"default:0" json:"balance"`
	IsAdmin  bool    `gorm:"default:false" json:"is_admin"`
}

type Article struct {
	gorm.Model
	Title          string  `json:"title"`
	Content        string  `gorm:"type:text" json:"content"`
	AuthorID       uint    `json:"author_id"`
	Author         User    `json:"author"`
	Tags           string  `json:"tags"` // Comma separated tags for simplicity
	IsPaid         bool    `gorm:"default:false" json:"is_paid"`
	Price          float64 `gorm:"default:0" json:"price"`
	ViewCount      int     `gorm:"default:0" json:"view_count"`
	BookmarkCount  int     `gorm:"default:0" json:"bookmark_count"`
	VectorStatus   string  `gorm:"default:'pending'" json:"vector_status"` // pending, processing, completed, failed
	VectorProgress int     `gorm:"default:0" json:"vector_progress"`       // 0-100
}

type Chunk struct {
	gorm.Model
	ArticleID uint   `gorm:"index" json:"article_id"`
	Content   string `gorm:"type:text" json:"content"`
	ChunkIndex int    `json:"chunk_index"`
	VectorID   int64  `json:"vector_id"` // Milvus ID
}

type Bookmark struct {
	gorm.Model
	UserID    uint    `gorm:"uniqueIndex:idx_user_article" json:"user_id"`
	ArticleID uint    `gorm:"uniqueIndex:idx_user_article" json:"article_id"`
	Article   Article `json:"article"`
}

type WalletLog struct {
	gorm.Model
	UserID      uint    `json:"user_id"`
	Amount      float64 `json:"amount"` // Positive for deposit, negative for spend
	Type        string  `json:"type"`   // "deposit", "purchase"
	Description string  `json:"description"`
}

// Purchase record to track if a user bought an article
type Purchase struct {
	gorm.Model
	UserID    uint `gorm:"uniqueIndex:idx_user_purchase" json:"user_id"`
	ArticleID uint `gorm:"uniqueIndex:idx_user_purchase" json:"article_id"`
}
