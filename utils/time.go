package utils

import (
	"fmt"
	"time"
)

func GetTimeAgo(t time.Time) string {
	duration := time.Since(t)

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%d秒前", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%d分钟前", int(duration.Minutes()))
	case duration < time.Hour*24:
		return fmt.Sprintf("%d小时前", int(duration.Hours()))
	case duration < time.Hour*24*30:
		return fmt.Sprintf("%d天前", int(duration.Hours()/24))
	default:
		return t.Format("2025-11-26")
	}
}
