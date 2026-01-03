package models

import (
    "time"
)

type PostFavorite struct {
    ID        uint      `gorm:"primaryKey" json:"id"`
    UserID    int       `gorm:"not null" json:"user_id"`
    PostID    int       `gorm:"not null" json:"post_id"`
    CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (PostFavorite) TableName() string {
    return "post_favorites"
}
