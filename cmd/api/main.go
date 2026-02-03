package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stroy-assistent/internal/delivery/http"
	tgbot "stroy-assistent/internal/delivery/telegram"
	"stroy-assistent/internal/repository/postgres"
	"stroy-assistent/internal/service"
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
	dsn := "host=" + getEnv("DB_HOST", "localhost") +
		" port=" + getEnv("DB_PORT", "5432") +
		" user=" + getEnv("DB_USER", "app") +
		" password=" + getEnv("DB_PASSWORD", "secret") +
		" dbname=" + getEnv("DB_NAME", "stroy_assistent") +
		" sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect database: ", err)
	}

	// Миграции (в реальном проекте — через goose/pg-migrate)
	// db.AutoMigrate(&domain.Project{}, &domain.Task{}, ...)

	// Инициализация сервисов
	projectHandler := http.NewProjectHandler(db)

	// Gin роутер
	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api/v1")
	projectHandler.RegisterRoutes(api)

	// Telegram бот (если токен задан)
	if botToken := os.Getenv("TELEGRAM_BOT_TOKEN"); botToken != "" {
		adminID := getIntEnv("TELEGRAM_ADMIN_ID", 0)
		bot, err := tgbot.NewBot(botToken, postgres.NewProjectRepository(db), adminID)
		if err != nil {
			logger.Warn("failed to init telegram bot: ", err)
		} else {
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				if err := bot.Start(ctx); err != nil {
					logger.Error("telegram bot stopped: ", err)
				}
			}()
			defer cancel()
			logger.Info("Telegram bot started")
		}
	}

	// Запуск сервера
	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed: ", err)
		}
	}()

	logger.Infof("Server started on port %s", port)

	// Ожидание сигнала завершения
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
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getIntEnv(key string, fallback int64) int64 {
	if value := os.Getenv(key); value != "" {
		var n int64
		if _, err := fmt.Sscanf(value, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}
