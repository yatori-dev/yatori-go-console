package web

import "yatori-go-console/web/controller"

func (router Group) StaticRouter() {
	// 主页已在ServerInit.go中用router.StaticFile注册
}
func (router Group) ApiV1Router() {
	var userApi controller.UserApi
	router.GET("/v1/accountList", userApi.AccountListController)                      //拉取账号列表
	router.POST("/v1/addAccount", userApi.AddAccountController)                       //添加账号
	router.POST("/v1/deleteAccount", userApi.DeleteAccountController)                 //删除账号
	router.POST("/v1/updateAccount", userApi.UpdateAccountController)                 //修改账号信息
	router.POST("/v1/accountLoginCheck", userApi.AccountLoginCheckController)         //账号登录检测，用于检测账号密码是否正确
	router.GET("/v1/getAccountInformForUid/:uid", userApi.GetAccountInformController) //拉取配置数据
	router.GET("/v1/getAccountCourseList/:uid", userApi.AccountCourseListController)  //获取课程列表
	router.GET("/v1/startBrush/:uid", userApi.StartBrushController)                   //启动刷课
	router.GET("/v1/stopBrush/:uid", userApi.StopBrushController)                     //停止刷课

	router.GET("/v1/streamLog/:id", userApi.StreamLog) //推送日志
}
