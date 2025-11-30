package welearn

import (
	"fmt"
	"log"
	"math"
	"sync"
	"time"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"

	action "github.com/yatori-dev/yatori-go-core/aggregation/qingshuxuetang"
	qsxt "github.com/yatori-dev/yatori-go-core/api/qingshuxuetang"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

// 用于过滤Cqie账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "QSXT" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*qsxt.QsxtUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*qsxt.QsxtUserCache {
	var UserCaches []*qsxt.QsxtUserCache
	for _, user := range users {
		if user.AccountType == "QSXT" {
			cache := &qsxt.QsxtUserCache{Account: user.Account, Password: user.Password}
			_, err := action.QsxtLoginAction(cache) // 登录
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, err.Error())
				log.Fatal(err) //登录失败则直接退出
			}
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Green, "登录成功")
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.User, cache *qsxt.QsxtUserCache) {
	// projectList, _ := enaea.ProjectListAction(cache) //拉取项目列表
	courseList, err := action.PullCourseListAction(cache)
	if err != nil {
		panic(err)
	}
	for _, course := range courseList {
		videosLock.Add(1)
		nodeListStudy(setting, user, cache, &course) //多携程刷课
		videosLock.Done()
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
func nodeListStudy(setting config.Setting, user *config.User, userCache *qsxt.QsxtUserCache, course *action.QsxtCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.CourseName, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.CourseName, user.CoursesCustom.IncludeCourses) {
		return
	}
	if course.StudyStatusName != "在修" { //过滤非在修课程
		return
	}
	//执行刷课---------------------------------
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "【", course.CourseName, "】 ", lg.Purple, "正在学习该课程")
	//nodeList := ketangx.PullNodeListAction(userCache, course) //拉取对应课程的视频列表
	//分数不够且视频模式是开启的情况下才进行学习
	if user.CoursesCustom.VideoModel != 0 && course.CoursewareLearnGainScore < course.CoursewareLearnTotalScore {
		// 拉取视屏结点
		videoList, err := action.PullCourseNodeListAction(userCache, *course) //拉取对应课程的章节
		if err != nil {
			panic(err)
		}
		for _, video := range videoList {
			//如果超过了则直接跳过
			if course.CoursewareLearnGainScore >= course.CoursewareLearnTotalScore {
				break
			}
			//视频处理逻辑
			switch user.CoursesCustom.VideoModel {
			case 1:
				nodeSubmitTimeAction(setting, user, userCache, course, video) //刷学时
				break
			}
		}
	}

	//考试开启切分数不够的情况下才进行学习
	if user.CoursesCustom.AutoExam == 1 && course.CourseWorkGainScore < course.CourseWorkTotalScore {
		workList, err := action.PullWorkListAction(userCache, *course)
		if err != nil {
			panic(err)
		}
		for _, work := range workList {
			nodeWorkAction(setting, user, userCache, course, &work)
		}
	}
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "【"+course.CourseName+"】 ", lg.Green, "课程学习完毕")

}

// 累计学时
func nodeSubmitTimeAction(setting config.Setting, user *config.User, cache *qsxt.QsxtUserCache, course *action.QsxtCourse, node action.QsxtNode) {
	if node.NodeType == "chapter" {
		return
	}

	endTime := int(math.Ceil(float64(node.Duration / 1000)))
	if endTime == 0 {
		endTime = 300 //如果为0，则默认5分钟观看时间
	}
	//获取配置中的学时
	totalTime := node.TotalStudyDuration
	//比阈值大就直接返回
	if totalTime > endTime {
		return
	}

	startId, err2 := action.StartStudyTimeAction(cache, node)
	if err2 != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.BoldRed, "开始学习接口异常：", err2.Error())
	}

	studyTime := 0                 //当前累计学习了多久
	maxTime := endTime - totalTime //最大学习多长时间
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.Green, "正在开始学习，进度: ", fmt.Sprintf("%d/%d", studyTime+totalTime, endTime), "服务器返回信息: ", startId)
	for {
		time.Sleep(60 * time.Second)
		submitResult, err3 := action.SubmitStudyTimeAction(cache, node, startId, false)
		if err3 != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.BoldRed, "学习异常：", err3.Error())
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.Green, "学时提交成功，进度: ", fmt.Sprintf("%d/%d", studyTime+totalTime, endTime), "服务器返回信息: ", submitResult)
		studyTime += 60
		if studyTime >= maxTime {
			break
		}
	}
	submitResult, err3 := action.SubmitStudyTimeAction(cache, node, startId, true)
	if err3 != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.BoldRed, "结束学习接口异常：", err3.Error())
	}
	action.UpdateCourseScore(cache, course) //看完一个视频就更新一次分数
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+node.NodeName+"】", lg.Green, "学习完毕，进度: ", fmt.Sprintf("%d/%d", studyTime+totalTime, endTime), "服务器返回信息: ", submitResult)
}

// 作业
func nodeWorkAction(setting config.Setting, user *config.User, cache *qsxt.QsxtUserCache, course *action.QsxtCourse, work *action.QsxtWork) {
	if work.AnswerStatus != 2 { //如果答过了则直接退出
		return
	}
	submitStatus := false
	if user.CoursesCustom.ExamAutoSubmit == 0 {
		submitStatus = false
	} else if user.CoursesCustom.ExamAutoSubmit == 1 {
		submitStatus = true
	}
	submitResult, err3 := action.WriteWorkAction(cache, *work, submitStatus)
	if err3 != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+work.Title+"】", lg.BoldRed, "作业提交异常：", err3.Error())
	}
	action.UpdateCourseScore(cache, course) //答完一个作业就更新一次分数
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Default, "【"+course.CourseName+"】 ", "【"+work.Title+"】", lg.Green, "作业已自动完成,服务器返回信息: ", submitResult)
}
