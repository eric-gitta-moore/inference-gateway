package main

import (
	"log"
	"mime/multipart"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 定义枚举类型
type ModelTask string

const (
	FacialRecognition ModelTask = "facial-recognition"
	Search            ModelTask = "clip"
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

type PredictRequest struct {
	Image   *multipart.FileHeader `form:"image"`
	Text    string                `form:"text"`
	Entries PipelineRequest       `form:"entries"`
}

func handlePredictRequest(c *gin.Context) {
	var req PredictRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, req)
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.POST("/predict", handlePredictRequest)
	err := r.Run(":8080")
	if err != nil {
		log.Panicln("Server is running on port 8080")
	}
}
