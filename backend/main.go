package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// --- モデル定義 ---

type User struct {
	gorm.Model
	Email    string `gorm:"size:191;uniqueIndex;not null" json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type Task struct {
	gorm.Model
	Title  string `json:"title" binding:"required"`
	Status string `json:"status"`
	UserID uint   `json:"user_id"` // 誰のタスクか
}

// --- JWT設定 ---
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func main() {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("super-secret-key-change-this")
	}

	// --- DB接続 ---
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
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

	// マイグレーション
	if err := db.AutoMigrate(&User{}, &Task{}); err != nil {
		log.Fatal("テーブル作成失敗: ", err)
	}

	r := gin.Default()

	// --- CORS設定 ---
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"} // 本番ではFrontendのURLだけに絞るべき
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// --- 公開API (ログイン不要) ---
	r.POST("/signup", func(c *gin.Context) {
		var input User
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		input.Password = string(hashedPassword)
		if result := db.Create(&input); result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email already exists or DB error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
	})

	r.POST("/login", func(c *gin.Context) {
		var input User
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var user User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}

		// トークン生成 (UserIDを埋め込む)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": user.ID,
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"token": tokenString})
	})

	// --- 認証ミドルウェア ---
	authMiddleware := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		// トークンからUserIDを取り出す
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// JSONの数値はfloat64になるので変換が必要
			userID := uint(claims["sub"].(float64))
			c.Set("userID", userID) // コンテキストに保存して次の処理へ
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
		c.Next()
	}

	// --- 保護されたAPI (要ログイン) ---
	protected := r.Group("/")
	protected.Use(authMiddleware) // ここから下は全て検問を通る

	// タスク追加
	protected.POST("/tasks", func(c *gin.Context) {
		var newTask Task
		if err := c.ShouldBindJSON(&newTask); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		// ログイン中のユーザーIDを強制的にセット
		userID, _ := c.Get("userID")
		newTask.UserID = userID.(uint)

		if result := db.Create(&newTask); result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
			return
		}
		c.JSON(http.StatusCreated, newTask)
	})

	// タスク一覧（自分のだけ）
	protected.GET("/tasks", func(c *gin.Context) {
		var tasks []Task
		userID, _ := c.Get("userID")
		// WHERE user_id = ? を追加して他人のタスクを見せない
		db.Where("user_id = ?", userID).Find(&tasks)
		c.JSON(http.StatusOK, tasks)
	})

	// タスク削除（自分のだけ）
	protected.DELETE("/tasks/:id", func(c *gin.Context) {
		var task Task
		userID, _ := c.Get("userID")
		id := c.Param("id")

		// IDが一致し、かつUserIDも一致するものしか消せない
		if err := db.Where("id = ? AND user_id = ?", id, userID).First(&task).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found or permission denied"})
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
