package controller

import (
	"bufio"
	"fmt"
	"os"
	"time"
	"yatori-go-console/web/service"

	"github.com/gin-gonic/gin"
)

type UserApi struct{}

// 获取账号列表
func (UserApi) AccountListController(c *gin.Context) {
	//c.JSON(200, gin.H{})
	service.UserListService(c)
}

// 添加账号
func (UserApi) AddAccountController(c *gin.Context) {
	service.AddUserService(c)
}

// 删除账号
func (UserApi) DeleteAccountController(c *gin.Context) {
	service.DeleteUserService(c)
}

// 账号登录检测，用于检测账号密码是否正确
func (UserApi) AccountLoginCheckController(c *gin.Context) {
	service.AccountLoginCheckService(c)
}

// 获取账号配置信息
func (UserApi) GetAccountInformController(c *gin.Context) {
	service.GetAccountInformService(c)
}

// 获取课程列表
func (UserApi) AccountCourseListController(c *gin.Context) {
	service.AccountCourseListService(c)
}

// 登录账号
func (UserApi) LoginAccountController(c *gin.Context) {
	service.LoginUserService(c)
}

// 更新账号信息
func (UserApi) UpdateAccountController(c *gin.Context) {
	service.UpdateUserService(c)
}

// 启动刷课接口
func (UserApi) StartBrushController(c *gin.Context) {
	service.StartBrushService(c)
}

// 暂停刷课接口
func (UserApi) StopBrushController(c *gin.Context) {
	service.StopBrushService(c)
}

// 获取账号任务状态
func (UserApi) AccountStatusController(c *gin.Context) {
	service.AccountStatusService(c)
}

// 获取账号运行日志
func (UserApi) AccountLogController(c *gin.Context) {
	service.AccountLogService(c)
}

// 读取全局设置
func (UserApi) GetSettingController(c *gin.Context) {
	service.GetSettingService(c)
}

// 保存全局设置
func (UserApi) SaveSettingController(c *gin.Context) {
	service.SaveSettingService(c)
}

// 日志同步接口
func (UserApi) StreamLog(c *gin.Context) {
	logID := c.Param("id")

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	file, err := os.Open(fmt.Sprintf(`./assets/logs/%s.log`, logID))
	if err != nil {
		c.String(500, "error open log file")
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// 循环不断推送
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// SSE 协议格式
		c.Writer.Write([]byte("data: " + line + "\n\n"))
		c.Writer.Flush()
	}
}
