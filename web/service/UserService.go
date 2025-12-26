package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/entity/vo"
	"yatori-go-console/global"
	"yatori-go-console/web/activity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 拉取账号列表
func UserListService(c *gin.Context) {
	users, total, err := dao.QueryUsers(global.GlobalDB, 1, 10)
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "拉取账号成功",
		Data: gin.H{
			"users": users,
			"total": total,
		},
	})
}

// 添加账号
func AddUserService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req vo.AddAccountRequest
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}
	//检测账号是否已存在
	userPo := pojo.UserPO{
		AccountType: req.AccountType,
		Url:         req.Url,
		Account:     req.Account,
	}
	user, _ := dao.QueryUser(global.GlobalDB, userPo)
	if user != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号已存在",
		})
		return
	}

	uuidV7, _ := uuid.NewV7()
	userPo.Uid = uuidV7.String()   //设置uuid值
	userPo.Password = req.Password //设置密码
	userConfig := config.User{
		AccountType: userPo.AccountType,
		URL:         userPo.Url,
		Account:     userPo.Account,
		Password:    userPo.Password,
	}

	userConfigJson, err2 := json.Marshal(userConfig)
	if err2 != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err2.Error(),
		})
		return
	}
	userPo.UserConfigJson = string(userConfigJson) //赋值Config配置

	err := dao.InsertUser(global.GlobalDB, &userPo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	//登录成功
	c.JSON(200,
		vo.Response{
			Code:    200,
			Message: "添加账号成功",
			Data:    &userPo,
		})
}

// 删除账号
func DeleteUserService(c *gin.Context) {
	var req vo.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "数据转换失败",
		})
		return
	}
	//如果uid不为空则采用uid方式删除
	if req.Uid != "" {
		err := dao.DeleteUser(global.GlobalDB, &pojo.UserPO{Uid: req.Uid})
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: "删除失败",
			})
			return
		}
	} else if req.AccountType != "" && req.Account != "" { //如果uid方式没有，则直接使用账号和账号类型方式联合查询删除
		err := dao.DeleteUser(global.GlobalDB, &pojo.UserPO{
			AccountType: req.AccountType,
			Url:         req.Url,
			Account:     req.Account,
		})
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: "删除失败",
			})
			return
		}
	}

	c.JSON(200,
		vo.Response{
			Code:    200,
			Message: "删除成功",
		})
}

// 检查账号密码是否正确
func AccountLoginCheckService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req vo.AccountLoginCheckRequest
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK,
			vo.Response{
				Code:    400,
				Message: "请求参数错误: " + err.Error(),
			})
		return
	}
	//如果是uid检测登录则先查询
	if req.Uid != "" {
		user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: req.Uid})
		if user == nil {
			c.JSON(http.StatusOK,
				vo.Response{
					Code:    400,
					Message: "该账号不存在",
				})
			return
		}
		//登录逻辑......
	}

	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "账号登录正常",
	})
}

// 获取账号配置信息
func GetAccountInformService(c *gin.Context) {
	uid := c.Param("uid")
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号不存在",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "拉取信息成功",
		"data": gin.H{
			"user": user,
		},
	})
}

// 登录账号
func LoginUserService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req pojo.UserPO
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
	})
}

// 更新账号信息
func UpdateUserService(c *gin.Context) {
	var req pojo.UserPO

	// 绑定 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		data, _ := c.GetRawData()
		fmt.Println(string(data))
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "JSON 解析失败",
		})
		return
	}

	// Uid 必须存在
	if req.Uid == "" {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "UID 不能为空",
		})
		return
	}

	// 将结构体转为 map 并过滤空字段
	updateMap := make(map[string]interface{})

	// 手动挑选可修改字段（最安全方式）
	if req.AccountType != "" {
		updateMap["account_type"] = req.AccountType
	}
	if req.Url != "" {
		updateMap["url"] = req.Url
	}
	if req.Account != "" {
		updateMap["account"] = req.Account
	}
	if req.Password != "" {
		updateMap["password"] = req.Password
	}

	// 空字段检查
	if len(updateMap) == 0 {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "没有可更新的字段",
		})
		return
	}

	// 调用 DAO 更新
	if err := dao.UpdateUser(global.GlobalDB, req.Uid, updateMap); err != nil {
		c.JSON(500, gin.H{
			"code": 500,
			"msg":  err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 拉取账号状态信息
func AccountCourseListService(c *gin.Context) {
	// 2. 解析 JSON 到结构体
	uid := c.Param("uid")
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号不存在",
		})
		return
	}

	userActivity := global.GetUserActivity(*user)
	//如果没有活动中的账号则添加活动账号
	if userActivity == nil {
		//构建用户活动
		createActivity := activity.BuildUserActivity(*user)
		userActivity = &createActivity
		global.PutUserActivity(*user, &createActivity)
	}
	//如果是学习通
	if xxt, ok1 := (*userActivity).(activity.XXTAbility); ok1 {
		list, err := xxt.PullCourseList()
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println(list)
		c.JSON(http.StatusOK, gin.H{
			"code": 200,
			"msg":  "拉取信息成功",
			"data": gin.H{
				"courseList": list,
			},
		})
	}

}

// 开始刷课
func StartBrushService(c *gin.Context) {
	uid := c.Param("uid")
	user, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	userActivity := global.GetUserActivity(*user)
	if userActivity != nil {
		//userActivity.IsRunning = false
	}
	//activity := activity.UserActivity{UserPO: *user}
	//global.PutUserActivity(*user, &activity)
	//userActivity = global.GetUserActivity(*user)
	//err1 := activity.UserLoginOperation()
	//if err1 != nil {
	//	c.JSON(400, gin.H{
	//		"code": 400,
	//		"msg":  err1.Error(),
	//	})
	//}
	//
	//userActivity.Start()
	//userActivity.UserPO = *user
	//go userActivity.UserBlock()
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "启动成功",
	})
}

// 停止任务
func StopBrushService(c *gin.Context) {
	uid := c.Param("uid")
	user, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if err != nil {
		c.JSON(400, gin.H{})
	}
	if user == nil {
		c.JSON(400, gin.H{})
		return
	}
	userActivity := global.GetUserActivity(*user)
	if userActivity == nil {
		c.JSON(400, gin.H{})
	}
	//userActivity.Kill()
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "停止成功",
	})
}

// 拉取课程列表
func CourseListService(c *gin.Context) {

}
