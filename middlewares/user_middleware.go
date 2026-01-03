package middlewares

import (
	"fmt"
	"strings"
	"gin-doniai/database"
	"gin-doniai/models"
	"gin-doniai/workers"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)


// UserAndOnlineStatusMiddleware 合并的用户信息和在线状态中间件
func UserAndOnlineStatusMiddleware(onlineStatusChan chan workers.OnlineStatusUpdate) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")

		var user *models.User
		if userID != nil {
			// 确保 userID 是有效的数字类型
			if userIDVal, ok := userID.(uint); ok && userIDVal > 0 {
				var currentUser models.User
				if err := database.DB.First(&currentUser, userIDVal).Error; err == nil {
					user = &currentUser
				}
			} else if userIDVal, ok := userID.(int); ok && userIDVal > 0 {
				var currentUser models.User
				if err := database.DB.First(&currentUser, uint(userIDVal)).Error; err == nil {
					user = &currentUser
				}
			}
		} else {
			// 当没有user_id时，不进行重定向以避免循环重定向
			// 只是简单地设置user为nil
			fmt.Println("Middleware - Session中没有user_id")
		}

		// 设置用户信息到上下文
		if user != nil {
            c.Set("user", user)
        }

		// 处理在线状态更新（只对非静态资源请求处理）
		if user != nil && !strings.HasPrefix(c.Request.URL.Path, "/static") {
			// 发送在线状态更新消息到队列
			select {
			case onlineStatusChan <- workers.OnlineStatusUpdate{
				UserID:    user.ID,
				IP:        c.ClientIP(),
				UserAgent: c.GetHeader("User-Agent"),
			}:
			default:
				// 队列满时丢弃，避免阻塞
				fmt.Println("在线状态队列已满，丢弃更新")
			}
		}

		c.Next()
	}
}

