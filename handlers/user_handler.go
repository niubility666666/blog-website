package handlers

import (
	"net/http"

	"gin-doniai/database"
	"gin-doniai/models"
    "gin-doniai/utils"
	"github.com/gin-gonic/gin"
)

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := database.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "用户创建成功",
		"user":    user,
	})
}

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	var users []models.User

	result := database.DB.Find(&users)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// GetUser 获取单个用户
func GetUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	result := database.DB.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	// 先查找用户是否存在
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 绑定更新数据
	var updateData models.User
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新用户
	result := database.DB.Model(&user).Updates(updateData)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "用户更新成功",
		"user":    user,
	})
}

// DeleteUser 删除用户（软删除）
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	// 先查找用户是否存在
	if err := database.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	// 软删除
	result := database.DB.Delete(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户删除成功"})
}

// 硬删除（永久删除）
func ForceDeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User

	result := database.DB.Unscoped().Delete(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "用户永久删除成功"})
}

// UpdateUserProfile 更新用户资料
func UpdateUserProfile(c *gin.Context) {
    // 从上下文获取当前用户
    userObj, exists := c.Get("user")
    if !exists || userObj == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "用户未登录",
        })
        return
    }

    currentUser := userObj.(*models.User)

    // 绑定请求数据
    var updateData struct {
        Motto         string `json:"motto"`
        Github        string `json:"github"`
        GoogleAccount string `json:"google_account"`
    }

    if err := c.ShouldBindJSON(&updateData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "请求数据格式错误",
        })
        return
    }

    // 更新用户信息
    updates := make(map[string]interface{})
    if updateData.Motto != "" {
        updates["motto"] = updateData.Motto
    }
    if updateData.Github != "" {
        updates["github"] = updateData.Github
    }
    if updateData.GoogleAccount != "" {
        updates["google_account"] = updateData.GoogleAccount
    }

    if err := database.DB.Model(&models.User{}).Where("id = ?", currentUser.ID).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "更新失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "个人信息更新成功",
    })
}

// UpdateUserPassword 修改用户密码
func UpdateUserPassword(c *gin.Context) {
    // 从上下文获取当前用户
    userObj, exists := c.Get("user")
    if !exists || userObj == nil {
        c.JSON(http.StatusUnauthorized, gin.H{
            "success": false,
            "message": "用户未登录",
        })
        return
    }

    currentUser := userObj.(*models.User)

    // 绑定请求数据
    var passwordData struct {
        CurrentPassword string `json:"current_password" binding:"required"`
        NewPassword     string `json:"new_password" binding:"required"`
    }

    if err := c.ShouldBindJSON(&passwordData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "请求数据格式错误",
        })
        return
    }

    // 验证当前密码是否正确
    if !utils.CheckPassword(passwordData.CurrentPassword, currentUser.Password) {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "message": "当前密码错误",
        })
        return
    }

    // 验证新密码长度
    if len(passwordData.NewPassword) < 6 {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "新密码长度至少6位",
        })
        return
    }

    // 加密新密码
    hashedPassword, err := utils.HashPassword(passwordData.NewPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "密码加密失败",
        })
        return
    }

    // 更新密码
    if err := database.DB.Model(&models.User{}).Where("id = ?", currentUser.ID).Update("password", hashedPassword).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "密码更新失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "密码修改成功",
    })
}

// UserIDFromContext 从上下文中获取用户ID
func UserIDFromContext(c *gin.Context) *uint {
    userObj, exists := c.Get("user")
    if exists && userObj != nil {
        if user, ok := userObj.(*models.User); ok {
            return &user.ID
        }
    }
    return nil
}
