package models

import (
    "gorm.io/gorm"
    "time"
)

type Comment struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Content   string         `json:"content" gorm:"type:text;not null"`    // 评论内容
    PostID    uint           `json:"post_id" gorm:"not null"`              // 关联的文章ID
    UserID    uint           `json:"user_id" gorm:"not null"`              // 评论用户ID
    ParentID  uint           `json:"parent_id" gorm:"default:0"`           // 父评论ID(用于回复)
    IsRecommended bool       `json:"is_recommended" gorm:"default:false"`  // 是否推荐
    RecommendRank int        `json:"recommend_rank" gorm:"default:0"`      // 推荐排序
    StatusCode int           `json:"status_code" gorm:"default:1"`         // 1:正常 2:禁用 3:待审核
    LikeCount    int         `json:"like_count" gorm:"default:0"`    // 点赞数
    DislikeCount int         `json:"dislike_count" gorm:"default:0"` // 反对数
    ReplyCount   int         `json:"reply_count" gorm:"default:0"`   // 回复数
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

    // 添加用户关联字段
    User      User           `json:"user" gorm:"foreignKey:UserID"`
}

// 表名
func (Comment) TableName() string {
    return "comments"
}
