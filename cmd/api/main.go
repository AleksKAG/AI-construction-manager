package main

import (
	"context"
	
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
	
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Загрузка .env
	if err := godotenv.Load(); err != nil && os.Getenv("ENV") != "production" {
		log.Println("No .env file found")
	}

	logger := logrus.New()
	if os.Getenv("ENV") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{})
	}

	// ====================== SQLite ======================
	dbPath := "data/app.db"
	if err := os.MkdirAll("data", 0755); err != nil {
		logger.Fatal("failed to create data directory: ", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Fatal("failed to connect to sqlite: ", err)
	}

	// Миграции
	err = db.AutoMigrate(
		&models.ProjectObject{},
		&models.ProjectGraph{},
		&models.GanttTask{},
	)
	if err != nil {
		logger.Fatal("failed to migrate database: ", err)
	}
	logger.Infof("SQLite connected: %s", dbPath)

	// Репозиторий
	projectRepo := repository.NewProjectRepository(db)
	services.LoadSampleData(projectRepo)

	// ====================== Gin ======================
	r := gin.Default()
	r.Use(gin.Recovery())

	// Статические файлы и шаблоны
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/*.html")

	// HTML страницы
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/objects", func(c *gin.Context) { c.HTML(http.StatusOK, "objects.html", nil) })
	r.GET("/object/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "object-detail.html", gin.H{"ObjectID": c.Param("id")})
	})
	r.GET("/upload", func(c *gin.Context) { c.HTML(http.StatusOK, "upload.html", nil) })

	// API
	api := r.Group("/api/v1")
	{
		api.GET("/menu", handlers.MenuHandler)
		api.GET("/objects", handlers.ListObjects(projectRepo))
		api.POST("/objects", handlers.CreateObject(projectRepo))
		api.GET("/objects/:id", handlers.GetObject(projectRepo))
		api.PUT("/objects/:id", handlers.UpdateObject(projectRepo))
		api.DELETE("/objects/:id", handlers.DeleteObject(projectRepo))
		api.POST("/upload", handlers.UploadFile(projectRepo))
	}

	port := getEnv("PORT", "8080")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		logger.Infof("Server started on http://0.0.0.0:%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	logger.Info("Server exited")
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
