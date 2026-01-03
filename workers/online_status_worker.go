package workers

import (
	"time"
	"gin-doniai/handlers"
)

type OnlineStatusUpdate struct {
	UserID    uint
	IP        string
	UserAgent string
}

func HandleOnlineStatusUpdates(onlineStatusChan <-chan OnlineStatusUpdate) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var updates []OnlineStatusUpdate
	batchSize := 50 // 批量处理大小

	for {
		select {
		case update := <-onlineStatusChan:
			updates = append(updates, update)

			// 达到批次大小时立即处理
			if len(updates) >= batchSize {
				processBatchOnlineStatus(updates)
				updates = updates[:0] // 清空切片
			}

		case <-ticker.C:
			// 定时处理剩余的更新
			if len(updates) > 0 {
				processBatchOnlineStatus(updates)
				updates = updates[:0]
			}
		}
	}
}

func processBatchOnlineStatus(updates []OnlineStatusUpdate) {
	for _, update := range updates {
		// 创建模拟的 gin.Context 用于处理
		// 或者创建新的批量处理方法
		handlers.UpdateUserOnlineStatusWithInfo(update.UserID, update.IP, update.UserAgent)
	}
}
