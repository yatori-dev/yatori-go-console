package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"yatori-go-console/dao"
	"yatori-go-console/global"
	"yatori-go-console/logic/progress"
	"yatori-go-console/logic/runner"
	"yatori-go-console/web/activity"

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
	progress.Init(dbInit)
	runner.SetWebMode(true)

	// 数据库里存在全局设置时覆盖 config.yaml（账号配置已迁移到数据库）
	if setting, err := dao.LoadSetting(dbInit); err == nil && setting != nil {
		global.Setting = *setting
	}
	activity.SetDefaultSetting(global.Setting)

	// 初始化服务器（可用 YATORI_WEB_ADDR 覆盖监听地址）
	addr := os.Getenv("YATORI_WEB_ADDR")
	if addr == "" {
		addr = ":8092"
	}
	initServer := serverInit()
	initServer.Run(addr)
}

// Group 封装 gin.RouterGroup
type Group struct {
	*gin.RouterGroup
}

// serverInit 初始化 Gin 服务
func serverInit() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.CustomRecovery(func(c *gin.Context, recovered any) {
		log.Printf("panic recovered: %v\n%s", recovered, string(debug.Stack()))
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": fmt.Sprintf("服务器内部错误: %v", recovered),
		})
	}))

	router.Use(Cors())

	// 1. API 路由
	apiGroup := router.Group("/api")
	apiGroup.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "API working"})
	})
	Group{apiGroup}.ApiV1Router()

	// 2. 前端静态资源（Next.js 导出目录）
	router.GET("/web/*filepath", func(c *gin.Context) {
		serveWeb(c, c.Param("filepath"))
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/web/")
	})

	// 3. 未匹配路由
	router.NoRoute(func(c *gin.Context) {
		requestPath := c.Request.URL.Path
		if strings.HasPrefix(requestPath, "/api") {
			c.JSON(404, gin.H{"error": "API endpoint not found", "path": requestPath})
			return
		}
		if strings.HasPrefix(requestPath, "/web") {
			serveWeb(c, strings.TrimPrefix(requestPath, "/web"))
			return
		}
		c.JSON(404, gin.H{"error": "Page not found", "path": requestPath})
	})

	return router
}

// webRootDir 返回前端静态资源目录。
// 优先使用 frontend/out（Next.js 本地构建产物），其次使用打包后的 assets/web。
func webRootDir() string {
	for _, dir := range []string{"./frontend/out", "./assets/web"} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				return dir
			}
		}
	}
	return "./assets/web"
}

// serveWeb 处理 Next.js 静态导出路由：
//
//	/web/accounts        -> accounts.html 或 accounts/index.html
//	/web/_next/xxx.js    -> _next/xxx.js
//	/web/                -> index.html
func serveWeb(c *gin.Context, requestPath string) {
	root := webRootDir()
	cleaned := path.Clean("/" + requestPath)
	if strings.Contains(cleaned, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path"})
		return
	}
	if cleaned == "/" {
		cleaned = "/index.html"
	}

	relative := filepath.FromSlash(strings.TrimPrefix(cleaned, "/"))
	candidates := []string{
		filepath.Join(root, relative),
		filepath.Join(root, relative, "index.html"),
		filepath.Join(root, relative+".html"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}
	}

	// 静态导出没有服务端路由，未命中时回退首页；保留旧版配置文件生成器页面兼容。
	if _, err := os.Stat(filepath.Join(root, "index.html")); err == nil {
		c.File(filepath.Join(root, "index.html"))
		return
	}
	if _, err := os.Stat(filepath.Join(root, "配置文件生成器.html")); err == nil {
		c.File(filepath.Join(root, "配置文件生成器.html"))
		return
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Frontend app not found", "path": requestPath})
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
	assetsPath := webRootDir()
	if _, err := os.Stat(assetsPath); os.IsNotExist(err) {
		return fmt.Errorf("静态资源目录不存在: %s，请先执行 frontend 构建", assetsPath)
	}
	for _, file := range []string{"index.html", "_next/static"} {
		if _, err := os.Stat(filepath.Join(assetsPath, file)); os.IsNotExist(err) {
			return fmt.Errorf("必要的静态资源文件不存在: %s", filepath.Join(assetsPath, file))
		}
	}
	return nil
}
