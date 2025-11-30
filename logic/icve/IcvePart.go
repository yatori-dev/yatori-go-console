package icve

import (
	"fmt"
	"log"
	"sync"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	action "github.com/yatori-dev/yatori-go-core/aggregation/icve"
	"github.com/yatori-dev/yatori-go-core/api/icve"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤ICVE账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "ICVE" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*icve.IcveUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*icve.IcveUserCache {
	var UserCaches []*icve.IcveUserCache
	for _, user := range users {
		if user.AccountType == "ICVE" {
			if len(user.Password) > 30 {
				cache := &icve.IcveUserCache{Account: user.Account, Password: user.Password}
				err := action.IcveCookieLogin(cache)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Red, err.Error())
					log.Fatal(err) //登录失败则直接退出
				}
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Green, "登录成功")
				UserCaches = append(UserCaches, cache)
			} else {
				cache := &icve.IcveUserCache{Account: user.Account, Password: user.Password}

				err := action.IcveLoginAction(cache)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Red, err.Error())
					log.Fatal(err) //登录失败则直接退出
				}
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Green, "登录成功")
				UserCaches = append(UserCaches, cache)
			}

		}
	}
	return UserCaches
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.User, cache *icve.IcveUserCache) {
	courseList, err := action.PullZYKCourseAction(cache)
	if err != nil {
		panic(err)
	}
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
func nodeListStudy(setting config.Setting, user *config.User, userCache *icve.IcveUserCache, course *action.IcveCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.CourseName, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.CourseName, user.CoursesCustom.IncludeCourses) {
		return
	}
	//如果课程已结束，那么直接跳过
	if course.Status == "3" {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Yellow, "【"+course.CourseName+"】 ", "该课程已结束，已自动跳过...")
		return
	}
	//执行刷课---------------------------------

	//失效重登检测
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, "【"+course.CourseName+"】 ")
	// 提交学时
	chapterList, err := action.PullZYKCourseNodeAction(userCache, *course) //拉取对应课程的章节
	if err != nil {
		panic(err)
	}
	for _, point := range chapterList {
		//任务点处理逻辑
		switch user.CoursesCustom.VideoModel {
		case 1:
			nodeCompleteAction(setting, user, userCache, course, point) //刷学时
			break
		}
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", "【"+course.CourseName+"】 ", "学习完毕")

}

// videoAction 刷视频逻辑抽离，普通模式就是秒刷
func nodeCompleteAction(setting config.Setting, user *config.User, UserCache *icve.IcveUserCache, course *action.IcveCourse, node action.IcveCourseNode) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}
	//如果完成了的直接跳过
	if node.Speed >= 100 {
		return
	}
	submitResult, err2 := action.SubmitZYKStudyTimeAction(UserCache, node)
	if err2 != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.Name+"】", lg.BoldRed, "学习异常：", err2.Error())
		return
	}

	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.Name+"】", lg.Green, "学习完毕,学习状态：", submitResult)
}
