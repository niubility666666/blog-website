package handlers

// 需要添加正确的导入
import (
	"fmt"
	"gin-doniai/database"
	"gin-doniai/models"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// createComment 创建评论
func CreateComment(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if !exists || userObj == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}

    // 添加额外的类型检查
    user, ok := userObj.(*models.User)
    if !ok || user == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "用户未登录",
        })
        return
    }

	// 解析请求数据
	var requestData struct {
		Content  string `json:"content" binding:"required"`
		PostID   uint   `json:"post_id" binding:"required"`
		ParentID uint   `json:"parent_id" binding:"omitempty"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 解析和转换评论内容
	processedContent := processCommentContent(requestData.Content)

	// 创建评论对象
	comment := models.Comment{
		Content:  processedContent,
		PostID:   requestData.PostID,
		UserID:   user.ID,
		ParentID: requestData.ParentID,
	}

	// 保存到数据库
	if err := database.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "评论创建失败: " + err.Error(),
		})
		return
	}

	// 更新帖子的回复数
	database.DB.Model(&models.Post{}).Where("id = ?", requestData.PostID).UpdateColumn("replies", gorm.Expr("replies + ?", 1))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "评论发表成功",
		"data":    comment,
	})
}

// getComments 获取评论列表
func GetComments(c *gin.Context) {
	postID := c.Query("post_id")
	if postID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少post_id参数",
		})
		return
	}

	var comments []models.Comment
	if err := database.DB.Where("post_id = ?", postID).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comments,
	})
}

// getComment 获取单个评论
func GetComment(c *gin.Context) {
	id := c.Param("id")

	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "评论未找到",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    comment,
	})
}

// updateComment 更新评论
func UpdateComment(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if !exists || userObj == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	user = userObj.(*models.User)

	id := c.Param("id")

	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "评论未找到",
		})
		return
	}

	// 检查是否有权限更新评论（必须是评论作者）
	if comment.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权限更新此评论",
		})
		return
	}

	var requestData struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	comment.Content = requestData.Content
	if err := database.DB.Save(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "更新评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "评论更新成功",
		"data":    comment,
	})
}

// deleteComment 删除评论
func DeleteComment(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if !exists || userObj == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "用户未登录",
		})
		return
	}
	user = userObj.(*models.User)

	id := c.Param("id")

	var comment models.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "评论未找到",
		})
		return
	}

	// 检查是否有权限删除评论（必须是评论作者或管理员）
	if comment.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "无权限删除此评论",
		})
		return
	}

	if err := database.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除评论失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "评论删除成功",
	})
}

// LikeComment 评论点赞
func LikeComment(c *gin.Context) {
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

	// 获取评论ID
	commentId := c.Param("id")

	// 解析请求数据
	var requestData struct {
		Action string `json:"action"` // like 或 unlike
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 查询评论
	var comment models.Comment
	if err := database.DB.First(&comment, commentId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "评论未找到",
		})
		return
	}

	// 检查用户是否在给自己点赞
	if comment.UserID == user.ID && requestData.Action == "like" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不能给自己的评论点赞",
		})
		return
	}

	// 更新点赞数
	if requestData.Action == "like" {
		database.DB.Model(&comment).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1))
	} else if requestData.Action == "unlike" {
		database.DB.Model(&comment).UpdateColumn("like_count", gorm.Expr("like_count - ?", 1))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "操作成功",
		"data": gin.H{
			"like_count": comment.LikeCount,
		},
	})
}

func processCommentContent(content string) string {
	// 使用正则表达式匹配 @用户名 和 #ID 模式
	// 匹配 @用户名 (假设用户名不包含空格)
	reUser := regexp.MustCompile(`@(\S+)`)
	content = reUser.ReplaceAllString(content, `<a href="/user/$1" target="_blank">@$1</a>`)

	// 匹配 #ID (数字)
	reComment := regexp.MustCompile(`#(\d+)`)
	content = reComment.ReplaceAllString(content, `<a href="#comment-$1">#$1</a>`)

	// 将整个内容包装在 <p> 标签中
	return fmt.Sprintf("<p>%s</p>", content)
}
