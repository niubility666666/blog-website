package caches

import (
    "sync"
    "time"
    "gin-doniai/models"
    "gin-doniai/handlers"
)

// 添加全局缓存变量
var (
    categoryCache     []models.Category
    cacheMutex        sync.RWMutex
    cacheExpiry       time.Time
    cacheDuration     = 10 * time.Minute // 缓存10分钟
)

// 获取缓存的推荐分类
func CachedRecommendedCategories() ([]models.Category, error) {
    cacheMutex.RLock()
    // 检查缓存是否有效
    if time.Now().Before(cacheExpiry) && len(categoryCache) > 0 {
        defer cacheMutex.RUnlock()
        return categoryCache, nil
    }
    cacheMutex.RUnlock()

    // 缓存过期或为空，获取新数据
    cacheMutex.Lock()
    defer cacheMutex.Unlock()

    // 双重检查，防止并发情况下重复获取
    if time.Now().Before(cacheExpiry) && len(categoryCache) > 0 {
        return categoryCache, nil
    }

    categories, err := handlers.GetRecommendedCategories()
    if err != nil {
        return nil, err
    }

    categoryCache = categories
    cacheExpiry = time.Now().Add(cacheDuration)
    return categories, nil
}
