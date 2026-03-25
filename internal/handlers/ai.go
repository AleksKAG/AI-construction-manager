package handlers

import (
 "net/http"

 "github.com/gin-gonic/gin"
 "github.com/AleksKAG/ai-construction-manager/internal/service"
)

type AIHandler struct {
 service *service.AIService
}

func NewAIHandler() *AIHandler {
 return &AIHandler{
  service: service.NewAIService(),
 }
}

func (h *AIHandler) Analyze(c *gin.Context) {
 var req struct {
  Text string `json:"text"`
 }

 if err := c.ShouldBindJSON(&req); err != nil {
  c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
  return
 }

 result, err := h.service.Analyze(req.Text)
 if err != nil {
  c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
  return
 }

 c.JSON(http.StatusOK, gin.H{
  "result": result,
 })
}
