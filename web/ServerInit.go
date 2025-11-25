package bootstrap

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"yatori-go-server/dao"
	"yatori-go-server/global"

	"github.com/gin-gonic/gin"
)

// 统一初始化
func Init() {
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
	router.Static("/assets", "./assets/web/assets") //静态资源
	router.Use(Cors())                              //设置跨域
	router.Use(LoggerMiddleware())
	apiGroup := router.Group("/")
	routerGroup := Group{apiGroup}
	routerGroup.Router()
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
