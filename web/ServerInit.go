package web

import (
	"bytes"
	"io"
	"log"
	"net/http"
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

	// 静态资源（Vue 构建产物）
	router.Static("/assets", "./assets/web/assets")             // JS/CSS 静态资源
	router.StaticFile("/", "./assets/web/index.html")           // 首页
	router.StaticFile("/index.html", "./assets/web/index.html") // 兼容

	router.Use(Cors())             // CORS
	router.Use(LoggerMiddleware()) // 日志中间件

	// API 路由
	apiGroup := router.Group("/")
	routerGroup := Group{apiGroup}
	routerGroup.Router()

	// ⭐ 关键：Vue Router history fallback（解决直接打开 /account 报 404）
	router.NoRoute(func(c *gin.Context) {
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
