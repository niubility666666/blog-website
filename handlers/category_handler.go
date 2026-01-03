package handlers

import (
	"gin-doniai/database"
	"gin-doniai/models"
)

func GetRecommendedCategories() ([]models.Category, error) {
	var categories []models.Category
	err := database.DB.Where("is_recommended = ? AND status_code = ?", 1, 1).Find(&categories).Error
	return categories, err
}
