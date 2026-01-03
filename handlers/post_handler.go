package handlers

import (
	"net/http"

	"gin-doniai/database"
	"gin-doniai/models"

	"github.com/gin-gonic/gin"
)

// CreatePost 创建文章
func CreatePost(c *gin.Context) {
    // 从上下文获取用户信息
    userObj, exists := c.Get("user")
    if !exists || userObj == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "用户未登录",
        })
        return
    }
    user := userObj.(*models.User)

    // 解析请求数据
    var requestData struct {
        Title      string `json:"title" binding:"required"`
        CategoryId int    `json:"category_id" binding:"required"`
        Content    string `json:"content" binding:"required"`
        Tags       string `json:"tags"`
        ReadLimit  int    `json:"read_limit"`
    }

    if err := c.ShouldBindJSON(&requestData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "请求参数错误: " + err.Error(),
        })
        return
    }

    // 根据 category_id 查询分类名称
    var category models.Category
    if err := database.DB.First(&category, requestData.CategoryId).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "无效的分类ID",
        })
        return
    }

    // 创建文章对象
    post := models.Post{
        Title:     requestData.Title,
        Category:  category.Name, // 使用查询到的分类名称
        CategoryId: requestData.CategoryId,
        Content:   requestData.Content,
        Tags:      requestData.Tags,
        UserId:    int(user.ID),
        Author:    user.Name,
        ReadLimit: requestData.ReadLimit,
    }

    result := database.DB.Create(&post)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "文章创建失败: " + result.Error.Error(),
        })
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "success": true,
        "message": "文章创建成功",
        "post":    post,
    })
}


// LikePost 文章点赞功能
func LikePost(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	if !exists || userObj == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	user := userObj.(*models.User)

	// 获取文章ID
	id := c.Param("id")
	var post models.Post

	// 查找文章
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "文章不存在",
		})
		return
	}

	// 解析请求数据
	var requestData struct {
		Action string `json:"action" binding:"required,oneof=like unlike"` // 点赞或取消点赞
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户是否已经点赞过（可以通过关联表实现）
	var postLike models.PostLike
	err := database.DB.Where("user_id = ? AND post_id = ?", user.ID, post.ID).First(&postLike).Error

	if requestData.Action == "like" {
		// 点赞操作
		if err != nil {
			// 用户尚未点赞，创建点赞记录
			postLike = models.PostLike{
				UserID: int(user.ID),
				PostID: int(post.ID),
			}

			if err := database.DB.Create(&postLike).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "点赞失败: " + err.Error(),
				})
				return
			}

			// 增加文章点赞数
			database.DB.Model(&post).Update("likes", post.Likes+1)
		}
	} else {
		// 取消点赞操作
		if err == nil {
			// 用户已点赞，删除点赞记录
			if err := database.DB.Delete(&postLike).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "取消点赞失败: " + err.Error(),
				})
				return
			}

			// 减少文章点赞数
			if post.Likes > 0 {
				database.DB.Model(&post).Update("likes", post.Likes-1)
			}
		}
	}

	// 获取最新的点赞数
	var updatedPost models.Post
	database.DB.First(&updatedPost, id)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "操作成功",
		"likes":   updatedPost.Likes,
	})
}

// FavoritePost 文章收藏功能
func FavoritePost(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	if !exists || userObj == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	user := userObj.(*models.User)

	// 获取文章ID
	id := c.Param("id")
	var post models.Post

	// 查找文章
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "文章不存在",
		})
		return
	}

	// 解析请求数据
	var requestData struct {
		Action string `json:"action" binding:"required,oneof=favorite unfavorite"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查用户是否已经收藏过
	var postFavorite models.PostFavorite
	err := database.DB.Where("user_id = ? AND post_id = ?", user.ID, post.ID).First(&postFavorite).Error

	if requestData.Action == "favorite" {
		// 收藏操作
		if err != nil {
			// 用户尚未收藏，创建收藏记录
			postFavorite = models.PostFavorite{
				UserID: int(user.ID),
				PostID: int(post.ID),
			}

			if err := database.DB.Create(&postFavorite).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "收藏失败: " + err.Error(),
				})
				return
			}

			// 增加文章收藏数
			database.DB.Model(&post).Update("favorites", post.Favorites+1)
		}
	} else {
		// 取消收藏操作
		if err == nil {
			// 用户已收藏，删除收藏记录
			if err := database.DB.Delete(&postFavorite).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"message": "取消收藏失败: " + err.Error(),
				})
				return
			}

			// 减少文章收藏数
			if post.Favorites > 0 {
				database.DB.Model(&post).Update("favorites", post.Favorites-1)
			}
		}
	}

	// 获取最新的收藏数
	var updatedPost models.Post
	database.DB.First(&updatedPost, id)

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"message":   "操作成功",
		"favorites": updatedPost.Favorites,
	})
}

// GetPosts 获取所有用户
func GetPosts(c *gin.Context) {
	var posts []models.Post

	result := database.DB.Find(&posts)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"posts": posts,
		"count": len(posts),
	})
}

// GetPost 获取单个文章
func GetPost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	result := database.DB.First(&post, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"post": post})
}

// UpdatePost 更新文章
func UpdatePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	// 先查找文章是否存在
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 绑定更新数据
	var updateData models.Post
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新文章
	result := database.DB.Model(&post).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文章更新成功",
		"post":    post,
	})
}

// DeletePost 删除文章（软删除）
func DeletePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	// 先查找用户是否存在
	if err := database.DB.First(&post, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "文章不存在"})
		return
	}

	// 软删除
	result := database.DB.Delete(&post)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章删除成功"})
}

// 硬删除（永久删除）
func ForceDeletePost(c *gin.Context) {
	id := c.Param("id")
	var post models.Post

	result := database.DB.Unscoped().Delete(&post, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文章永久删除成功"})
}
