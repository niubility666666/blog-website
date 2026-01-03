package workers

import (
    "fmt"
    "time"
    "gin-doniai/database"
    "gin-doniai/models"
    "gorm.io/gorm"
)

type ViewEvent struct {
    PostID    uint
    UserID    *uint
    IP        string
    UserAgent string
    Timestamp time.Time
}

func HandleViewNumUpdates(viewChan chan ViewEvent) {
    // 使用map记录用户/IP对文章的访问时间
    viewRecords := make(map[string]time.Time)
    cleanupTicker := time.NewTicker(10 * time.Minute) // 定期清理过期记录
    defer cleanupTicker.Stop()

    for {
        select {
        case event := <-viewChan:
            // 生成唯一的访问标识符
            var key string
            if event.UserID != nil {
                key = fmt.Sprintf("user:%d:post:%d", *event.UserID, event.PostID)
            } else {
                key = fmt.Sprintf("ip:%s:post:%d", event.IP, event.PostID)
            }

            // 检查是否在60秒内已经记录过
            if lastViewTime, exists := viewRecords[key]; exists {
                if time.Since(lastViewTime) < 60*time.Second {
                    // 60秒内已记录过，跳过
                    continue
                }
            }

            // 更新记录时间
            viewRecords[key] = event.Timestamp

            // 更新文章浏览数
            if err := database.DB.Model(&models.Post{}).
                Where("id = ?", event.PostID).
                Update("views", gorm.Expr("views + ?", 1)).Error; err != nil {
                fmt.Printf("更新文章浏览数失败: %v\n", err)
            }

        case <-cleanupTicker.C:
            // 清理60秒前的记录
            cutoffTime := time.Now().Add(-60 * time.Second)
            for key, viewTime := range viewRecords {
                if viewTime.Before(cutoffTime) {
                    delete(viewRecords, key)
                }
            }
        }
    }
}


