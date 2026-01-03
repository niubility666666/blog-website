package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"gin-doniai/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	once sync.Once
)

// GetInstance 获取数据库实例（单例模式）
func GetInstance() *gorm.DB {
	once.Do(func() {
		initDB()
	})
	return DB
}

// initDB 初始化数据库连接
func initDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("警告: 未能加载 .env 文件")
	}

	// 从环境变量获取数据库配置
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbname := os.Getenv("DB_NAME")

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		username, password, host, port, dbname)

	var dbErr error
	DB, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		log.Fatal("数据库连接失败:", dbErr)
	}

	// 配置连接池
	sqlDB, _ := DB.DB()
	sqlDB.SetMaxIdleConns(10)                // 空闲连接数
	sqlDB.SetMaxOpenConns(100)               // 最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour)      // 连接最大存活时间

	// 自动迁移（创建表）
	DB.AutoMigrate(&models.User{})
    DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Post{})
	DB.AutoMigrate(&models.Comment{})
	DB.AutoMigrate(&models.PostLike{})
	DB.AutoMigrate(&models.PostFavorite{})
	DB.AutoMigrate(&models.UserOnlineStatus{})
    DB.AutoMigrate(&models.PasswordReset{})
}

func InitDB() {
	GetInstance()
}
