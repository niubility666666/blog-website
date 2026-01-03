// models/password_reset.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type PasswordReset struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Email     string         `gorm:"not null" json:"email"`
    Token     string         `gorm:"not null;uniqueIndex" json:"token"`
    ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
    Used      bool           `gorm:"default:false" json:"used"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// 表名
func (PasswordReset) TableName() string {
    return "password_reset"
}