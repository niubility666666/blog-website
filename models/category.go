package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	Name          string         `json:"name" gorm:"size:40;not null"`
	Alias         string         `json:"alias" gorm:"size:40"`
	IsRecommended bool           `json:"is_recommended" gorm:"default:false"`
	RecommendRank int            `json:"recommend_rank" gorm:"default:0"`
	StatusCode    int            `json:"status_code" gorm:"default:1"` // 1:正常 2:禁用 3:待审核
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// 表名
func (Category) TableName() string {
	return "categories"
}
