package web

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"yatori-go-console/dao"
	"yatori-go-console/global"

	"github.com/gin-gonic/gin"
)

// ServiceInit 统一初始化
func ServiceInit() {
	// 初始化数据库
	dbInit, err := dao.SqliteInit()
	if err != nil {
		panic(err)
	}
	global.GlobalDB = dbInit

	// 初始化服务器
	initServer := serverInit()
	initServer.Run(":8080")
}

// Group 封装 gin.RouterGroup
type Group struct {
	*gin.RouterGroup
}

// serverInit 初始化 Gin 服务
func serverInit() *gin.Engine {
	router := gin.Default()

	router.Use(Cors())
	router.Use(LoggerMiddleware())

	// 1️⃣ API 路由 - 放在最前面以确保优先匹配
	apiGroup := router.Group("/api")
	apiGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API working"})
	})
	Group{apiGroup}.ApiV1Router()

	// 2️⃣ 单一路由处理 - 避免冲突
	router.GET("/web/*filepath", func(c *gin.Context) {
		filepathParam := c.Param("filepath")

		// 如果是根路径，返回 index.html
		if filepathParam == "/" || filepathParam == "" {
			indexPath := "./assets/web/index.html"
			if _, err := os.Stat(indexPath); err == nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.File(indexPath)
			} else {
				c.JSON(500, gin.H{"error": "index.html not found"})
			}
			return
		}

		// 检查是否是静态资源（通过扩展名判断）
		ext := filepath.Ext(filepathParam)
		if ext != "" && ext != ".html" {
			// 尝试查找静态文件
			staticPath := filepath.Join("./assets/web", filepathParam[1:]) // 去掉开头的 "/"
			if _, err := os.Stat(staticPath); err == nil {
				c.File(staticPath)
				return
			}
			// 如果静态文件不存在，返回 404
			c.JSON(404, gin.H{
				"error": "Static resource not found",
				"path":  filepathParam,
			})
			return
		}

		// 对于路由请求，返回 index.html 以启用前端路由
		indexPath := "./assets/web/index.html"
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			c.JSON(500, gin.H{"error": "Main index.html file not found"})
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(indexPath)
	})

	// 3️⃣ 处理其他未匹配的路由
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是 API 请求，返回 API 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{
				"error": "API endpoint not found",
				"path":  path,
			})
			return
		}

		// 如果是 /web 路径，但上面的路由没有匹配，返回 index.html
		if strings.HasPrefix(path, "/web") {
			indexPath := "./assets/web/index.html"
			if _, err := os.Stat(indexPath); err == nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.File(indexPath)
			} else {
				c.JSON(404, gin.H{"error": "Frontend app not found"})
			}
			return
		}

		// 其他路径返回 404
		c.JSON(404, gin.H{
			"error": "Page not found",
			"path":  path,
		})
	})

	return router
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// checkAssetsDir 检查静态资源目录是否存在
func checkAssetsDir() error {
	assetsPath := "./assets/web"
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return fmt.Errorf("静态资源目录不存在: %s，请确保 Next.js 项目已构建并输出到该目录", assetsPath)
	}

	// 检查必要的文件是否存在
	requiredFiles := []string{
		"index.html",
		"_next/static",
	}

	for _, file := range requiredFiles {
		fullPath := filepath.Join(assetsPath, file)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return fmt.Errorf("必要的静态资源文件不存在: %s", fullPath)
		}
	}

	return nil
}

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		log.Printf("Request Body: %s", string(bodyBytes))
		c.Next()
	}
}
