package service

import (
	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/vo"
	"yatori-go-console/global"
	"yatori-go-console/web/activity"

	"github.com/gin-gonic/gin"
)

// GetSettingService 读取全局设置（AI、邮箱、题库等）。
func GetSettingService(c *gin.Context) {
	ok(c, "读取设置成功", gin.H{"setting": global.Setting})
}

// SaveSettingService 保存全局设置到 SQLite，并在内存中立即生效。
func SaveSettingService(c *gin.Context) {
	var req config.Setting
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "设置参数错误: "+err.Error())
		return
	}
	if err := dao.SaveSetting(global.GlobalDB, req); err != nil {
		fail(c, "保存设置失败: "+err.Error())
		return
	}
	global.Setting = req
	activity.SetDefaultSetting(req)
	c.JSON(200, vo.Response{Code: 200, Message: "设置已保存", Data: gin.H{"setting": req}})
}
