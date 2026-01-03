package handlers

import (
	"fmt"
	"gin-doniai/database"
	"gin-doniai/models"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// UpdateUserOnlineStatus 更新用户在线状态
func UpdateUserOnlineStatus(c *gin.Context) {
    // 从上下文获取用户信息
    userObj, exists := c.Get("user")
    if !exists || userObj == nil {
        return // 未登录用户不处理
    }

    // 安全地进行类型断言
    user, ok := userObj.(*models.User)
    // 增强检查确保 user 对象不为 nil
    if !ok || user == nil || user.ID == 0 {
        return // 用户信息无效时不处理
    }

    // 获取session
    session := sessions.Default(c)

    // 正确获取session ID的方式
    var sessionID string
    if sessionIDInterface := session.Get("session_id"); sessionIDInterface != nil {
        sessionID = sessionIDInterface.(string)
    } else {
        // 如果没有存储的session_id，则使用session的ID
        sessionID = fmt.Sprintf("%v", session.ID())
        // 可选：将session ID存储到session中供后续使用
        session.Set("session_id", sessionID)
        session.Save()
    }

    // 获取客户端IP和User-Agent
    clientIP := c.ClientIP()
    userAgent := c.GetHeader("User-Agent")

    // 更新或创建在线状态记录
    onlineStatus := models.UserOnlineStatus{
        UserID:         user.ID,
        LastActiveTime: time.Now(),
        SessionID:      sessionID,
        IPAddress:      clientIP,
        UserAgent:      userAgent,
    }

    // 使用Upsert操作更新在线状态
    database.DB.Where("user_id = ?", user.ID).Assign(onlineStatus).FirstOrCreate(&onlineStatus)
}



// GetOnlineUserCount 获取在线用户数
func GetOnlineUserCount(c *gin.Context) {
	var count int64

	// 统计最近30分钟内活跃的用户数
	cutoffTime := time.Now().Add(-30 * time.Minute)
	database.DB.Model(&models.UserOnlineStatus{}).
		Where("last_active_time > ?", cutoffTime).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"online_count": count,
	})
}

// CleanupExpiredOnlineStatus 清理过期的在线状态记录
func CleanupExpiredOnlineStatus() {
	cutoffTime := time.Now().Add(-30 * time.Minute)
	database.DB.Where("last_active_time < ?", cutoffTime).
		Delete(&models.UserOnlineStatus{})
}

// UpdateUserOnlineStatusWithInfo 更新用户在线状态（用于消息队列处理）
func UpdateUserOnlineStatusWithInfo(userID uint, clientIP string, userAgent string) {
    // 获取session（注意：在批量处理中可能需要不同的session处理方式）
    // 这里我们简化处理，只使用必要的信息

    // 获取当前时间
    currentTime := time.Now()

    // 创建在线状态记录
    onlineStatus := models.UserOnlineStatus{
        UserID:         userID,
        LastActiveTime: currentTime,
        IPAddress:      clientIP,
        SessionID:      fmt.Sprintf("batch-%d-%s", userID, currentTime.Format("20060102150405")),
        UserAgent:      userAgent,
    }

    // 使用Upsert操作更新在线状态
    database.DB.Where("user_id = ?", userID).Assign(onlineStatus).FirstOrCreate(&onlineStatus)
}

