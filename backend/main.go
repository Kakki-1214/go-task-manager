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

// --- 1. モデル定義 (DBの設計図) ---
type Task struct {
	gorm.Model        // ID, CreatedAt, UpdatedAtなどを自動管理
	Title      string `json:"title" binding:"required"` // 必須入力
	Status     string `json:"status"`                   // "pending", "done" など
}

func main() {
	// --- 2. DB接続設定 ---
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

	// --- 3. マイグレーション (テーブル自動生成) ---
	// Task構造体を見て、自動的にDBにテーブルを作ってくれる魔法の機能
	db.AutoMigrate(&Task{})

	r := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"Content-Type"}
	r.Use(cors.New(config))

	// --- 4. ルーティングと処理 (APIの実装) ---

	// [POST] タスク作成
	r.POST("/tasks", func(c *gin.Context) {
		var newTask Task
		// JSONを受け取って構造体にマッピング
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// DBに保存
		result := db.Create(&newTask)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusCreated, newTask)
	})

	// [GET] タスク全取得
	r.GET("/tasks", func(c *gin.Context) {
		var tasks []Task
		db.Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	// [GET] タスク詳細取得
	r.GET("/tasks/:id", func(c *gin.Context) {
		var task Task
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		c.JSON(http.StatusOK, task)
	})

	// [PUT] タスク更新
	r.PUT("/tasks/:id", func(c *gin.Context) {
		var task Task
		// まず存在確認
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		// 新しいデータを読み込み
		var updateData Task
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// 更新実行
		db.Model(&task).Updates(updateData)
		c.JSON(http.StatusOK, task)
	})

	// [DELETE] タスク削除
	r.DELETE("/tasks/:id", func(c *gin.Context) {
		var task Task
		if err := db.First(&task, c.Param("id")).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		db.Delete(&task)
		c.JSON(http.StatusOK, gin.H{"message": "Deleted successfully"})
	})

	// r.Run(":8080")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // ローカル開発用
	}
	r.Run(":" + port)
}
