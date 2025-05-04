package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义枚举类型
type ModelTask string

const (
	FacialRecognition ModelTask = "facial-recognition"
	Search            ModelTask = "search"
)

type ModelType string

const (
	Detection   ModelType = "detection"
	Recognition ModelType = "recognition"
	Textual     ModelType = "textual"
	Visual      ModelType = "visual"
)

// PipelineEntry 结构体
type PipelineEntry struct {
	ModelName string                 `json:"modelName"`
	Options   map[string]interface{} `json:"options"`
}

// PipelineRequest 结构体
type PipelineRequest map[ModelTask]map[ModelType]PipelineEntry

func handlePipelineRequest(c *gin.Context) {
	var request PipelineRequest

	// 使用 ShouldBindJSON 来解析请求体
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// 处理请求...
	// 例如：
	for task, types := range request {
		for modelType, entry := range types {
			// 处理每个模型配置
			log.Printf("Task: %s, Type: %s, Model: %s\n",
				task, modelType, entry.ModelName)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/pipeline", handlePipelineRequest)
	err := r.Run(":8080")
	if err != nil {
		log.Panicln("Server is running on port 8080")
	}
}
