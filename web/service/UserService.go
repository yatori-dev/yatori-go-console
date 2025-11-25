package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"yatori-go-console/dao"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/global"
	"yatori-go-console/logic"
)

func UserListService(c *gin.Context) {
	users, total, err := dao.QueryUsers(global.GlobalDB, 1, 10)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "拉取账号失败",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"msg":   "拉取账号成功",
		"users": users,
		"total": total,
	})
}
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

	updateMap["is_proxy"] = req.IsProxy

	updateMap["email_inform_sw"] = req.EmailInformSw

	updateMap["inform_emails"] = req.InformEmails

	if req.WeLearnTime != "" {
		updateMap["weLearn_time"] = req.WeLearnTime
	}

	updateMap["cxNode"] = req.CxNode

	updateMap["shuffle_sw"] = req.ShuffleSw

	updateMap["video_model"] = req.VideoModel

	updateMap["auto_exam"] = req.AutoExam

	updateMap["exam_auto_submit"] = req.ExamAutoSubmit

	updateMap["exclude_courses"] = req.ExcludeCourses

	updateMap["include_courses"] = req.IncludeCourses

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

	activity := global.GetUserActivity(*user)
	//如果没有活动中的账号则添加活动账号
	if activity == nil {
		//构建用户活动
		createActivity := logic.UserActivity{
			UserPO: *user,
		}
		global.PutUserActivity(*user, &createActivity)
		activity = global.GetUserActivity(*user)
	}
	if !activity.IsLogin {
		//登录
		err := activity.UserLoginOperation()
		activity.IsLogin = true
		//如果登录失败
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"code": 400,
				"msg":  err.Error(),
			})
			return
		}
	}

	//拉取课程列表
	list, err := activity.PullCourseList()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "拉取信息成功",
		"data": gin.H{
			"isLogin":    activity.IsLogin,
			"isRunning":  activity.IsRunning,
			"courseList": list,
		},
	})

}

// 添加用户
func AddUserService(c *gin.Context) {
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
	req.InformEmails = []string{}
	req.IncludeCourses = []string{}
	req.ExcludeCourses = []string{}
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, req)
	if user != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号已存在",
		})
		return
	}

	uuidV7, _ := uuid.NewV7()
	req.Uid = uuidV7.String() //设置uuid值
	//构建用户活动
	activity := logic.UserActivity{
		UserPO: pojo.UserPO{
			Uid:         req.Uid,
			AccountType: req.AccountType,
			Account:     req.Account,
			Password:    req.Password},
	}

	//登录
	err := activity.UserLoginOperation()

	//如果登录失败
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	activity.IsLogin = true //
	global.UserActivityMap[fmt.Sprintf("%s-%s-%s", req.AccountType, req.Url, req.Account)] = &activity
	err = dao.InsertUser(global.GlobalDB, &req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	//登录成功
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "添加账号成功",
		"user": &req,
	})
}

// 删除账号
func DeleteUserService(c *gin.Context) {
	var req pojo.UserPO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "数据转换失败",
		})
		return
	}
	err := dao.DeleteUser(global.GlobalDB, req.Uid)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "删除失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
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
	if userActivity == nil {
		activity := logic.UserActivity{UserPO: *user}
		global.PutUserActivity(*user, &activity)
		userActivity = global.GetUserActivity(*user)
		err1 := activity.UserLoginOperation()
		if err1 != nil {
			c.JSON(400, gin.H{
				"code": 400,
				"msg":  err1.Error(),
			})
		}
	}

	go userActivity.UserBlock()
	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "启动成功",
	})
}

// 拉取课程列表
func CourseListService(c *gin.Context) {

}
