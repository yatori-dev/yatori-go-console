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

	// 2️⃣ 前端静态资源 + 页面路由：用 NoRoute 处理（避免与 /api 段冲突）
	// - apiGroup 已注册 /api/* 路由，会优先匹配
	// - 其他 GET 路径走 NoRoute，返回静态资源 / page.html / index.html
	// - /web/*filepath 兼容老路径，重定向到根
	router.GET("/web/*filepath", func(c *gin.Context) {
		filepathParam := c.Param("filepath")
		c.Redirect(301, filepathParam)
	})

	// 3️⃣ 处理其他未匹配的路由（SPA fallback）
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 请求返回 404
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{
				"error": "API endpoint not found",
				"path":  path,
			})
			return
		}

		// 非 GET 请求（如 POST/PUT 到前端路径）返回 405
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			c.JSON(405, gin.H{"error": "Method not allowed"})
			return
		}

		// 根路径返回 index.html
		if path == "/" || path == "" {
			indexPath := "./assets/web/index.html"
			if _, err := os.Stat(indexPath); err == nil {
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.File(indexPath)
			} else {
				c.JSON(500, gin.H{"error": "index.html not found"})
			}
			return
		}

		// 静态资源（通过扩展名判断）
		ext := filepath.Ext(path)
		if ext != "" && ext != ".html" {
			staticPath := filepath.Join("./assets/web", path[1:]) // 去掉开头的 "/"
			if _, err := os.Stat(staticPath); err == nil {
				// 对 JS 文件设置无缓存头，防止浏览器缓存旧 chunk
				if ext == ".js" {
					c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
					c.Header("Pragma", "no-cache")
					c.Header("Expires", "0")
				}
				c.File(staticPath)
				return
			}
			c.JSON(404, gin.H{
				"error": "Static resource not found",
				"path":  path,
			})
			return
		}

		// 路由请求：尝试 page.html，否则 fallback 到 index.html（SPA）
		cleanPath := strings.TrimRight(path, "/")
		pagePath := "./assets/web" + cleanPath + ".html"
		if _, err := os.Stat(pagePath); err == nil {
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.File(pagePath)
			return
		}

		// 不存在的 page：fallback 到 index.html 让前端路由处理
		indexPath := "./assets/web/index.html"
		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			c.JSON(500, gin.H{"error": "Main index.html file not found"})
			return
		}

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.File(indexPath)
	})

	return router
}

// Cors 跨域中间件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

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
