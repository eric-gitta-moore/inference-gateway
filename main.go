package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
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

type ModelParams struct {
	ModelName string                 `json:"modelName"`
	Options   map[string]interface{} `json:"options"`
}

type PipelineEntry struct {
	Detection   *ModelParams `json:"detection,omitempty"`
	Recognition *ModelParams `json:"recognition,omitempty"`
	Textual     *ModelParams `json:"textual,omitempty"`
	Visual      *ModelParams `json:"visual,omitempty"`
}

// PipelineRequest 结构体
type PipelineRequest struct {
	OCR             *PipelineEntry `json:"ocr,omitempty"`
	CLIP            *PipelineEntry `json:"clip,omitempty"`
	FaceRecognition *PipelineEntry `json:"face-recognition,omitempty"`
}

type PredictRequest struct {
	Image   *multipart.FileHeader `form:"image,omitempty"`
	Text    *string               `form:"text,omitempty"`
	Entries PipelineRequest       `form:"entries"`
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

// OCR 响应相关的结构体
type OCRBox struct {
	X      string `json:"x"`
	Y      string `json:"y"`
	Width  string `json:"width"`
	Height string `json:"height"`
}

// OCR 响应相关的结构体
type OCRBoxResponse struct {
	X1 float64 `json:"x1"`
	Y1 float64 `json:"y1"`
	X2 float64 `json:"x2"`
	Y2 float64 `json:"y2"`
}

type OCRResult struct {
	Texts  []string `json:"texts"`
	Scores []string `json:"scores"`
	Boxes  []OCRBox `json:"boxes"`
}

type OCRResponse struct {
	Result OCRResult `json:"result"`
}

// 创建 resty 客户端
var httpUtil = resty.New().SetDebug(true)

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

var (
	IMMICH_API        = getEnvOrDefault("IMMICH_API", "http://localhost:3003")
	MT_PHOTOS_API     = getEnvOrDefault("MT_PHOTOS_API", "http://localhost:8060")
	MT_PHOTOS_API_KEY = getEnvOrDefault("MT_PHOTOS_API_KEY", "mt_photos_ai_extra")
	PORT              = getEnvOrDefault("PORT", "8080")
)

// 处理 OCR 搜索任务
func handleOCRSearch(c *gin.Context, req PredictRequest) {
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
	resp, err := httpUtil.R().
		SetHeader("api-key", MT_PHOTOS_API_KEY).
		SetFileReader("file", req.Image.Filename, file).
		SetResult(&ocrResp).
		Post(MT_PHOTOS_API + "/ocr/rec")

	if err != nil || resp.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "转发请求失败：" + resp.String(),
		})
		return
	}

	// 返回结构化的响应
	c.JSON(resp.StatusCode(), gin.H{
		"ocr": strings.Join(ocrResp.Result.Texts, " "),
		"result": gin.H{
			"texts":  ocrResp.Result.Texts,
			"scores": ocrResp.Result.Scores,
			"boxes": lo.Map(ocrResp.Result.Boxes, func(box OCRBox, idx int) OCRBoxResponse {
				var boxResp OCRBoxResponse
				var err error
				if boxResp.X1, err = strconv.ParseFloat(box.X, 64); err != nil {
					return boxResp
				}
				if boxResp.Y1, err = strconv.ParseFloat(box.Y, 64); err != nil {
					return boxResp
				}
				if boxResp.X2, err = strconv.ParseFloat(box.Width, 64); err != nil {
					return boxResp
				}
				if boxResp.Y2, err = strconv.ParseFloat(box.Height, 64); err != nil {
					return boxResp
				}
				return boxResp
			}),
		},
	})
}

type CLIPResponse struct {
	Result []string `json:"result"`
}

func handleCLIPSearch(c *gin.Context, req PredictRequest) {
	task := *req.Entries.CLIP
	var resp *resty.Response
	var err error
	var clipResp CLIPResponse
	reqInstance := httpUtil.R().
		SetHeader("api-key", MT_PHOTOS_API_KEY)
	if task.Textual != nil {
		if req.Text == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "文本任务需要提供文本",
			})
			return
		}
		text := *req.Text
		resp, err = reqInstance.
			SetBody(gin.H{
				"text": text,
			}).
			SetResult(&clipResp).
			Post(MT_PHOTOS_API + "/clip/txt")
	} else {
		if task.Visual == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "不支持的 CLIP 任务类型",
			})
			return
		}
		if req.Image == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "图片任务需要提供图片",
			})
			return
		}
		file, err := req.Image.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "读取图片失败：" + err.Error(),
			})
			return
		}
		defer file.Close()
		resp, err = reqInstance.
			SetFileReader("file", req.Image.Filename, file).
			SetResult(&clipResp).
			Post(MT_PHOTOS_API + "/clip/img")
	}

	if err != nil || resp.StatusCode() != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "转发请求失败：" + resp.String(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"clip": fmt.Sprintf("[%v]", strings.Join(clipResp.Result, ",")),
	})
}

func handleImmichML(c *gin.Context, req PredictRequest) {
	target, _ := url.Parse(IMMICH_API)
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ServeHTTP(c.Writer, c.Request)
}

func handlePredictRequest(c *gin.Context) {
	// 先保存请求体
	var bytes bytes.Buffer
	c.Request.Body = io.NopCloser(io.TeeReader(c.Request.Body, &bytes))

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
	if req.Entries.CLIP != nil {
		handleCLIPSearch(c, req)
		return
	}

	// 重新设置请求体，确保后续处理可以使用
	c.Request.Body = io.NopCloser(&bytes)
	handleImmichML(c, req)
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/predict", handlePredictRequest)

	// 在启动服务器前打印日志
	log.Printf("Server is starting on port %s", PORT)
	err := r.Run(":" + PORT)
	if err != nil {
		log.Panicln("Server failed to start:", err)
	}
}
