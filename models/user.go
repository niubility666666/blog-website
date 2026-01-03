package models

import (
    "gorm.io/gorm"
    "time"
)

type User struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Name      string         `json:"name" gorm:"size:100;not null"`
    Email     string         `json:"email" gorm:"size:100;uniqueIndex;not null"`
    Password  string         `json:"password" gorm:"size:255;not null"`
    Avatar    string         `json:"avatar" gorm:"size:255;not null"`
    Age       int            `json:"age" gorm:"default:0"`
    Level     int            `json:"level" gorm:"default:1"`
    AgreeTerms bool          `json:"agree_terms" gorm:"default:false"` // 修改为布尔类型
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`

    // 添加缺失的字段
    Motto         string    `json:"motto"`          // 个人格言
    Github        string    `json:"github"`         // GitHub账号
    GoogleAccount string    `json:"google_account"` // Google账户
}

// 表名
func (User) TableName() string {
    return "users"
}