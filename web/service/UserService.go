package service

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"
	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/dto"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/entity/vo"
	"yatori-go-console/global"
	"yatori-go-console/utils"
	"yatori-go-console/web/activity"

	"github.com/gin-gonic/gin"
)

func AccountLogsService(c *gin.Context) {
	uid := c.Param("uid")
	data, err := getLocalConfigUserLogs(uid, 300)
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "拉取日志成功",
		Data:    data,
	})
}

// 拉取账号列表
func UserListService(c *gin.Context) {
	users, err := syncUsersFromConfigManager()
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	//转换列表--------------
	resUserList := []map[string]any{}
	for _, user := range users {
		toMap := utils.StructToMap(user)
		userActivity := global.GetUserActivity(user)
		if userActivity != nil {
			if xxt, ok := (*userActivity).(*activity.XXTActivity); ok {
				toMap["isRunning"] = xxt.IsRunning
			}
		} else {
			toMap["isRunning"] = false
		}

		resUserList = append(resUserList, toMap)
	}
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "拉取账号成功",
		Data: gin.H{
			"users": resUserList,
			"total": len(resUserList),
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
	addedUserDTO, err := upsertLocalConfigUser(dto.ConfigManagerUser{
		AccountType:  req.AccountType,
		URL:          req.Url,
		Account:      req.Account,
		Password:     req.Password,
		IsProxy:      0,
		InformEmails: []string{},
		CoursesCustom: config.CoursesCustom{
			IncludeCourses:  []string{},
			ExcludeCourses:  []string{},
			CoursesSettings: []config.CoursesSettings{},
		},
	})
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	users, err := syncUsersFromConfigManager()
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	var addedUser *pojo.UserPO
	for i := range users {
		if users[i].Uid == addedUserDTO.Uid {
			addedUser = &users[i]
			break
		}
	}
	if addedUser == nil {
		addedUser, err = configManagerUserToPO(*addedUserDTO)
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: err.Error(),
			})
			return
		}
	}

	c.JSON(200, vo.Response{
		Code:    200,
		Message: "添加账号成功",
		Data:    addedUser,
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
		if err := deleteLocalConfigUser(req.Uid, strings.TrimSpace(c.GetHeader("X-Admin-Pass"))); err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: err.Error(),
			})
			return
		}
	} else if req.AccountType != "" && req.Account != "" { //如果uid方式没有，则直接使用账号和账号类型方式联合查询删除
		syncedUsers, err := syncUsersFromConfigManager()
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: err.Error(),
			})
			return
		}
		for _, user := range syncedUsers {
			if user.AccountType == req.AccountType && user.Url == req.Url && user.Account == req.Account {
				if err := deleteLocalConfigUser(user.Uid, strings.TrimSpace(c.GetHeader("X-Admin-Pass"))); err != nil {
					c.JSON(http.StatusOK, vo.Response{
						Code:    400,
						Message: err.Error(),
					})
					return
				}
				break
			}
		}
	}

	if _, err := syncUsersFromConfigManager(); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
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
	user, err := getLocalConfigUserByUID(uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	userPO, err := configManagerUserToPO(*user)
	if err == nil {
		_ = dao.UpsertUser(global.GlobalDB, userPO)
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
	var req dto.ConfigManagerUser

	// 绑定 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		data, _ := c.GetRawData()
		fmt.Println(string(data))
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "JSON 解析失败",
		})
		return
	}
	// Uid 必须存在
	if strings.TrimSpace(req.Uid) == "" {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "UID 不能为空",
		})
		return
	}
	if err := validateAutoExecutionWindow(req.CoursesCustom.AutoRunStartTime, req.CoursesCustom.AutoRunEndTime); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}

	user, err := getLocalConfigUserByUID(req.Uid)
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	changed := false
	if req.AccountType != "" {
		user.AccountType = req.AccountType
		changed = true
	}
	if req.URL != "" {
		user.URL = req.URL
		changed = true
	}
	if req.Account != "" {
		user.Account = req.Account
		changed = true
	}
	if req.Password != "" {
		user.Password = req.Password
		changed = true
	}
	if req.RemarkName != user.RemarkName {
		user.RemarkName = req.RemarkName
		changed = true
	}
	if req.InformEmails != nil {
		user.InformEmails = req.InformEmails
		changed = true
	}
	if !reflect.DeepEqual(req.CoursesCustom, config.CoursesCustom{}) {
		user.CoursesCustom = req.CoursesCustom
		changed = true
	}
	if req.IsProxy != user.IsProxy {
		user.IsProxy = req.IsProxy
		changed = true
	}

	if !changed {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "没有可更新的字段",
		})
		return
	}

	if _, err := updateLocalConfigUser(req.Uid, func(localUser *dto.ConfigManagerUser) error {
		*localUser = *user
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	if _, err := syncUsersFromConfigManager(); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}
	if err := syncAutoExecutionSchedules(time.Now()); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 拉取课程列表
func AccountCourseListService(c *gin.Context) {
	// 2. 解析 JSON 到结构体
	uid := c.Param("uid")
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if user == nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "该账号不存在",
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
		//转换为标准类型的列表数据
		courseList := []vo.CourseInformResponse{}
		for _, course := range list {
			courseList = append(courseList, vo.CourseInformResponse{
				CourseId:   course.CourseID,
				CourseName: course.CourseName,
				Progress:   float32(course.JobRate),
				Instructor: course.CourseTeacher,
			})
		}
		//fmt.Println(list)
		c.JSON(http.StatusOK, vo.Response{
			Code:    200,
			Message: "拉取信息成功",
			Data:    gin.H{"courseList": courseList},
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
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	userActivity := global.GetUserActivity(*user)
	if userActivity == nil {
		createActivity := activity.BuildUserActivity(*user)
		if createActivity == nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: "当前账号类型暂不支持启动",
			})
			return
		}
		userActivity = &createActivity
		global.PutUserActivity(*user, &createActivity)
	}
	go func() {
		if err := (*userActivity).Start(); err != nil {
			fmt.Println("start brush failed:", err)
		}
	}()

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
	// 根据账号类型断言为具体活动类型并设置IsRunning
	if xxt, ok := (*userActivity).(*activity.XXTActivity); ok {
		xxt.IsRunning = false
	} else if yinghua, ok := (*userActivity).(*activity.YingHuaActivity); ok {
		yinghua.IsRunning = false
	}
	(*userActivity).Stop()
	//userActivity.Kill()
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "停止成功",
	})
}
