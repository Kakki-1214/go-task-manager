package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// --- モデル定義 ---

// User モデル: ユーザー情報
type User struct {
	gorm.Model
	Email    string `gorm:"size:191;uniqueIndex;not null" json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"` // DBにはハッシュ化して保存
}

// Task モデル: タスク情報（UserIDを追加）
type Task struct {
	gorm.Model
	Title  string `json:"title" binding:"required"`
	Status string `json:"status"`
	UserID uint   `json:"user_id"` // 誰のタスクか識別するID
}

// --- JWT設定 ---
// 本番では必ず複雑なランダム文字列を環境変数に入れること
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func main() {
	// JWT_SECRETがない場合の安全策（開発用）
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("super-secret-key-change-this")
	}

	// --- DB接続設定 ---
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

	// マイグレーション（Userテーブル作成、Taskテーブル更新）
	// db.AutoMigrate(&User{}, &Task{})
	if err := db.AutoMigrate(&User{}, &Task{}); err != nil {
		log.Fatal("テーブル作成に失敗しました: ", err)
	}

	r := gin.Default()

	// --- CORS設定 ---
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"} // Authorizationヘッダーを許可
	r.Use(cors.New(config))

	// --- 認証API ---

	// 1. ユーザー登録 (Sign Up)
	r.POST("/signup", func(c *gin.Context) {
		var input User
		// バリデーション
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// パスワードのハッシュ化
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		input.Password = string(hashedPassword)

		// DB保存
		if result := db.Create(&input); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	})

	// 2. ログイン (Login) -> JWT発行
	r.POST("/login", func(c *gin.Context) {
		var input User
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// ユーザー検索
		var user User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// パスワード照合
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// JWTトークン生成
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,                               // ユーザーID
			"exp": time.Now().Add(time.Hour * 24).Unix(), // 有効期限: 24時間
		})

		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// トークンを返す
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	// --- タスクAPI (まだ認証ガードなし。次回実装) ---
	r.POST("/tasks", func(c *gin.Context) {
		var newTask Task
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// ※ 本来はここで「ログイン中のユーザーID」を入れる
		result := db.Create(&newTask)
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusCreated, newTask)
	})

	r.GET("/tasks", func(c *gin.Context) {
		var tasks []Task
		db.Find(&tasks) // ※ 本来は「自分のタスクだけ」を取得する
		c.JSON(http.StatusOK, tasks)
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
