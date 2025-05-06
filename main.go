package main

import (
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"resty.dev/v3"
)

// 定义枚举类型
type ModelTask string

const (
	FacialRecognition ModelTask = "facial-recognition"
	Search            ModelTask = "clip"
	OCRSearch         ModelTask = "ocr-search"
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

type OCRSearchEntry struct {
	Recognition PipelineEntry `json:"recognition,omitempty"`
	Textual     PipelineEntry `json:"textual,omitempty"`
}

// EntryRequest 结构体
type EntryRequest struct {
	OCR *OCRSearchEntry `json:"ocr,omitempty"`
}

type PredictRequest struct {
	Image   *multipart.FileHeader `form:"image,omitempty"`
	Text    string                `form:"text,omitempty"`
	Entries EntryRequest          `form:"entries"`
}

const Token = "mt_photos_ai_extra"

// 鉴权中间件
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 x-auth-token
		token := c.GetHeader("x-auth-token")

		if token != Token {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的认证令牌",
			})
			c.Abort()
			return
		}

		// 验证通过，继续处理请求
		c.Next()
	}
}

// 处理人脸识别任务
func handleFacialRecognition(c *gin.Context, req PredictRequest) {
	// TODO: 实现人脸识别的具体逻辑
	c.JSON(http.StatusOK, gin.H{
		"task":    "facial-recognition",
		"message": "处理人脸识别任务",
	})
}

// 处理搜索任务
func handleSearch(c *gin.Context, req PredictRequest) {
	// TODO: 实现搜索的具体逻辑
	c.JSON(http.StatusOK, gin.H{
		"task":    "search",
		"message": "处理搜索任务",
	})
}

// OCR 响应相关的结构体
type OCRBox struct {
	X      string `json:"x"`
	Y      string `json:"y"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

type OCRResult struct {
	Texts  []string `json:"texts"`
	Scores []string `json:"scores"`
	Boxes  []OCRBox `json:"boxes"`
}

type OCRResponse struct {
	Result OCRResult `json:"result"`
}

// 处理 OCR 搜索任务
func handleOCRSearch(c *gin.Context, req PredictRequest) {
	// 创建 resty 客户端
	client := resty.New()

	file, err := req.Image.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "读取图片失败：" + err.Error(),
		})
		return
	}
	defer file.Close()


	var ocrResp OCRResponse
	// 使用 resty 发送请求
	resp, err := client.R().
		SetDebug(true).
		SetHeader("api-key", "mt_photos_ai_extra").
		SetFileReader("file", req.Image.Filename, file).
		SetResult(&ocrResp).
		Post("http://localhost:8060/ocr/rec")

	if err != nil || resp.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "转发请求失败：" + resp.String(),
		})
		return
	}

	// 返回结构化的响应
	c.JSON(resp.StatusCode(), gin.H{
		"ocr":    strings.Join(ocrResp.Result.Texts, " "),
		"result": ocrResp.Result,
	})
}

func handleOtherTasks(c *gin.Context, req PredictRequest) {
	c.JSON(http.StatusOK, gin.H{
		"task":    "other",
		"message": "处理其他任务",
	})
}

func handlePredictRequest(c *gin.Context) {
	var req PredictRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if req.Entries.OCR != nil {
		handleOCRSearch(c, req)
		return
	}

	handleOtherTasks(c, req)

	// // 遍历请求中的任务类型
	// for task, _ := range req.Entries {
	// 	switch task {
	// 	case FacialRecognition:
	// 		handleFacialRecognition(c, req)
	// 	case Search:
	// 		handleSearch(c, req)
	// 	case OCRSearch:
	// 		handleOCRSearch(c, req)
	// 	default:
	// 		c.JSON(http.StatusBadRequest, gin.H{
	// 			"error": "不支持的任务类型",
	// 		})
	// 		return
	// 	}
	// }
}

func main() {
	r := gin.Default()

	// 不需要验证的路由组
	noAuth := r.Group("/")
	{
		noAuth.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "pong",
			})
		})
	}

	// 需要验证的路由组
	auth := r.Group("/")
	auth.Use(authMiddleware())
	{
		auth.POST("/predict", handlePredictRequest)
	}

	err := r.Run(":8080")
	if err != nil {
		log.Panicln("Server is running on port 8080")
	}
}
