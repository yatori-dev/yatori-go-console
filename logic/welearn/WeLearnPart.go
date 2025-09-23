package welearn

import (
	"fmt"
	"log"
	"sync"
	"time"
	"yatori-go-console/config"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	action "github.com/yatori-dev/yatori-go-core/aggregation/welearn"
	"github.com/yatori-dev/yatori-go-core/api/welearn"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤Cqie账号
func FilterAccount(configData *config.JSONDataForConfig) []config.Users {
	var users []config.Users //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "WELEARN" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.Users, userCaches []*welearn.WeLearnUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.Users) []*welearn.WeLearnUserCache {
	var UserCaches []*welearn.WeLearnUserCache
	for _, user := range users {
		if user.AccountType == "WELEARN" {
			cache := &welearn.WeLearnUserCache{Account: user.Account, Password: user.Password}
			err := action.WeLearnLoginAction(cache) // 登录
			if err != nil {
				lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, err.Error())
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

func userBlock(setting config.Setting, user *config.Users, cache *welearn.WeLearnUserCache) {
	// projectList, _ := enaea.ProjectListAction(cache) //拉取项目列表
	courseList, err := action.WeLearnPullCourseListAction(cache)
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

	lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	if setting.BasicSetting.CompletionTone == 1 { //如果声音提示开启，那么播放
		soundMut.Lock()
		utils2.PlayNoticeSound() //播放提示音
		soundMut.Unlock()
	}
	usersLock.Done()
}

// 章节节点的抽离函数
func nodeListStudy(setting config.Setting, user *config.Users, userCache *welearn.WeLearnUserCache, course *action.WeLearnCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.Name, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.Name, user.CoursesCustom.IncludeCourses) {
		return
	}
	//执行刷课---------------------------------

	//nodeList := ketangx.PullNodeListAction(userCache, course) //拉取对应课程的视频列表
	//失效重登检测
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, "【"+course.Name+"】 ")
	// 提交学时
	chapterList, err := action.WeLearnPullCourseChapterAction(userCache, *course) //拉取对应课程的章节
	if err != nil {
		panic(err)
	}
	for _, chapter := range chapterList {
		pointList, err1 := action.WeLearnPullChapterPointAction(userCache, *course, chapter)
		if err1 != nil {
			panic(err1)
		}
		for _, point := range pointList {
			//视频处理逻辑
			switch user.CoursesCustom.VideoModel {
			case 1:
				nodeSubmitTimeAction(setting, user, userCache, course, point) //刷学时
				break
			case 2:
				nodeCompleteAction(setting, user, userCache, course, point) //刷完成度
			}
		}
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", "【"+course.Name+"】 ", "学习完毕")

}

// videoAction 刷视频逻辑抽离，普通模式就是秒刷
func nodeCompleteAction(setting config.Setting, user *config.Users, UserCache *welearn.WeLearnUserCache, course *action.WeLearnCourse, node action.WeLearnPoint) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}
	if !node.IsVisible {
		lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Name+"】", lg.Yellow, "该任务点还未解锁，已自动跳过")
		return
	}
	//如果完成了的直接跳过
	if node.IsComplete == "completed" {
		return
	}
	//action, err := ketangx.CompleteVideoAction(UserCache, &node)
	err := action.WeLearnCompletePointAction(UserCache, *course, node)
	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Name+"】", lg.BoldRed, "学习异常：", err.Error())
		return
	}

	lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Name+"】", lg.Green, "学习完毕")

}

// 累计学时
func nodeSubmitTimeAction(setting config.Setting, user *config.Users, UserCache *welearn.WeLearnUserCache, course *action.WeLearnCourse, node action.WeLearnPoint) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}
	if !node.IsVisible {
		lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Location+"】", lg.Yellow, "该任务点还未解锁，已自动跳过")
		return
	}
	//如果完成了的直接跳过
	//if node.IsComplete == "completed" {
	//	return
	//}

	_, progressMeasure, sessionTime, totalTime, scaled, err := action.WeLearnSubmitStudyTimeAction(UserCache, *course, node)
	if err != nil {
		fmt.Println(err)
	}
	endTime := 1600
	//比阈值大就直接返回
	if totalTime > endTime {
		return
	}
	for {
		api, err1 := UserCache.KeepPointSessionPlan1Api(course.Cid, node.Id, course.Uid, course.ClassId, sessionTime, totalTime, 3, nil)
		if err1 != nil {
			fmt.Println(err1)
		}
		//fmt.Println(totalTime, api)
		lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Location+"】", lg.Green, "学时提交成功，进度: ", fmt.Sprintf("%d/%d", totalTime, endTime), "服务器返回信息: ", api)
		if sessionTime >= endTime {
			break
		}
		sessionTime += 60
		totalTime += 60
		time.Sleep(time.Duration(60) * time.Second)
	}
	_, err2 := UserCache.SubmitStudyPlan2Api(course.Cid, node.Id, course.Uid, scaled, course.ClassId, progressMeasure, "completed", 3, nil)
	if err2 != nil {
		fmt.Println(err2)
	}
	//fmt.Println(submitApi2)

	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Location+"】", lg.BoldRed, "学习异常：", err.Error())
		return
	}

	lg.Print(lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Default, "【"+course.Name+"】 ", "【"+node.Location+"】", lg.Green, "学习完毕")

}
