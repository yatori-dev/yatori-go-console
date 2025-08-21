package enaea

import (
	"fmt"
	"log"
	"os"
	"sync"
	time2 "time"
	"yatori-go-console/config"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	"github.com/yatori-dev/yatori-go-core/aggregation/enaea"
	enaeaApi "github.com/yatori-dev/yatori-go-core/api/enaea"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤学习公社账号
func FilterAccount(configData *config.JSONDataForConfig) []config.Users {
	var users []config.Users //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "ENAEA" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.Users, userCaches []*enaeaApi.EnaeaUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.Users) []*enaeaApi.EnaeaUserCache {
	var UserCaches []*enaeaApi.EnaeaUserCache
	for _, user := range users {
		if user.AccountType == "ENAEA" {
			cache := &enaeaApi.EnaeaUserCache{Account: user.Account, Password: user.Password}
			_, error := enaea.EnaeaLoginAction(cache) // 登录
			if error != nil {
				lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, error.Error())
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

func userBlock(setting config.Setting, user *config.Users, cache *enaeaApi.EnaeaUserCache) {

	projectList, _ := enaea.ProjectListAction(cache) //拉取项目列表
	for _, course := range projectList {
		//过滤项目---------------------------------
		//排除指定项目
		if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.ClusterName, user.CoursesCustom.ExcludeCourses) {
			continue
		}
		//包含指定课程
		if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.ClusterName, user.CoursesCustom.IncludeCourses) {
			continue
		}
		courseList, err := enaea.CourseListAction(cache, course.CircleId)
		if err != nil {
			lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.Default, "] ", lg.BoldRed, "拉取项目列表错误", err.Error())
			os.Exit(0)
		}
		lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "正在学习项目", " 【"+course.ClusterName+"】 ")
		for _, item := range courseList { //遍历所有待刷视频
			nodeListStudy(setting, user, cache, &item) //多携程刷课
		}
	}

	lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	if setting.BasicSetting.CompletionTone == 1 { //如果声音提示开启，那么播放
		soundMut.Lock()
		utils2.PlayNoticeSound() //播放提示音
		soundMut.Unlock()
	}
	usersLock.Done()
}

// 章节节点的抽离函数
func nodeListStudy(setting config.Setting, user *config.Users, userCache *enaeaApi.EnaeaUserCache, course *enaea.EnaeaCourse) {
	//执行刷课---------------------------------
	nodeList, err := enaea.VideoListAction(userCache, course) //拉取对应课程的视频列表
	//失效重登检测
	for err != nil {
		enaea.LoginTimeoutAfreshAction(userCache, err)
		nodeList1, err1 := enaea.VideoListAction(userCache, course) //拉取对应课程的视频列表
		nodeList = nodeList1
		err = err1
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, "【"+course.TitleTag+"】", "【"+course.CourseTitle+"】 ")
	// 提交学时
	for _, node := range nodeList {
		//视频处理逻辑
		videoAction(setting, user, userCache, node)
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", " 【"+course.TitleTag+"】", "【"+course.CourseTitle+"】 ", "学习完毕")

}

// videoAction 刷视频逻辑抽离
func videoAction(setting config.Setting, user *config.Users, UserCache *enaeaApi.EnaeaUserCache, node enaea.EnaeaVideo) {
	if user.CoursesCustom.VideoModel == 0 { //是否打开了自动刷视频开关
		return
	}

	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在学习视频：", lg.Default, " 【"+node.TitleTag+"】", "【"+node.CourseName+"】", "【"+node.CourseContentStr+"】 ")
	err := enaea.StatisticTicForCCVideAction(UserCache, &node)
	if err != nil {
		lg.Print(lg.INFO, `[`, UserCache.Account, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
	}
	for {
		if node.StudyProgress >= 100 {
			modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", " 【"+node.TitleTag+"】", " 【"+node.CourseName+"】", "【"+node.CourseContentStr+"】", " ", lg.Blue, "学习完毕")
			break //如果看完了，也就是进度为100那么直接跳过
		}
		//提交学时
		var err error
		if user.CoursesCustom.VideoModel == 0 {
			err = enaea.SubmitStudyTimeAction(UserCache, &node, time2.Now().UnixMilli(), 0)
		} else if user.CoursesCustom.VideoModel == 1 {
			err = enaea.SubmitStudyTimeAction(UserCache, &node, 60, 1)
		}

		if err != nil {
			if err.Error() != "request frequently!" {
				lg.Print(lg.INFO, `[`, UserCache.Account, `] `, " 【"+node.TitleTag+"】", "【"+node.CourseName+"】", "【"+node.CourseContentStr+"】 ", lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
			}
		}
		//失效重登检测
		enaea.LoginTimeoutAfreshAction(UserCache, err)

		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, "[", lg.Green, UserCache.Account, lg.Default, "] ", " 【"+node.TitleTag+"】", "【"+node.CourseName+"】", "【"+node.CourseContentStr+"】  >>> ", "提交状态：", "成功", lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", node.StudyProgress), "%")
		time2.Sleep(25 * time2.Second) //每隔25s进行一次学时提交
		if node.StudyProgress >= 100 {
			break //如果看完该视频则直接下一个
		}
	}
}
