package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default() // 获取 engine

	err := router.Run(":8080") // 指定端口，运行 Gin
	if err != nil {
		log.Panicln("服务器启动失败：", err.Error())
	}
}
