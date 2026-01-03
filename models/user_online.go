package models

import (
    "time"
)

type UserOnlineStatus struct {
    UserID         uint      `json:"user_id" gorm:"primaryKey"`
    LastActiveTime time.Time `json:"last_active_time" gorm:"default:CURRENT_TIMESTAMP"`
    SessionID      string    `json:"session_id" gorm:"size:255"`
    IPAddress      string    `json:"ip_address" gorm:"size:45"`
    UserAgent      string    `json:"user_agent" gorm:"type:text"`

    // 关联用户信息
    User User `json:"user" gorm:"foreignKey:UserID"`
}

func (UserOnlineStatus) TableName() string {
    return "user_online_status"
}
