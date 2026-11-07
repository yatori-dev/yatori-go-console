package web

import "yatori-go-console/web/controller"

func (router Group) StaticRouter() {
	// 静态资源在 ServerInit.go 中统一处理
}

func (router Group) ApiV1Router() {
	var userApi controller.UserApi
	router.GET("/v1/accountList", userApi.AccountListController)                      // 拉取账号列表
	router.POST("/v1/addAccount", userApi.AddAccountController)                       // 添加账号
	router.POST("/v1/deleteAccount", userApi.DeleteAccountController)                 // 删除账号
	router.POST("/v1/updateAccount", userApi.UpdateAccountController)                 // 修改账号信息
	router.POST("/v1/accountLoginCheck", userApi.AccountLoginCheckController)         // 账号登录检测
	router.GET("/v1/getAccountInformForUid/:uid", userApi.GetAccountInformController) // 拉取账号配置
	router.GET("/v1/getAccountCourseList/:uid", userApi.AccountCourseListController)  // 获取课程列表
	router.GET("/v1/accountStatus/:uid", userApi.AccountStatusController)             // 任务状态
	router.GET("/v1/accountLog/:uid", userApi.AccountLogController)                   // 账号运行日志
	router.POST("/v1/startBrush/:uid", userApi.StartBrushController)                  // 启动刷课
	router.GET("/v1/startBrush/:uid", userApi.StartBrushController)                   // 兼容旧调用
	router.POST("/v1/stopBrush/:uid", userApi.StopBrushController)                    // 停止刷课
	router.GET("/v1/stopBrush/:uid", userApi.StopBrushController)                     // 兼容旧调用

	router.GET("/v1/setting", userApi.GetSettingController)   // 读取全局设置
	router.POST("/v1/setting", userApi.SaveSettingController) // 保存全局设置

	router.GET("/v1/streamLog/:id", userApi.StreamLog) // 日志推送
}
