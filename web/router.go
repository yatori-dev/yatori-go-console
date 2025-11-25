package bootstrap

import "yatori-go-console/web/controller"

// 路由
func (router Group) Router() {
	var userApi controller.UserApi
	router.GET("/", userApi.IndexHtml)                                                //主页
	router.GET("/api/accountList", userApi.AccountListController)                     //拉取账号列表
	router.POST("/api/loginAccount", userApi.LoginAccountController)                  //用户登录
	router.POST("/api/addAccount", userApi.AddAccountController)                      //添加账号
	router.POST("/api/updateAccount", userApi.UpdateAccountController)                //修改账号信息
	router.POST("/api/deleteAccount", userApi.DeleteAccountController)                //删除账号
	router.GET("/api/getAccountInform/:uid", userApi.GetAccountInformController)      //拉取配置数据
	router.GET("/api/getAccountCourseList/:uid", userApi.AccountCourseListController) //获取课程列表
	//router.GET("/api/courseList", userApi.CourseListController)
	router.GET("/api/startBrush/:uid", userApi.StartBrushController)
	router.GET("/api/stopBrush/:uid")
	router.GET("/api/setVideoModel")
	router.GET("/api/setExamModel")
	router.GET("/api/streamLog/:id", userApi.StreamLog) //推送日志
}
