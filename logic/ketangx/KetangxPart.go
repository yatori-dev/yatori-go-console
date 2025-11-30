package ketangx

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/ketangx"
	ketangxApi "github.com/yatori-dev/yatori-go-core/api/ketangx"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤Cqie账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "KETANGX" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*ketangxApi.KetangxUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*ketangxApi.KetangxUserCache {
	var UserCaches []*ketangxApi.KetangxUserCache
	for _, user := range users {
		if user.AccountType == "KETANGX" {
			cache := &ketangxApi.KetangxUserCache{Account: user.Account, Password: user.Password}
			err := ketangx.LoginAction(cache) // 登录
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, err.Error())
				log.Fatal(err) //登录失败则直接退出
			}
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.User, cache *ketangxApi.KetangxUserCache) {
	// projectList, _ := enaea.ProjectListAction(cache) //拉取项目列表
	courseList := ketangx.PullCourseListAction(cache)
	for _, course := range courseList {
		videosLock.Add(1)
		go func() {
			nodeListStudy(setting, user, cache, &course) //多携程刷课
			videosLock.Done()
		}()
	}
	videosLock.Wait() //等待课程刷完

	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	//如果开启了邮箱通知
	if setting.EmailInform.Sw == 1 && len(user.InformEmails) > 0 {
		utils2.SendMail(setting.EmailInform.SMTPHost, setting.EmailInform.SMTPPort, setting.EmailInform.UserName, setting.EmailInform.Password, user.InformEmails, fmt.Sprintf("账号：[%s]</br>平台：[%s]</br>通知：所有课程已执行完毕", user.Account, user.AccountType))
	}
	if setting.BasicSetting.CompletionTone == 1 { //如果声音提示开启，那么播放
		soundMut.Lock()
		utils2.PlayNoticeSound() //播放提示音
		soundMut.Unlock()
	}
	usersLock.Done()
}

// 章节节点的抽离函数
func nodeListStudy(setting config.Setting, user *config.User, userCache *ketangxApi.KetangxUserCache, course *ketangx.KetangxCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.Title, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.Title, user.CoursesCustom.IncludeCourses) {
		return
	}
	//执行刷课---------------------------------
	nodeList := ketangx.PullNodeListAction(userCache, course) //拉取对应课程的视频列表
	//失效重登检测
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, "【"+course.Title+"】 ")
	// 提交学时
	for _, node := range nodeList {
		//视频处理逻辑
		switch user.CoursesCustom.VideoModel {
		case 1:
			videoAction(setting, user, userCache, course, node) //常规
			break
		}
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", "【"+course.Title+"】 ", "学习完毕")

}

// videoAction 刷视频逻辑抽离，普通模式就是秒刷
func videoAction(setting config.Setting, user *config.User, UserCache *ketangxApi.KetangxUserCache, course *ketangx.KetangxCourse, node ketangx.KetangxNode) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}
	if node.IsComplete {
		return
	}
	action, err := ketangx.CompleteVideoAction(UserCache, &node)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Title+"】 ", "【"+node.Title+"】", lg.BoldRed, "结点类型: ", "<", node.Type, "> ", "学习异常：", err.Error())
		return
	}
	status := gojsonq.New().JSONString(action).Find("Success")
	if status != nil && !status.(bool) {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Title+"】 ", "【"+node.Title+"】", lg.BoldRed, "结点类型: ", "<", node.Type, "> ", "学习异常：", action)
		return
	}
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Title+"】 ", "【"+node.Title+"】", "结点类型: ", "<", lg.Yellow, node.Type, lg.Default, "> ", lg.Green, "学习完毕，服务器返回状态:"+strconv.FormatBool(status.(bool)))
	//modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO,fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]),"[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在学习视频：", lg.Default, "【"+node.Title+"】 ")

	//modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO,fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]),"[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "视频：", lg.Default, "【"+node.Title+"】 ", lg.Green, "学习完毕")
}
