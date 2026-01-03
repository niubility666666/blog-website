package handlers

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "net/http"
    "time"
    "gin-doniai/database"
    "gin-doniai/models"
    "gin-doniai/utils"
    "github.com/gin-gonic/gin"
)

func ForgotPassword(c *gin.Context) {
    var requestData struct {
        Email string `json:"email" binding:"required,email"`
    }

    // 绑定请求数据
    if err := c.ShouldBindJSON(&requestData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "请求数据格式错误",
        })
        return
    }

    // 检查用户是否存在
    var user models.User
    if err := database.DB.Where("email = ?", requestData.Email).First(&user).Error; err != nil {
        // 为了安全起见，即使用户不存在也返回成功消息
        c.JSON(http.StatusOK, gin.H{
            "success": true,
            "message": "如果该邮箱存在，重置密码的链接已发送到您的邮箱",
        })
        return
    }

    // 生成重置令牌
    token := generateResetToken()

    // 设置令牌过期时间为1小时
    expiresAt := time.Now().Add(time.Hour)

    // 创建密码重置记录
    passwordReset := models.PasswordReset{
        Email:     requestData.Email,
        Token:     token,
        ExpiresAt: expiresAt,
    }

    // 保存到数据库
    if err := database.DB.Create(&passwordReset).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "服务器内部错误，请稍后重试",
        })
        return
    }

    // 构建重置链接
    resetLink := fmt.Sprintf("http://%s/reset-password?token=%s", c.Request.Host, token)

    // TODO: 发送邮件（这里只是模拟）
    // 实际项目中应该集成邮件服务，如SMTP
    fmt.Printf("发送重置密码邮件到: %s\n重置链接: %s\n", requestData.Email, resetLink)

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "重置密码的链接已发送到您的邮箱，请查收",
    })
}

// ResetPassword 处理重置密码请求
func ResetPassword(c *gin.Context) {
    token := c.Query("token")
    if token == "" {
        c.HTML(http.StatusBadRequest, "reset-password.tmpl", gin.H{
            "error": "无效的重置链接",
        })
        return
    }

    // 查找重置记录
    var passwordReset models.PasswordReset
    if err := database.DB.Where("token = ? AND used = ?", token, false).First(&passwordReset).Error; err != nil {
        c.HTML(http.StatusBadRequest, "reset-password.tmpl", gin.H{
            "error": "重置链接无效或已过期",
        })
        return
    }

    // 检查令牌是否过期
    if time.Now().After(passwordReset.ExpiresAt) {
        c.HTML(http.StatusBadRequest, "reset-password.tmpl", gin.H{
            "error": "重置链接已过期",
        })
        return
    }

    // 渲染重置密码页面
    c.HTML(http.StatusOK, "reset-password.tmpl", gin.H{
        "token": token,
    })
}

// ProcessResetPassword 处理重置密码表单提交
func ProcessResetPassword(c *gin.Context) {
    var requestData struct {
        Token    string `json:"token" binding:"required"`
        Password string `json:"password" binding:"required,min=6"`
    }

    // 绑定请求数据
    if err := c.ShouldBindJSON(&requestData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "请求数据格式错误",
        })
        return
    }

    // 查找重置记录
    var passwordReset models.PasswordReset
    if err := database.DB.Where("token = ? AND used = ?", requestData.Token, false).First(&passwordReset).Error; err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "重置链接无效或已过期",
        })
        return
    }

    // 检查令牌是否过期
    if time.Now().After(passwordReset.ExpiresAt) {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "message": "重置链接已过期",
        })
        return
    }

    // 查找用户
    var user models.User
    if err := database.DB.Where("email = ?", passwordReset.Email).First(&user).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "用户不存在",
        })
        return
    }

    // 加密新密码
    hashedPassword, err := utils.HashPassword(requestData.Password)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "密码加密失败",
        })
        return
    }

    // 更新用户密码
    if err := database.DB.Model(&user).Update("password", hashedPassword).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "message": "更新密码失败",
        })
        return
    }

    // 标记令牌为已使用
    if err := database.DB.Model(&passwordReset).Update("used", true).Error; err != nil {
        // 即使标记失败也不影响主要流程
        fmt.Printf("标记密码重置令牌为已使用失败: %v\n", err)
    }

    // 返回成功响应
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "密码重置成功，您可以使用新密码登录了",
    })
}

func generateResetToken() string {
    return generateUUID()
}

func generateSecureToken() string {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return generateUUID()
    }
    return base64.URLEncoding.EncodeToString(bytes)
}

func generateUUID() string {
    uuid := make([]byte, 16)
    rand.Read(uuid)
    uuid[6] = (uuid[6] & 0x0f) | 0x40
    uuid[8] = (uuid[8] & 0x3f) | 0x80
    return fmt.Sprintf("%x-%x-%x-%x-%x",
        uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
