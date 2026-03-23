package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AleksKAG/ai-construction-manager/internal/handlers"
	"github.com/AleksKAG/ai-construction-manager/internal/models"
	"github.com/AleksKAG/ai-construction-manager/internal/repository"
	"github.com/AleksKAG/ai-construction-manager/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v9"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil && os.Getenv("ENV") != "production" {
		log.Println("No .env file found")
	}

	// Логгер
	logger := logrus.New()
	if os.Getenv("ENV") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	}

	// Подключение к БД
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"), getEnv("DB_PORT", "5432"), getEnv("DB_USER", "app"),
		getEnv("DB_PASSWORD", "secret"), getEnv("DB_NAME", "construction_ai"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logger.Fatal("failed to connect database: ", err)
	}

	// Миграции
	db.AutoMigrate(&models.ProjectObject{}, &models.ProjectGraph{}, &models.GanttTask{})

	// Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})

	// Репозитории и сервисы
	projectRepo := repository.NewProjectRepository(db, redisClient)
	services.LoadSampleData(projectRepo) // Загрузка sample data в DB

	// Serve static files
r.Static("/static", "./web/static")

// HTML pages routing
r.LoadHTMLGlob("web/*.html")

// Page routes
r.GET("/", func(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "APIBase": getEnv("API_BASE", "/api/v1"),
    })
})
r.GET("/login", func(c *gin.Context) {
    c.HTML(http.StatusOK, "login.html", nil)
})
r.GET("/objects", func(c *gin.Context) {
    c.HTML(http.StatusOK, "objects.html", nil)
})
r.GET("/object/:id", func(c *gin.Context) {
    c.HTML(http.StatusOK, "object-detail.html", gin.H{
        "ObjectID": c.Param("id"),
    })
})
r.GET("/upload", func(c *gin.Context) {
    c.HTML(http.StatusOK, "upload.html", nil)
})
r.GET("/estimates/:id", func(c *gin.Context) {
    c.HTML(http.StatusOK, "estimates.html", gin.H{
        "ObjectID": c.Param("id"),
    })
})

	// Handlers
	r := gin.Default()
	r.Use(gin.Recovery())
	r.Use(AuthMiddleware()) // Auth middleware

	api := r.Group("/api/v1")
	{
		api.GET("/menu", handlers.MenuHandler)
		api.GET("/objects", handlers.ListObjects(projectRepo))
		api.POST("/objects", handlers.CreateObject(projectRepo))
		api.GET("/objects/:id", handlers.GetObject(projectRepo))
		api.PUT("/objects/:id", handlers.UpdateObject(projectRepo))
		api.DELETE("/objects/:id", handlers.DeleteObject(projectRepo))
		api.GET("/graphs", handlers.ListGraphs(projectRepo))
		api.GET("/graphs/:object_id", handlers.GetGraphForObject(projectRepo))
		api.POST("/upload", handlers.UploadFile(projectRepo))
		api.GET("/estimate/:id", handlers.GetEstimate(projectRepo))
	}

	// Telegram бот
	botToken := getEnv("TELEGRAM_BOT_TOKEN", "")
	if botToken != "" {
		bot, err := tgbotapi.NewBotAPI(botToken)
		if err != nil {
			logger.Warn("failed to init telegram bot: ", err)
		} else {
			// Интеграция с handlers (пример: бот использует API)
			u := tgbotapi.NewUpdate(0)
			u.Timeout = 60
			updates := bot.GetUpdatesChan(u)
			go func() {
				for update := range updates {
					if update.Message != nil {
						// Обработка команд (/menu, /objects и т.д.)
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Команда получена")
						bot.Send(msg)
					}
				}
			}()
			logger.Info("Telegram bot started")
		}
	}

	// Запуск сервера
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed: ", err)
		}
	}()

	logger.Infof("Server started on port %s", port)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: ", err)
	}

	logger.Info("Server exited gracefully")
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// AuthMiddleware - JWT auth
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization required"})
			c.Abort()
			return
		}
		tokenStr = tokenStr[len("Bearer "):]

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(getEnv("JWT_SECRET", "secret")), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		c.Next()
	}
}
