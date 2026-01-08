package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// --- モデル定義 ---
type Task struct {
	gorm.Model
	Title  string `json:"title" binding:"required"`
	Status string `json:"status"`
}

func main() {
	// --- DB接続設定 ---
	// Render等の環境変数(DB_DSN)があればそれを優先、なければローカル用を作成
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)
	}

	var db *gorm.DB
	var err error
	count := 0

	// リトライ接続処理
	for {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			fmt.Println("DB接続成功！")
			break
		}
		fmt.Printf("DB接続待機中 (%d)... %v\n", count+1, err)
		count++
		if count > 30 {
			log.Fatal("DB接続タイムアウト: ", err)
		}
		time.Sleep(3 * time.Second)
	}

	// マイグレーション
	db.AutoMigrate(&Task{})

	r := gin.Default()

	// --- CORS設定 ---
	// Vercel等のフロントエンドからのアクセスを許可
	config := cors.DefaultConfig()
	// 全許可（開発・デモ用としてはこれでOK）
	config.AllowAllOrigins = true
	// もし厳密にやるなら以下のように指定
	// config.AllowOrigins = []string{"http://localhost:3000", "https://task-front.vercel.app"}
	r.Use(cors.New(config))

	// --- API実装 ---
	r.POST("/tasks", func(c *gin.Context) {
		var newTask Task
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result := db.Create(&newTask)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusCreated, newTask)
	})

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []Task
		db.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	r.GET("/tasks/:id", func(c *gin.Context) {
		var task Task
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	r.PUT("/tasks/:id", func(c *gin.Context) {
		var task Task
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		var updateData Task
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		db.Model(&task).Updates(updateData)
		c.JSON(http.StatusOK, task)
	})

	r.DELETE("/tasks/:id", func(c *gin.Context) {
		var task Task
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		db.Delete(&task)
		c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
	})

	// --- ポート設定 (Render対応) ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
