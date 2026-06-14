package haiqikeji

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/aggregation/ketangx"
	hqkjApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	ketangxApi "github.com/yatori-dev/yatori-go-core/api/ketangx"

	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

var usersLock sync.WaitGroup //用户锁

// 用于过滤Hqkj账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集账号
	for _, user := range configData.Users {
		if user.AccountType == "HQKJ" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*hqkjApi.HqkjUserCache) {
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)

	}
	usersLock.Wait()
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*hqkjApi.HqkjUserCache {
	var UserCaches []*hqkjApi.HqkjUserCache
	for _, user := range users {
		if user.AccountType == "HQKJ" {
			cache := &hqkjApi.HqkjUserCache{PreUrl: user.URL, Account: user.Account, Password: user.Password}
			err := haiqikeji.HqkjLoginAction(cache) // 登录
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

func userBlock(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache) {
	courseList, err := haiqikeji.HqkjCourseListAction(cache)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]))
		return
	}
	var coursesLock sync.WaitGroup //视频锁
	for _, course := range courseList {
		//if course.Offline != 1 { //结束的课程过滤掉
		//	continue
		//}
		//if course.StartDate.After(time.Now()) || course.EndDate.Before(time.Now()) { //过滤掉过时课程
		coursesLock.Add(1)
		go func() {
			nodeListStudy(setting, user, cache, &course) //多携程刷课
			coursesLock.Done()
		}()
	}
	coursesLock.Wait() //等待课程刷完

	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	//如果开启了邮箱通知
	if setting.EmailInform.Sw == 1 && len(user.InformEmails) > 0 {
		utils2.SendMail(setting.EmailInform.SMTPHost, setting.EmailInform.SMTPPort, setting.EmailInform.UserName, setting.EmailInform.Password, user.InformEmails, fmt.Sprintf("账号：[%s]</br>平台：[%s]</br>通知：所有课程已执行完毕", user.Account, global.AccountTypeStr[user.AccountType]))
	}
	if setting.BasicSetting.CompletionTone == 1 { //如果声音提示开启，那么播放
		soundMut.Lock()
		utils2.PlayNoticeSound() //播放提示音
		soundMut.Unlock()
	}
	usersLock.Done()
}

// 章节节点的抽离函数
func nodeListStudy(setting config.Setting, user *config.User, userCache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.Name, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.Name, user.CoursesCustom.IncludeCourses) {
		return
	}
	if course.StartDate.After(time.Now()) || course.EndDate.Before(time.Now()) { //过滤掉过时课程
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】 >>> ", lg.Default, " ", "该课程的起止时间为：", lg.Red, fmt.Sprintf("[%s ~ %s] ", course.StartDate.Format("2006-01-02"), course.EndDate.Format("2006-01-02")), lg.Yellow, "因未在开课时间内，已跳过该课程")
		return
	}
	//执行刷课---------------------------------
	nodeList, err := haiqikeji.HqkjNodeListAction(userCache, *course) //拉取对应课程的视频列表
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]))
		return
	}
	//nodeList := ketangx.PullNodeListAction(userCache, course) //拉取对应课程的视频列表
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, " 【"+course.Name+"】 ")
	//视频处理逻辑
	switch user.CoursesCustom.VideoModel {
	case 1:
		normalModeAction(setting, user, userCache, course, nodeList)
		break
	case 2:
		fastModeAction(setting, user, userCache, course, nodeList)
		break
	}

	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", "【"+course.Name+"】 ", "学习完毕")

}

// 普通模式
func normalModeAction(setting config.Setting, user *config.User, UserCache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, nodeList []haiqikeji.HqkjNode) {
	// 提交学时
	for _, node := range nodeList {
		if node.TabVideo <= 0 { //过滤没有视频的
			continue
		}
		progress, err := haiqikeji.HqkjGetNodeProgressAction(UserCache, node)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取进度错误")
			continue
		}
		//检查是否看完
		if progress >= 100 {
			continue
		}

		sessionId, err := haiqikeji.HqkjStartStudyAction(UserCache, node)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取学习sessionId失败：", err.Error())
			return
		}
		nowTime := int(float64(progress) * 0.01 * float64(node.VideoDuration)) //计算当前学习到的时间点
		stopTime := 30                                                         //暂停时间
		time.Sleep(time.Duration(stopTime) * time.Second)
		for {
			nowAddV := stopTime //用于临时寄存减少量

			//如果添加的进度大于视频长度那么就直接等于视频长度
			if nowTime+stopTime > node.VideoDuration {
				nowAddV = node.VideoDuration - nowTime
			}
			nowTime += nowAddV //添加时间

			//计算当前视频进度
			submitProgress := int(float64(nowTime) / float64(node.VideoDuration) * 100)
			submitResult, err := haiqikeji.HqkjSubmitStudyTimeAction(UserCache, node, sessionId, submitProgress)
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "提交学时失败：", err.Error())
			}
			msg := gojsonq.New().JSONString(submitResult).Find("msg")
			if msg != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", "提交状态：", lg.Green, gojsonq.New().JSONString(submitResult).Find("msg").(string), lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", float64(submitProgress)), "%")
			} else {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", "提交状态：", lg.Green, gojsonq.New().JSONString(submitResult).Find("msg"), lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", float64(submitProgress)), "%")
			}
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取进度错误", err.Error())
				return
			}

			//如果进度达标了那就直接退出
			if submitProgress >= 100 {
				//保存结果
				endResult, err := haiqikeji.HqkjEndStudyAction(UserCache, sessionId)
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Default, " ", "服务器返回：", endResult)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), err.Error())
					break
				}
				//拉取进度看看到底成功保存没，没有那么就重新开始刷
				progress, err = haiqikeji.HqkjGetNodeProgressAction(UserCache, node) //拉取进度
				//对比进度是否小于当前，如果是的话，说明提交失败，那么就要累加失败次数
				if progress < 100 {
					sessionId, err = haiqikeji.HqkjStartStudyAction(UserCache, node)
					if err != nil {
						lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取学习sessionId失败：", err.Error())
						break //如果获取新sessionId失败则直接退出不执行这个章节了
					}
					time.Sleep(time.Duration(stopTime) * time.Second)
					nowTime = int(float64(node.VideoDuration) * 0.01 * float64(progress)) //恢复进度
					continue
				}
				break
			}
			time.Sleep(time.Duration(stopTime) * time.Second) //间隔为30s
		}
	}
}

// 快速模式
func fastModeAction(setting config.Setting, user *config.User, UserCache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, nodeList []haiqikeji.HqkjNode) {
	// 提交学时
	var videosLock sync.WaitGroup
	for _, node := range nodeList {
		videosLock.Add(1)
		go func(node haiqikeji.HqkjNode) {
			progress, err := haiqikeji.HqkjGetNodeProgressAction(UserCache, node)
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取进度错误")
				videosLock.Done()
				return
			}
			//检查是否看完
			//if progress >= 100 {
			//	videosLock.Done()
			//	return
			//}

			var submitResult string
			//这里采用提交学后进行检查，防止提交的进度没有记录问题
			for {
				sessionId, err := haiqikeji.HqkjStartStudyAction(UserCache, node)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取学习sessionId失败：", err.Error())
					videosLock.Done()
					return
				}
				time.Sleep(30 * time.Second)
				submitResult, err = haiqikeji.HqkjSubmitStudyTimeAction(UserCache, node, sessionId, 100)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "提交学时失败：", err.Error())
					return
				}
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", "提交状态：", lg.Green, gojsonq.New().JSONString(submitResult).Find("msg").(string), lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", float64(100)), "%")
				//保存结果
				endResult, err := haiqikeji.HqkjEndStudyAction(UserCache, sessionId)
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Default, " ", "服务器返回：", endResult)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), err.Error())
					break
				}
				progress, err = haiqikeji.HqkjGetNodeProgressAction(UserCache, node)
				if err != nil {
					lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "拉取进度错误", err.Error())
					videosLock.Done()
					return
				}
				//检查是否看完
				if progress >= 100 {
					break
				}
				time.Sleep(30 * time.Second)
			}
			videosLock.Done()
		}(node)

	}
	videosLock.Wait()
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
}
