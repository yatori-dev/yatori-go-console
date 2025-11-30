package cqie

import (
	"fmt"
	"log"
	"sync"
	"time"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	"github.com/yatori-dev/yatori-go-core/aggregation/cqie"
	cqieApi "github.com/yatori-dev/yatori-go-core/api/cqie"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤Cqie账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "CQIE" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*cqieApi.CqieUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*cqieApi.CqieUserCache {
	var UserCaches []*cqieApi.CqieUserCache
	for _, user := range users {
		if user.AccountType == "CQIE" {
			cache := &cqieApi.CqieUserCache{Account: user.Account, Password: user.Password}
			error := cqie.CqieLoginAction(cache) // 登录
			if error != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, error.Error())
				log.Fatal(error) //登录失败则直接退出
			}
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.User, cache *cqieApi.CqieUserCache) {
	// projectList, _ := enaea.ProjectListAction(cache) //拉取项目列表
	courseList, _ := cqie.CqiePullCourseListAction(cache)
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
func nodeListStudy(setting config.Setting, user *config.User, userCache *cqieApi.CqieUserCache, course *cqie.CqieCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.CourseName, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.CourseName, user.CoursesCustom.IncludeCourses) {
		return
	}
	//执行刷课---------------------------------
	nodeList, _ := cqie.PullCourseVideoListAndProgress(userCache, course) //拉取对应课程的视频列表
	//失效重登检测
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, "【"+course.CourseName+"】 ")
	// 提交学时
	for _, node := range nodeList {
		//视频处理逻辑
		switch user.CoursesCustom.VideoModel {
		case 1:
			videoAction(setting, user, userCache, node) //常规
			break
		case 2:
			videoActionSecondBrush(setting, user, userCache, node) //暴力模式（秒刷
			break
		}
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", "【"+course.CourseName+"】 ", "学习完毕")

}

// videoAction 刷视频逻辑抽离
func videoAction(setting config.Setting, user *config.User, UserCache *cqieApi.CqieUserCache, node cqie.CqieVideo) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}

	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在学习视频：", lg.Default, "【"+node.VideoName+"】 ")
	nowTime := time.Now()
	startPos := node.StudyTime
	stopPos := node.StudyTime
	maxPos := node.StudyTime
	err := cqie.SaveVideoStudyTimeAction(UserCache, &node, startPos, stopPos) //每次刷课前都得先获取一遍，因为要获取学习分配的id
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), `[`, UserCache.Account, `] `, lg.BoldRed, err.Error())
	}
	for {
		if maxPos >= node.TimeLength+3 { //+3是为了防止漏时
			startPos = node.TimeLength
			stopPos = node.TimeLength
			maxPos = node.TimeLength
			break
		}
		if stopPos >= maxPos {
			maxPos = startPos + 3
		}
		err := cqie.SubmitStudyTimeAction(UserCache, &node, nowTime, startPos, stopPos, maxPos)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", "【"+node.VideoName+"】", lg.BoldRed, "提交学时异常：", err.Error())
		}
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", "【"+node.VideoName+"】  >>> ", "提交状态：", "成功", lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", float32(node.StudyTime)/float32(node.TimeLength)), "%")
		startPos = startPos + 3
		stopPos = stopPos + 3
		time.Sleep(3 * time.Second)
	}
	err = cqie.SaveVideoStudyTimeAction(UserCache, &node, startPos, stopPos) //学完之后保存学习点
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", "【"+node.VideoName+"】", lg.BoldRed, "保存学习点异常：", err.Error())
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "视频：", lg.Default, "【"+node.VideoName+"】 ", lg.Green, "学习完毕")
}

// videoAction 刷视频逻辑抽离(秒刷版本)
func videoActionSecondBrush(setting config.Setting, user *config.User, UserCache *cqieApi.CqieUserCache, node cqie.CqieVideo) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}
	nowTime := time.Now()
	startPos := node.StudyTime
	stopPos := node.StudyTime
	maxPos := node.StudyTime
	err := cqie.SaveVideoStudyTimeAction(UserCache, &node, startPos, stopPos) //每次刷课前都得先获取一遍，因为要获取学习分配的id
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), `[`, UserCache.Account, `] `, lg.BoldRed, err.Error())
	}
	err1 := cqie.SubmitStudyTimeAction(UserCache, &node, nowTime, startPos, stopPos, maxPos)
	if err1 != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", "【"+node.VideoName+"】", lg.BoldRed, "提交学时异常：", err.Error())
	}
	err = cqie.SaveVideoStudyTimeAction(UserCache, &node, startPos, stopPos) //学完之后保存学习点
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", "【"+node.VideoName+"】", lg.BoldRed, "保存学习点异常：", err.Error())
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "视频：", lg.Default, "【"+node.VideoName+"】 ", lg.Green, "学习完毕")
}
