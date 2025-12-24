package web

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"
	"yatori-go-console/dao"
	"yatori-go-console/global"

	"github.com/gin-gonic/gin"
)

// 统一初始化
func ServiceInit() {
	//初始化数据库
	dbInit, err := dao.SqliteInit()
	if err != nil {
		panic(err)
	}
	global.GlobalDB = dbInit

	//初始化服务器
	initServer := serverInit()
	initServer.Run(":8080")
}

type Group struct {
	*gin.RouterGroup
}

// 初始化gin
func serverInit() *gin.Engine {
	router := gin.Default()
	router.Use(Cors())             // CORS
	router.Use(LoggerMiddleware()) // 日志中间件

	// 1️⃣ API 路由（必须最先）
	apiGroup := router.Group("/api")
	apiRouterGroup := Group{apiGroup}
	apiRouterGroup.ApiV1Router()

	// 2️⃣ Next.js 核心静态资源（必须显式暴露）
	router.StaticFS("/_next", http.Dir("./assets/web/_next"))

	// 3️⃣ 其他静态资源（你原来的 /web 保留）
	router.StaticFS("/web", http.Dir("./assets/web"))

	// 4️⃣ 首页
	router.GET("/", func(c *gin.Context) {
		c.File("./assets/web/index.html")
	})

	// 5️⃣ SPA fallback（⚠️ 关键：放行 _next）
	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/_next/") {
			c.Status(404)
			return
		}
		c.File("./assets/web/index.html")
	})

	return router
}

// 跨域组件
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*") // 可替换为具体域名
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

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		bodyBytes, _ := io.ReadAll(c.Request.Body)
		// 重新赋值 Body，确保后续可读
		c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		// 打印原始 JSON
		log.Printf("Request Body: %s", string(bodyBytes))

		c.Next()
	}
}
