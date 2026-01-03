package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
	"encoding/xml"
    "gin-doniai/middlewares"
	"gin-doniai/caches"
	"gin-doniai/database"
	"gin-doniai/handlers"
	"gin-doniai/models"
	"gin-doniai/utils"
	"gin-doniai/workers"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// 在 main.go 顶部添加全局变量
var (
	onlineStatusChan      chan workers.OnlineStatusUpdate
    viewEventChan chan workers.ViewEvent
	globalConfig          GlobalConfig
	recommendedCategories []models.Category
)

type GlobalConfig struct {
	SiteName   string
	Theme      string
	Version    string
	Categories []models.Category
}

type OnlineStatusUpdate struct {
	UserID    uint
	IP        string
	UserAgent string
}

func main() {
	database.InitDB()

	// 初始化全局配置
	globalConfig = GlobalConfig{
		SiteName: "Doniai",
		Theme:    "light",
		Version:  "1.0.0",
	}

	// 获取推荐分类并注入到全局配置
	if categories, err := caches.CachedRecommendedCategories(); err == nil {
		recommendedCategories = categories
		globalConfig.Categories = categories
	} else {
		fmt.Printf("获取推荐分类失败: %v\n", err)
	}

	gin.SetMode(gin.DebugMode)
    // gin.SetMode(gin.ReleaseMode)
	// 初始化在线状态更新通道
	onlineStatusChan = make(chan workers.OnlineStatusUpdate, 1000) // 缓冲1000个消息
	// 启动在线状态更新处理器
    go workers.HandleOnlineStatusUpdates(onlineStatusChan)

    viewEventChan = make(chan workers.ViewEvent, 1000)  // 缓冲1000个消息

    // 启动浏览事件处理器
    go workers.HandleViewNumUpdates(viewEventChan)

	router := gin.Default()
	router.SetFuncMap(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"loop": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
        "currentYear": func() int {
            return time.Now().Year()
        },
        "timeAgo": func(t time.Time) string {
            return utils.GetTimeAgo(t)
        },
		"global": func() GlobalConfig {
			return globalConfig
		},
	})
	// 设置session存储
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("mysession", store))
	// 在路由定义之前应用用户中间件
	router.Use(middlewares.UserAndOnlineStatusMiddleware(onlineStatusChan))

	// 加载模板文件
	router.LoadHTMLGlob("templates/**/*")

	// 静态文件服务
	router.Static("/static", "./static")

	// 路由定义
	router.GET("/", homeHandler)
	router.GET("/categories/:type", homeHandler)
	router.GET("/about", aboutHandler)
	router.GET("/post-:id-1", detailHandler)
	router.GET("/register", registerHandler)
	router.POST("/register", registerSubmit)
	router.GET("/login", loginHandler)
	router.POST("/login", loginSubmit)
	router.GET("/logout", logoutHandler)
	router.GET("/profile", profileHandler)
	router.GET("/posts", articleHandler)
	router.GET("/publish", publishHandler)
	router.GET("/settings", settingsHandler)
	router.GET("/rss", rssHandler)
	// 添加搜索路由
	router.GET("/search", searchPostsHandler)
	router.GET("/member", searchUsersHandler)

	// 在 main.go 的路由定义部分添加
    router.GET("/auth/github", handlers.GitHubLogin)
    router.GET("/auth/github/callback", handlers.GitHubCallback)
    router.GET("/auth/google", handlers.GoogleLogin)
    router.GET("/auth/google/callback", handlers.GoogleCallback)


	// 在 main.go 的路由部分添加
	router.GET("/api/online/count", handlers.GetOnlineUserCount)
	// 在路由定义部分添加
    router.POST("/api/auth/forgot-password", handlers.ForgotPassword)
    router.GET("/reset-password", handlers.ResetPassword)
    router.POST("/api/auth/reset-password", handlers.ProcessResetPassword)


	// 在 main.go 的路由定义部分添加评论路由
	commentRoutes := router.Group("/api/comments")
	{
		commentRoutes.POST("/", handlers.CreateComment)
		commentRoutes.GET("/", handlers.GetComments)
		commentRoutes.GET("/:id", handlers.GetComment)
		commentRoutes.PUT("/:id", handlers.UpdateComment)
		commentRoutes.DELETE("/:id", handlers.DeleteComment)
		commentRoutes.POST("/:id/like", handlers.LikeComment)
	}

	userRoutes := router.Group("/api/users")
	{
		userRoutes.POST("/", handlers.CreateUser)                 // 创建用户
		userRoutes.GET("/", handlers.GetUsers)                    // 获取所有用户
		userRoutes.GET("/:id", handlers.GetUser)                  // 获取单个用户
		userRoutes.PUT("/:id", handlers.UpdateUser)               // 更新用户
		userRoutes.DELETE("/:id", handlers.DeleteUser)            // 删除用户（软删除）
		userRoutes.DELETE("/:id/force", handlers.ForceDeleteUser) // 强制删除
		userRoutes.PUT("/profile", handlers.UpdateUserProfile)    // 更新用户资料
        userRoutes.PUT("/password", handlers.UpdateUserPassword) // 修改用户密码
	}

	postRoutes := router.Group("/api/posts")
	{
		postRoutes.POST("/", handlers.CreatePost)                 // 创建文章
		postRoutes.GET("/", handlers.GetPosts)                    // 获取所有文章
		postRoutes.GET("/:id", handlers.GetPost)                  // 获取单个文章
		postRoutes.PUT("/:id", handlers.UpdatePost)               // 更新文章
		postRoutes.DELETE("/:id", handlers.DeletePost)            // 删除文章（软删除）
		postRoutes.POST("/:id/like", handlers.LikePost)           // 文章点赞
		postRoutes.DELETE("/:id/force", handlers.ForceDeletePost) // 强制删除
		postRoutes.POST("/:id/favorite", handlers.FavoritePost)   // 文章收藏
	}

    router.NoRoute(func(c *gin.Context) {
        c.HTML(http.StatusNotFound, "404.tmpl", gin.H{
            "Message": "页面未找到",
        })
    })

	// 启动定时清理任务
	go func() {
		ticker := time.NewTicker(10 * time.Minute) // 每10分钟清理一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				handlers.CleanupExpiredOnlineStatus()
			}
		}
	}()
	router.Run(":8080")
}


func homeHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		fmt.Println("未获取到用户信息")
	}
	// 获取页码参数，默认为第1页
	pageStr := c.Query("page")
	// 获取路由categories/:type中type参数
	categoryType := c.Param("type")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 每页显示的帖子数量
	limit := 10
	offset := (page - 1) * limit

	// 查询总记录数
	var total int64
	dbQuery := database.DB.Model(&models.Post{})

	var categoryId uint
	if categoryType != "" {
		var category models.Category
	    if err := database.DB.Where("alias = ?", categoryType).First(&category).Error; err != nil {
            // 当找不到分类时，返回404页面而不是继续执行
            c.HTML(http.StatusNotFound, "404.tmpl", gin.H{
                "Message": "分类未找到",
            })
            return
        } else {
            categoryId = category.ID
            dbQuery = dbQuery.Where("category_id = ?", categoryId)
        }
	}
	dbQuery.Count(&total)

	// 查询当前页的帖子
	var posts []models.Post
	postQuery := database.DB.Where("category_id > ?", 0).Order("created_at DESC").Offset(offset).Limit(limit)

	if categoryId > 0 {
		postQuery = postQuery.Where("category_id = ?", categoryId)
	}
	postQuery.Find(&posts)

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// 创建带有友好时间的帖子结构
	type PostWithFriendlyTime struct {
		models.Post
		TimeAgo string
	}

	var postsWithTimeAgo []PostWithFriendlyTime
	for _, post := range posts {
		timeAgo := utils.GetTimeAgo(post.CreatedAt)
		postsWithTimeAgo = append(postsWithTimeAgo, PostWithFriendlyTime{
			Post:    post,
			TimeAgo: timeAgo,
		})
	}

	// 获取在线用户数
	var onlineCount int64
	cutoffTime := time.Now().Add(-30 * time.Minute)
	database.DB.Model(&models.UserOnlineStatus{}).
		Where("last_active_time > ?", cutoffTime).
		Count(&onlineCount)

	// 获取站点统计信息
	var userCount, postCount, commentCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	database.DB.Model(&models.Post{}).Count(&postCount)
	database.DB.Model(&models.Comment{}).Count(&commentCount)

	// 获取所有分类
	var categories []models.Category
	database.DB.Where("status_code = ?", 1).Find(&categories)

	data := gin.H{
		"CurrentTime":  time.Now().Format("2006-01-02 15:04:05"),
		"posts":        postsWithTimeAgo,
		"currentPage":  page,
		"totalPages":   totalPages,
		"hasPrev":      page > 1,
		"hasNext":      page < totalPages,
		"prevPage":     page - 1,
		"nextPage":     page + 1,
		"user":         user,
		"userCount":    userCount,
		"postCount":    postCount,
		"commentCount": commentCount,
		"onlineCount":  onlineCount,
		"categories":   categories,
	}

	c.HTML(http.StatusOK, "home.tmpl", data)
}

func aboutHandler(c *gin.Context) {
	data := gin.H{
		"TeamMembers": []struct {
			Name string
			Role string
		}{
			{"张三", "开发工程师"},
			{"李四", "UI设计师"},
			{"王五", "产品经理"},
		},
	}
	c.HTML(http.StatusOK, "about.tmpl", data)
}

func detailHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	}
	// 从路由中获取文章ID
	idWithSuffix := c.Param("id-1") // 获取 "29-1"
	// 分割字符串获取真正的ID
	idParts := strings.Split(idWithSuffix, "-")
	var id string
	if len(idParts) > 0 {
		id = idParts[0] // 获取 "29"
	}

    UserId := handlers.UserIDFromContext(c)
    postId, err := strconv.ParseUint(id, 10, 32)
    if err != nil {
        postId = 0
    }
    // 发送浏览事件
    viewEvent := workers.ViewEvent{
        PostID:     uint(postId),
        UserID:    UserId,
        IP:        c.ClientIP(),
        UserAgent: c.Request.UserAgent(),
        Timestamp: time.Now(),
    }

    select {
    case viewEventChan <- viewEvent:
    default:
        fmt.Println("浏览事件通道已满")
    }

	// 查询数据库获取文章详情，并预加载用户信息
	var post models.Post
	if err := database.DB.Preload("User").First(&post, id).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.tmpl", gin.H{
			"Message": "文章未找到",
		})
		return
	}

	// 创建带有友好时间和回复评论的评论结构
	type CommentWithReplies struct {
		models.Comment
		TimeAgo string
		Replies []CommentWithReplies
		Content template.HTML
	}

	// 获取评论页码参数
	commentPageStr := c.Query("page")
	commentPage := 1
	if commentPageStr != "" {
		if p, err := strconv.Atoi(commentPageStr); err == nil && p > 0 {
			commentPage = p
		}
	}

	// 每页显示的评论数量
	commentLimit := 4
	commentOffset := (commentPage - 1) * commentLimit

	// 查询总评论数
	var totalComments int64
	database.DB.Model(&models.Comment{}).Where("post_id = ? AND parent_id = 0", id).Count(&totalComments)

	// 计算总页数
	totalCommentPages := int((totalComments + int64(commentLimit) - 1) / int64(commentLimit))

	// 查询该文章的评论（仅顶级评论），并预加载用户信息
	var comments []models.Comment
	database.DB.Where("post_id = ? AND parent_id = 0", id).Preload("User").
		Offset(commentOffset).Limit(commentLimit).Order("created_at DESC").Find(&comments)

	// 创建带有回复和友好时间的评论列表
	var commentsWithReplies []CommentWithReplies
	for _, comment := range comments {
		// 处理主评论的友好时间
		timeAgo := utils.GetTimeAgo(comment.CreatedAt)

		// 查询该评论的回复
		var replies []models.Comment
		database.DB.Where("parent_id = ?", comment.ID).Preload("User").Find(&replies)

		// 处理回复评论的友好时间
		var repliesWithTime []CommentWithReplies
		for _, reply := range replies {
			replyTimeAgo := utils.GetTimeAgo(reply.CreatedAt)
			repliesWithTime = append(repliesWithTime, CommentWithReplies{
				Comment: reply,
				TimeAgo: replyTimeAgo,
				Content: template.HTML(reply.Content),
			})
		}

		commentsWithReplies = append(commentsWithReplies, CommentWithReplies{
			Comment: comment,
			TimeAgo: timeAgo,
			Replies: repliesWithTime,
			Content: template.HTML(comment.Content),
		})
	}

	// 将标签字符串分割成数组
	// 使用工具方法处理标签
	tags := utils.ParseTags(post.Tags)
	// 将文章详情数据和用户信息传递给模板
	var postCount, replyCount, likeCount int64
	database.DB.Model(&models.Post{}).Where("user_id = ?", post.User.ID).Count(&postCount)
	database.DB.Model(&models.Post{}).Where("user_id = ?", post.User.ID).Select("SUM(replies)").Row().Scan(&replyCount)
	database.DB.Model(&models.Post{}).Where("user_id = ?", post.User.ID).Select("SUM(likes)").Row().Scan(&likeCount)
	// 将评论分页信息添加到模板数据
    fmt.Printf("当前文章ID: %s, 分类ID: %d\n", id, post.CategoryId)
	// 搜索3条相关的文章数据
	var relatedPosts []models.Post
    database.DB.Where("id != ? AND category_id = ?", id, post.CategoryId).
        Order("RAND()").
        Limit(3).
        Find(&relatedPosts)
	data := gin.H{
		"Post":               post,
		"User":               post.User,
		"Content":            template.HTML(post.Content),
		"Tags":               tags,
		"user":               user,
		"Comments":           commentsWithReplies,
		"commentCurrentPage": commentPage,
		"commentTotalPages":  totalCommentPages,
		"commentHasPrev":     commentPage > 1,
		"commentHasNext":     commentPage < totalCommentPages,
		"commentPrevPage":    commentPage - 1,
		"commentNextPage":    commentPage + 1,
		"commentTotalCount":  totalComments,
		"postCount":          postCount,
		"replyCount":         replyCount,
		"likeCount":          likeCount,
		"RelatedPosts":       relatedPosts,
	}

	c.HTML(http.StatusOK, "detail.tmpl", data)
}

func profileHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		// 用户未登录，重定向到登录页面
		c.Redirect(http.StatusFound, "/login")
		return
	}

    var postCount int64
    database.DB.Model(&models.Post{}).Where("category_id > ?", 0).Where("user_id = ?", user.ID).Count(&postCount)

	data := gin.H{
		"user":        user,
		"profileUser": user,
		"postCount":   postCount,
	}
	c.HTML(http.StatusOK, "profile.tmpl", data)
}

func articleHandler(c *gin.Context) {
    // 从上下文获取用户信息
    userObj, exists := c.Get("user")
    var user *models.User
    if exists && userObj != nil {
        user = userObj.(*models.User)
    } else {
        // 用户未登录，重定向到登录页面
        c.Redirect(http.StatusFound, "/login")
        return
    }

    // 获取tab参数
    tab := c.Query("tab")
    if tab == "" {
        tab = "articles" // 默认tab
    }

    // 获取页码参数，默认为第1页
    pageStr := c.Query("page")
    page := 1
    if pageStr != "" {
        if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
            page = p
        }
    }

    // 每页显示的数量
    limit := 5
    offset := (page - 1) * limit

    // 1. 根据用户id，查询用户的帖子，从posts表查
    var userPosts []models.Post
    var totalUserPosts int64
    database.DB.Model(&models.Post{}).Where("category_id > ?", 0).Where("user_id = ?", user.ID).Count(&totalUserPosts)
    database.DB.Where("category_id > ?", 0).Where("user_id = ?", user.ID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&userPosts)

    // 2. 根据用户id，查询用户的评论，从comments表查,并通过post_id关联查出posts表中的title
    type CommentWithPostTitle struct {
        models.Comment
        PostTitle string
        TimeAgo   string
    }

    var userComments []CommentWithPostTitle
    var totalUserComments int64

    // 先查询总数
    database.DB.Table("comments c").
        Joins("LEFT JOIN posts p ON c.post_id = p.id").
        Where("c.user_id = ?", user.ID).
        Count(&totalUserComments)

    // 查询评论及关联的文章标题
    database.DB.Table("comments c").
        Select("c.*, p.title as post_title").
        Joins("LEFT JOIN posts p ON c.post_id = p.id").
        Where("c.user_id = ?", user.ID).
        Order("c.created_at DESC").
        Offset(offset).
        Limit(limit).
        Scan(&userComments)

    // 处理评论的时间显示
    for i := range userComments {
        userComments[i].TimeAgo = utils.GetTimeAgo(userComments[i].CreatedAt)
    }

    // 3. 根据用户id，查询用户收藏的帖子，从post_favorites表查,并通过post_id关联查出posts表中的title
    type FavoritePost struct {
        models.Post
        TimeAgo string
    }

    var favoritePosts []FavoritePost
    var totalFavoritePosts int64

    // 先查询总数
    database.DB.Table("post_favorites pf").
        Joins("LEFT JOIN posts p ON pf.post_id = p.id").
        Where("pf.user_id = ?", user.ID).
        Count(&totalFavoritePosts)

    // 查询收藏的文章
    database.DB.Table("post_favorites pf").
        Select("p.*, pf.created_at as favorite_time").
        Joins("LEFT JOIN posts p ON pf.post_id = p.id").
        Where("pf.user_id = ?", user.ID).
        Order("pf.created_at DESC").
        Offset(offset).
        Limit(limit).
        Scan(&favoritePosts)

    // 处理收藏文章的时间显示
    for i := range favoritePosts {
        favoritePosts[i].TimeAgo = utils.GetTimeAgo(favoritePosts[i].CreatedAt)
    }

    // 计算总页数
    totalPostPages := int((totalUserPosts + int64(limit) - 1) / int64(limit))
    totalCommentPages := int((totalUserComments + int64(limit) - 1) / int64(limit))
    totalFavoritePages := int((totalFavoritePosts + int64(limit) - 1) / int64(limit))

    data := gin.H{
        "user":              user,
        "profileUser":       user,
        "articles":          userPosts,
        "comments":          userComments,
        "favorites":         favoritePosts,
        "currentPage":       page,
        "currentTab":        tab, // 添加当前tab信息
        "totalUserPosts":    totalUserPosts,
        "totalUserComments": totalUserComments,
        "totalFavoritePosts":totalFavoritePosts,
        "totalPostPages":    totalPostPages,
        "totalCommentPages": totalCommentPages,
        "totalFavoritePages": totalFavoritePages,
        "hasPrev":           page > 1,
        "hasNext":           page < getTotalPagesForTab(tab, totalPostPages, totalCommentPages, totalFavoritePages),
        "prevPage":          page - 1,
        "nextPage":          page + 1,
    }
    c.HTML(http.StatusOK, "article-list.tmpl", data)
}

// 辅助函数：根据tab获取对应的总页数
func getTotalPagesForTab(tab string, postPages, commentPages, favoritePages int) int {
    switch tab {
    case "articles":
        return postPages
    case "comments":
        return commentPages
    case "favorites":
        return favoritePages
    default:
        return postPages
    }
}



func settingsHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		// 用户未登录，重定向到登录页面
		c.Redirect(http.StatusFound, "/login")
		return
	}

	data := gin.H{
		"user": user,
	}
	c.HTML(http.StatusOK, "settings.tmpl", data)
}

func publishHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		// 用户未登录，重定向到登录页面
		c.Redirect(http.StatusFound, "/login")
		return
	}

	// 获取所有分类
	var categories []models.Category
	database.DB.Where("status_code = ?", 1).Find(&categories)

	data := gin.H{
		"user":       user,
		"categories": categories,
	}
	c.HTML(http.StatusOK, "publish.tmpl", data)
}

func registerHandler(c *gin.Context) {
	data := gin.H{
		"CurrentPath": "/register",
	}
	c.HTML(http.StatusOK, "auth.tmpl", data)
}

func loginHandler(c *gin.Context) {
	data := gin.H{
		"CurrentPath": "/login",
	}
	c.HTML(http.StatusOK, "auth.tmpl", data)
}

func logoutHandler(c *gin.Context) {
	// 获取session
	session := sessions.Default(c)

	// 清除session中的用户信息
	session.Clear()

	// 保存session更改
	if err := session.Save(); err != nil {
		fmt.Printf("登出时Session保存失败: %v\n", err)
	} else {
		fmt.Println("用户已成功登出")
	}

	// 重定向到首页
	c.Redirect(http.StatusFound, "/")
}

func loginSubmit(c *gin.Context) {

	// 测试密码：xZ3(Uq)sDQ6qYEY]
	// 获取表单提交的数据
	identifier := c.PostForm("email") // 可以是邮箱或用户名
	password := c.PostForm("password")
	remember := c.PostForm("remember")

	// 基本验证
	if identifier == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "用户名/邮箱和密码不能为空",
		})
		return
	}

	// 查询用户（支持邮箱或用户名登录）
	var user models.User
	if err := database.DB.Where("email = ? OR name = ?", identifier, identifier).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "用户不存在",
		})
		return
	}

	// 验证密码
	if !utils.CheckPassword(password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"status":  "error",
			"message": "密码错误",
		})
		return
	}

	// 获取session
	session := sessions.Default(c)

	// 设置用户信息到session
	session.Set("user_id", user.ID)

	// 根据"记住密码"选项设置过期时间
	if remember == "on" {
		// 设置30天过期
		session.Options(sessions.Options{
			MaxAge: 30 * 24 * 60 * 60, // 30天
		})
	} else {
		// 设置会话结束时失效（浏览器关闭时）
		session.Options(sessions.Options{
			MaxAge: 0, // 浏览器会话期间有效
		})
	}

	// 保存session

	// 保存session
	if err := session.Save(); err != nil {
		fmt.Printf("Session保存失败: %v\n", err)
	} else {
		fmt.Println("Session保存成功")
	}

	// 登录成功
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "登录成功",
		"data": gin.H{
			"user_id": user.ID,
			"email":   user.Email,
			"name":    user.Name,
		},
	})

}

func registerSubmit(c *gin.Context) {
	// 获取表单提交的数据
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	confirmPassword := c.PostForm("confirmPassword")
	agreeTerms := c.PostForm("agreeTerms")

	// 基本验证
	if username == "" || email == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "用户名、邮箱和密码不能为空",
		})
		return
	}

	if password != confirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "两次输入的密码不一致",
		})
		return
	}

	// 处理用户协议同意状态
	isAgreeTerms := false
	if agreeTerms == "on" {
		isAgreeTerms = true
	}

	if !isAgreeTerms {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "请同意用户协议",
		})
		return
	}

	// 检查用户是否已存在
	var existingUser models.User
	if err := database.DB.Where("email = ?", email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "该邮箱已被注册",
		})
		return
	}

	// 密码加密
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "密码加密失败",
		})
		return
	}
	// 随机一个avatar图像
	// 生成基于用户名的随机头像
	avatarURL := fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random", username)

	// 创建新用户
	newUser := models.User{
		Name:       username,
		Email:      email,
		Password:   hashedPassword,
		AgreeTerms: isAgreeTerms,
		Avatar:     avatarURL,
	}

	// 保存到数据库
	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "error",
			"message": "用户注册失败",
		})
		return
	}

	// 返回响应
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Login data received",
		"data": gin.H{
			"user_id": newUser.ID,
			"email":   newUser.Email,
		},
	})
}

func searchPostsHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		fmt.Println("未获取到用户信息")
	}
	// 获取页码参数，默认为第1页
	pageStr := c.Query("page")
	// 获取路由categories/:type中type参数
	qStr := c.Query("q")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 每页显示的帖子数量
	limit := 10
	offset := (page - 1) * limit

	// 查询总记录数
	var total int64
	database.DB.Model(&models.Post{}).Count(&total)

	// 查询当前页的帖子
	var posts []models.Post
	postQuery := database.DB.Order("created_at DESC").Offset(offset).Limit(limit)

	// 修改查询条件为模糊搜索
	postQuery = postQuery.Where("title LIKE ?", "%"+qStr+"%")

	postQuery.Find(&posts)

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// 创建带有友好时间的帖子结构
	type PostWithFriendlyTime struct {
		models.Post
		TimeAgo string
	}

	var postsWithTimeAgo []PostWithFriendlyTime
	for _, post := range posts {
		timeAgo := utils.GetTimeAgo(post.CreatedAt)
		postsWithTimeAgo = append(postsWithTimeAgo, PostWithFriendlyTime{
			Post:    post,
			TimeAgo: timeAgo,
		})
	}

	// 获取在线用户数
	var onlineCount int64
	cutoffTime := time.Now().Add(-30 * time.Minute)
	database.DB.Model(&models.UserOnlineStatus{}).
		Where("last_active_time > ?", cutoffTime).
		Count(&onlineCount)

	// 获取站点统计信息
	var userCount, postCount, commentCount int64
	database.DB.Model(&models.User{}).Count(&userCount)
	database.DB.Model(&models.Post{}).Count(&postCount)
	database.DB.Model(&models.Comment{}).Count(&commentCount)

	// 获取所有分类
	var categories []models.Category
	database.DB.Where("status_code = ?", 1).Find(&categories)

	// 1、统计注册用户
	// 2、统计文章数量
	// 3、统计回复评论数量
	// 4、数据表categories查询所有分类

	data := gin.H{
		"CurrentTime":  time.Now().Format("2006-01-02 15:04:05"),
		"posts":        postsWithTimeAgo,
		"currentPage":  page,
		"totalPages":   totalPages,
		"hasPrev":      page > 1,
		"hasNext":      page < totalPages,
		"prevPage":     page - 1,
		"nextPage":     page + 1,
		"user":         user,
		"userCount":    userCount,
		"postCount":    postCount,
		"commentCount": commentCount,
		"onlineCount":  onlineCount,
		"categories":   categories,
	}

	c.HTML(http.StatusOK, "search.tmpl", data)
}

func searchUsersHandler(c *gin.Context) {
	// 从上下文获取用户信息
	userObj, exists := c.Get("user")
	var user *models.User
	if exists && userObj != nil {
		user = userObj.(*models.User)
	} else {
		fmt.Println("未获取到用户信息")
	}

	// 获取页码参数和搜索关键字
	pageStr := c.Query("page")
	qStr := c.Query("q")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// 每页显示的用户数量
	limit := 10
	offset := (page - 1) * limit

	// 查询总记录数
	var total int64
	dbQuery := database.DB.Model(&models.User{})
	if qStr != "" {
		dbQuery = dbQuery.Where("name LIKE ? OR email LIKE ?", "%"+qStr+"%", "%"+qStr+"%")
	}
	dbQuery.Count(&total)

	// 查询当前页的用户
	var users []models.User
	userQuery := database.DB.Order("created_at DESC").Offset(offset).Limit(limit)
	if qStr != "" {
		userQuery = userQuery.Where("name LIKE ? OR email LIKE ?", "%"+qStr+"%", "%"+qStr+"%")
	}
	userQuery.Find(&users)

	// 计算总页数
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// 创建带有友好时间的用户结构
	type UserWithFriendlyTime struct {
		models.User
		TimeAgo string
	}

	var usersWithTimeAgo []UserWithFriendlyTime
	for _, user := range users {
		timeAgo := utils.GetTimeAgo(user.CreatedAt)
		usersWithTimeAgo = append(usersWithTimeAgo, UserWithFriendlyTime{
			User:    user,
			TimeAgo: timeAgo,
		})
	}

	data := gin.H{
		"users":         usersWithTimeAgo,
		"currentPage":   page,
		"totalPages":    totalPages,
		"hasPrev":       page > 1,
		"hasNext":       page < totalPages,
		"prevPage":      page - 1,
		"nextPage":      page + 1,
		"user":          user,
		"searchKeyword": qStr,
	}

	c.HTML(http.StatusOK, "member.tmpl", data)
}

func rssHandler(c *gin.Context) {
    // 查询最新的帖子
    var posts []models.Post
    database.DB.Order("created_at DESC").Limit(20).Find(&posts)

    // 获取当前域名
    scheme := "http"
    if c.Request.TLS != nil {
        scheme = "https"
    }
    currentDomain := scheme + "://" + c.Request.Host

    // 构造RSS数据结构
    type RSSItem struct {
        Title       string    `xml:"title"`
        Link        string    `xml:"link"`
        Description string    `xml:"description"`
        PubDate     time.Time `xml:"pubDate"`
        GUID        string    `xml:"guid"`
    }

    type RSSChannel struct {
        XMLName     xml.Name  `xml:"rss"`
        Version     string    `xml:"version,attr"`
        NSDC        string    `xml:"xmlns:dc,attr"`
        NSContent   string    `xml:"xmlns:content,attr"`
        NSAtom      string    `xml:"xmlns:atom,attr"`
        Channel     struct {
            Title       string    `xml:"title"`
            Link        string    `xml:"link"`
            Description string    `xml:"description"`
            Language    string    `xml:"language"`
            LastBuildDate string  `xml:"lastBuildDate"`
            Items       []RSSItem `xml:"item"`
        } `xml:"channel"`
    }

    rss := RSSChannel{
        Version:   "2.0",
        NSDC:      "http://purl.org/dc/elements/1.1/",
        NSContent: "http://purl.org/rss/1.0/modules/content/",
        NSAtom:    "http://www.w3.org/2005/Atom",
    }

    rss.Channel.Title = "Doniai技术社区"
    rss.Channel.Link = currentDomain
    rss.Channel.Description = "技术社区最新帖子"
    rss.Channel.Language = "zh-CN"
    rss.Channel.LastBuildDate = time.Now().Format(time.RFC1123Z)

    // 添加帖子项目
    for _, post := range posts {
        postLink := currentDomain + "/post-" + fmt.Sprintf("%d", post.ID) + "-1"
        rss.Channel.Items = append(rss.Channel.Items, RSSItem{
            Title:       post.Title,
            Link:        postLink,
            Description: fmt.Sprintf("<![CDATA[%s]]>", post.Content),
            PubDate:     post.CreatedAt,
            GUID:        postLink,
        })
    }

    c.XML(http.StatusOK, rss)
}



