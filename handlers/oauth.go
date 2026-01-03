// handlers/oauth.go
package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"gin-doniai/database"
	"gin-doniai/models"
	"gin-doniai/utils"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"github.com/joho/godotenv"
)

// OAuth配置
var (
	githubOAuthConfig *oauth2.Config
	googleOAuthConfig *oauth2.Config
	oauthStateString  = "oauthstate"
)

// GitHub用户信息结构体
type GitHubUser struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
}

// Google用户信息结构体
type GoogleUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func init() {
    // 加载.env文件
    if err := godotenv.Load(); err != nil {
        fmt.Println("警告: 未能加载 .env 文件")
    }
	// GitHub OAuth配置
	githubOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"), // 如: http://localhost:8080/auth/github/callback
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	// Google OAuth配置
	googleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"), // 如: http://localhost:8080/auth/google/callback
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// 生成随机state字符串
func generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// GitHub授权登录处理
func GitHubLogin(c *gin.Context) {
	state := generateState()
	session := sessions.Default(c)
	session.Set(oauthStateString, state)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	url := githubOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GitHub回调处理
func GitHubCallback(c *gin.Context) {
	session := sessions.Default(c)

	// 验证state参数
	state := c.Query("state")
	savedState := session.Get(oauthStateString)
	if state == "" || savedState == nil || state != savedState.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// 获取授权码
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
		return
	}

	// 交换access token
	token, err := githubOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 获取用户信息
	client := githubOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var githubUser GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&githubUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// 处理用户登录/注册
	user, err := handleOAuthUserLogin(c, githubUser.Login, githubUser.Email, githubUser.Name, githubUser.AvatarURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user login"})
		return
	}

	// 设置session
	session.Set("user_id", user.ID)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// 重定向到首页
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// Google授权登录处理
func GoogleLogin(c *gin.Context) {
	state := generateState()
	session := sessions.Default(c)
	session.Set(oauthStateString, state)
	err := session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	url := googleOAuthConfig.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// Google回调处理
func GoogleCallback(c *gin.Context) {
	session := sessions.Default(c)

	// 验证state参数
	state := c.Query("state")
	savedState := session.Get(oauthStateString)
	if state == "" || savedState == nil || state != savedState.(string) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	// 获取授权码
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code parameter"})
		return
	}

	// 交换access token
	token, err := googleOAuthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange token"})
		return
	}

	// 获取用户信息
	client := googleOAuthConfig.Client(context.Background(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}
	defer resp.Body.Close()

	var googleUser GoogleUser
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode user info"})
		return
	}

	// 处理用户登录/注册
	user, err := handleOAuthUserLogin(c, googleUser.Email, googleUser.Email, googleUser.Name, googleUser.Picture)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process user login"})
		return
	}

	// 设置session
	session.Set("user_id", user.ID)
	err = session.Save()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}

	// 重定向到首页
	c.Redirect(http.StatusTemporaryRedirect, "/")
}

// 处理OAuth用户登录/注册的通用函数
func handleOAuthUserLogin(c *gin.Context, identifier, email, name, avatarURL string) (*models.User, error) {
	var user models.User

	// 尝试查找现有用户
	if email != "" {
		// 优先通过邮箱查找
		database.DB.Where("email = ?", email).First(&user)
	}

	if user.ID == 0 && identifier != "" {
		// 通过标识符查找
		database.DB.Where("name = ?", identifier).First(&user)
	}

	// 如果用户不存在，则创建新用户
	if user.ID == 0 {
		// 生成随机密码
		passwordBytes := make([]byte, 32)
		rand.Read(passwordBytes)
		randomPassword := base64.URLEncoding.EncodeToString(passwordBytes)

		hashedPassword, err := utils.HashPassword(randomPassword)
		if err != nil {
			return nil, err
		}

		// 如果没有名字，使用标识符
		if name == "" {
			name = identifier
		}

		// 如果没有头像，生成默认头像
		if avatarURL == "" {
			avatarURL = fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random", url.QueryEscape(name))
		}

		user = models.User{
			Name:       name,
			Email:      email,
			Password:   hashedPassword,
			AgreeTerms: true, // OAuth用户默认同意条款
			Avatar:     avatarURL,
		}

		if err := database.DB.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	return &user, nil
}
