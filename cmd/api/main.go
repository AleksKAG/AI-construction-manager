package main

import (
 "context"
 "net/http"
 "os"
 "os/signal"
 "syscall"
 "time"

 "github.com/AleksKAG/ai-construction-manager/internal/handlers"
 "github.com/AleksKAG/ai-construction-manager/internal/models"
 "github.com/AleksKAG/ai-construction-manager/internal/repository"
 "github.com/gin-gonic/gin"

 "github.com/joho/godotenv"
 "github.com/sirupsen/logrus"
 "gorm.io/driver/sqlite"
 "gorm.io/gorm"
)

func main() {
 _ = godotenv.Load()

 logger := logrus.New()

 // путь через ENV (важно для Docker)
 dbPath := getEnv("DB_PATH", "data/app.db")

 if err := os.MkdirAll("data", 0755); err != nil {
  logger.Fatal(err)
 }

 db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
 if err != nil {
  logger.Fatal(err)
 }

 err = db.AutoMigrate(
  &models.ProjectObject{},
  &models.ProjectGraph{},
  &models.GanttTask{},
 )
 if err != nil {
  logger.Fatal(err)
 }

 repo := repository.NewProjectRepository(db)

 r := gin.Default()

 // static
 r.Static("/static", "./web/static")
 r.LoadHTMLGlob("web/*.html")

 // healthcheck
 r.GET("/health", func(c *gin.Context) {
  c.JSON(200, gin.H{"status": "ok"})
 })

 // HTML
 r.GET("/", func(c *gin.Context) {
  c.HTML(http.StatusOK, "index.html", nil)
 })

 r.GET("/objects", func(c *gin.Context) {
  c.HTML(http.StatusOK, "objects.html", nil)
 })

  api := r.Group("/api/v1")
 {
  api.GET("/menu", handlers.MenuHandler)
  api.GET("/objects", handlers.ListObjects(repo))
  api.POST("/objects", handlers.CreateObject(repo))
  api.GET("/objects/:id", handlers.GetObject(repo))
  api.PUT("/objects/:id", handlers.UpdateObject(repo))
  api.DELETE("/objects/:id", handlers.DeleteObject(repo))
  
  api.POST("/ai/analyze", handlers.NewAIHandler().Analyze)
 }

 port := getEnv("PORT", "8080")

 srv := &http.Server{
  Addr: ":" + port,
  Handler: r,
 }

 go func() {
  logger.Infof("Server started on :%s", port)
  if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
   logger.Fatal(err)
  }
 }()

 quit := make(chan os.Signal, 1)
 signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
 <-quit

 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
 defer cancel()

 srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
 if v := os.Getenv(key); v != "" {
  return v
 }
 return fallback
}
