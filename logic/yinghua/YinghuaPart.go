package yinghua

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"sync"
	time2 "time"
	"yatori-go-console/config"
	"yatori-go-console/global"
	utils2 "yatori-go-console/utils"
	modelLog "yatori-go-console/utils/log"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/external"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

// 用于过滤英华账号
func FilterAccount(configData *config.JSONDataForConfig) []config.User {
	var users []config.User //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "YINGHUA" {
			users = append(users, user)
		}
	}
	return users
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.User, userCaches []*yinghuaApi.YingHuaUserCache) {
	var usersLock sync.WaitGroup //用户锁
	//开始刷课
	for i, user := range userCaches {
		usersLock.Add(1)
		go func() {
			defer usersLock.Done()
			userBlock(setting, &users[i], user)
		}()

	}
	usersLock.Wait()

}

// ipProxy 代理IP设定
func ipProxy(user config.User, cache *yinghuaApi.YingHuaUserCache) {
	for {
		if user.IsProxy == 1 {
			//获取随机IP值
			cache.IpProxySW = true
			cache.ProxyIP = utils2.RandProxyStr()
		}
		time2.Sleep(10 * time2.Second)
	}
}

// 用户登录模块
func UserLoginOperation(users []config.User) []*yinghuaApi.YingHuaUserCache {
	var UserCaches []*yinghuaApi.YingHuaUserCache
	for _, user := range users {
		if user.AccountType == "YINGHUA" {
			cache := &yinghuaApi.YingHuaUserCache{PreUrl: user.URL, Account: user.Account, Password: user.Password}
			go ipProxy(user, cache)                   // 携程定时变换代理地址
			err1 := yinghua.YingHuaLoginAction(cache) // 登录
			if err1 != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, err1.Error())
				log.Fatal(err1) //登录失败则直接退出
			}
			go keepAliveLogin(cache) //携程保活
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.User, cache *yinghuaApi.YingHuaUserCache) {
	list, _ := yinghua.CourseListAction(cache) //拉取课程列表
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, cache.Account, lg.Default, "] ", lg.Purple, "正在定位上次学习位置...")
	var nodesLock sync.WaitGroup //节点锁
	for _, item := range list {  //遍历所有待刷视频
		nodesLock.Add(1)
		go func() {
			defer nodesLock.Done()
			nodeListStudy(setting, user, cache, &item) //多携程刷课
			//如果是暴力模式，等结束后再进行一次去红模式
			if user.CoursesCustom.VideoModel == 2 {

				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, user.Account, lg.Default, "] ", lg.Yellow, "暴力模式执行完毕，正在自动执行去红模式(需延迟一分钟执行)...")
				time2.Sleep(time2.Minute)
				resUser := *user                               //改为去红模式
				resUser.CoursesCustom.VideoModel = 3           //标记为去红模式并启动
				nodeListStudy(setting, &resUser, cache, &item) //递归执行去红模式

			}

		}()
	}
	nodesLock.Wait() //等待所有节点结束
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
}

// 用于登录保活
func keepAliveLogin(UserCache *yinghuaApi.YingHuaUserCache) {
	ticker := time2.NewTicker(time2.Minute * 5)
	//ticker := time2.NewTicker(time2.Second * 5)
	for {
		select {
		case <-ticker.C:
			api := yinghuaApi.KeepAliveApi(*UserCache, 8)
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr["YINGHUA"]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.DarkGray, "登录心跳保活状态：", api)
		}
	}
}

// 章节节点的抽离函数
func nodeListStudy(setting config.Setting, user *config.User, userCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(course.Name, user.CoursesCustom.ExcludeCourses) {
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(course.Name, user.CoursesCustom.IncludeCourses) {
		return
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "正在学习课程：", lg.Yellow, " 【"+course.Name+"】 ")

	//如果课程时间未到开课时间则直接return
	//{"_code":9,"status":false,"msg":"课程还未开始!","result":{}}
	if time2.Now().Before(course.StartDate) {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", " 【", course.Name, "】 >>> ", lg.Red, "该课程还未开始已自动跳过")
		return
	}
	//执行刷课---------------------------------
	nodeList, err := yinghua.VideosListAction(userCache, *course) //拉取对应课程的视频列表
	if err != nil {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", " 【", course.Name, "】 >>> ", lg.Red, "拉取视屏列表失败", err.Error())
	}
	var videosLock sync.WaitGroup //视频节点锁
	// 提交学时
	redAns := 0 //标红记录统计（标红为0才退出去红模式
	for _, node := range nodeList {
		//视频处理逻辑
		switch user.CoursesCustom.VideoModel { //根据视频模式进行刷课
		case 1:
			videoAction(setting, user, userCache, course, node) //普通模式
			break
		case 2:
			videosLock.Add(1)
			go func() {
				defer videosLock.Done()
				videoVioLenceAction(setting, user, userCache, course, node) //暴力模式
			}()
			break
		case 3:
			if node.ErrorMessage != "检测到可能使用并行播放刷课" { //统计标红
				redAns++
			}
			videoBadRedAction(setting, user, userCache, course, node) //去红模式
			break

		}
		//如果还有标红的则再运行一遍
		if user.CoursesCustom.VideoModel == 3 && redAns != 0 {
			nodeListStudy(setting, user, userCache, course)
			return
		}
		//作业处理逻辑
		workAction(setting, user, userCache, course, node)
		//考试处理逻辑
		examAction(setting, user, userCache, course, node)
		if setting.BasicSetting.LogModel == 1 {
			action, err := yinghua.CourseDetailAction(userCache, course.Id)
			if err != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Default, " 【"+course.Name+"】 ", lg.Red, "拉取课程进度失败", err.Error())
				break
			}
			modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Default, " 【"+course.Name+"】 ", "视频学习进度：", strconv.Itoa(action.VideoLearned), "/", strconv.Itoa(action.VideoCount), " ", "课程总学习进度：", fmt.Sprintf("%.2f", action.Progress*100), "%")
		}
	}
	videosLock.Wait() //等待所有视频刷完
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 1, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Green, "课程", " 【"+course.Name+"】 ", "学习完毕")
}

// videoAction 刷视频逻辑抽离
func videoAction(setting config.Setting, user *config.User, UserCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse, node yinghua.YingHuaNode) {
	if !node.TabVideo { //过滤非视频节点
		return
	}
	if int(node.Progress) == 100 { //过滤看完了的视屏
		return
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在学习视频：", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】")
	time := node.ViewedDuration //设置当前观看时间为最后看视频的时间
	studyId := "0"              //服务器端分配的学习ID
	for {
		time += 5
		if node.Progress == 100 {
			modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", " ", lg.Blue, "学习完毕")
			break //如果看完了，也就是进度为100那么直接跳过
		}
		//提交学时
		sub, err := yinghua.SubmitStudyTimeAction(UserCache, node.Id, studyId, time)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), `[`, UserCache.Account, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
		}
		//超时重登检测
		yinghua.LoginTimeoutAfreshAction(UserCache, sub)
		lg.Print(lg.DEBUG, "---", node.Id, sub)
		//如果提交学时不成功
		msgVal := gojsonq.New().JSONString(sub).Find("msg")
		msg, ok := msgVal.(string)
		if !ok || msg == "" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Red, "提交状态异常，msg 字段为空或格式错误", sub)
			time2.Sleep(10 * time2.Second)
			continue
		}
		if msg != "提交学时成功!" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", "提交状态：", lg.Red, sub)
			//{"_code":9,"status":false,"msg":"该课程解锁时间【2024-11-14 12:00:00】未到!","result":{}}，如果未到解锁时间则跳过
			reg1 := regexp.MustCompile(`该课程解锁时间【[^【]*】未到!`)
			if reg1.MatchString(msg) {
				modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", lg.Red, "该课程未到解锁时间已自动跳过")
				break
			}
			time2.Sleep(10 * time2.Second)
			continue
		}
		//打印日志部分
		studyIdVal := gojsonq.New().JSONString(sub).Find("result.data.studyId")
		if idFloat, ok := studyIdVal.(float64); ok {
			studyId = strconv.Itoa(int(idFloat))
		}
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", "提交状态：", lg.Green, msg, lg.Default, " ", "观看时间：", strconv.Itoa(time)+"/"+strconv.Itoa(node.VideoDuration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(time)/float32(node.VideoDuration)*100), "%")
		time2.Sleep(5 * time2.Second)
		if time >= node.VideoDuration {
			break //如果看完该视频则直接下一个
		}
	}
}

// videoAction 刷视频逻辑抽离(暴力模式)
func videoVioLenceAction(setting config.Setting, user *config.User, UserCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse, node yinghua.YingHuaNode) {
	if !node.TabVideo { //过滤非视频节点
		return
	}
	if int(node.Progress) == 100 { //过滤看完了的视屏
		return
	}

	if int(node.Progress) == 100 { //过滤看完了的视屏
		return
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在学习视频：", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ")
	time := node.ViewedDuration //设置当前观看时间为最后看视频的时间
	studyId := "0"              //服务器端分配的学习ID
	for {
		time += 5
		if node.Progress == 100 {
			modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", " ", lg.Blue, "学习完毕")
			break //如果看完了，也就是进度为100那么直接跳过
		}
		//提交学时
		sub, err := yinghua.SubmitStudyTimeAction(UserCache, node.Id, studyId, time)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), `[`, UserCache.Account, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
		}
		//超时重登检测
		yinghua.LoginTimeoutAfreshAction(UserCache, sub)
		lg.Print(lg.DEBUG, "---", node.Id, sub)
		//如果提交学时不成功
		msgVal := gojsonq.New().JSONString(sub).Find("msg")
		msg, ok := msgVal.(string)
		if !ok || msg == "" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Red, "提交状态异常，msg 字段为空或格式错误")
			time2.Sleep(10 * time2.Second)
			continue
		}

		if msg != "提交学时成功!" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", "提交状态：", lg.Red, sub)

			reg1 := regexp.MustCompile(`该课程解锁时间【[^【]*】未到!`)
			if reg1.MatchString(msg) {
				modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", lg.Red, "该课程未到解锁时间已自动跳过")
				break
			}
			time2.Sleep(10 * time2.Second)
			continue
		}
		//打印日志部分
		studyId = strconv.Itoa(int(gojsonq.New().JSONString(sub).Find("result.data.studyId").(float64)))
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", "提交状态：", lg.Green, msg, lg.Default, " ", "观看时间：", strconv.Itoa(time)+"/"+strconv.Itoa(node.VideoDuration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(time)/float32(node.VideoDuration)*100), "%")

		time2.Sleep(5 * time2.Second)
		if time >= node.VideoDuration {
			break //如果看完该视频则直接下一个
		}
	}
}

// videoBadRedAction 去红模式
func videoBadRedAction(setting config.Setting, user *config.User, UserCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse, node yinghua.YingHuaNode) {
	if !node.TabVideo { //过滤非视频节点
		return
	}
	//除去不需要消红的视屏
	if node.ErrorMessage != "检测到可能使用并行播放刷课" {
		return
	}
	modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", lg.Yellow, "正在消红视频：", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ")
	time := node.ViewedDuration //设置当前观看时间为最后看视频的时间

	studyId := "0" //服务器端分配的学习ID
	for {
		//提交学时
		sub, err := yinghua.SubmitStudyTimeAction(UserCache, node.Id, studyId, time)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), `[`, UserCache.Account, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
		}
		//超时重登检测
		yinghua.LoginTimeoutAfreshAction(UserCache, sub)
		lg.Print(lg.DEBUG, "---", node.Id, sub)
		//如果提交学时不成功
		msgVal := gojsonq.New().JSONString(sub).Find("msg")
		msg, ok := msgVal.(string)
		if !ok || msg == "" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", lg.Red, "提交状态异常，msg 字段为空或格式错误", sub)
			time2.Sleep(10 * time2.Second)
			continue
		}
		if msg != "提交学时成功!" {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", "提交状态：", lg.Red, sub)
			//{"_code":9,"status":false,"msg":"该课程解锁时间【2024-11-14 12:00:00】未到!","result":{}}，如果未到解锁时间则跳过
			reg1 := regexp.MustCompile(`该课程解锁时间【[^【]*】未到!`)
			if reg1.MatchString(msg) {
				modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", lg.Red, "该课程未到解锁时间已自动跳过")
				break
			}
			time2.Sleep(10 * time2.Second)
			continue
		}
		//打印日志部分
		studyIdVal := gojsonq.New().JSONString(sub).Find("result.data.studyId")
		if idFloat, ok := studyIdVal.(float64); ok {
			studyId = strconv.Itoa(int(idFloat))
		}
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, UserCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 >>> ", lg.Red, " 去红模式 ", lg.Default, "提交状态：", lg.Green, msg, lg.Default, " ", "观看时间：", strconv.Itoa(time)+"/"+strconv.Itoa(node.VideoDuration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(time)/float32(node.VideoDuration)*100), "%")
		time2.Sleep(8 * time2.Second) //隔8s下一个去红
		break                         //因为是去红模式，所以直接退出
	}
}

// workAction 作业处理逻辑
func workAction(setting config.Setting, user *config.User, userCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse, node yinghua.YingHuaNode) {
	if user.CoursesCustom.AutoExam == 0 { //是否打开了自动考试开关
		return
	}
	if !node.TabWork { //过滤非作业节点
		return
	}
	if user.CoursesCustom.AutoExam == 1 {
		//检测AI可用性
		err := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), lg.BoldRed, fmt.Sprintf("<%s>", setting.AiSetting.AiType), "AI不可用，错误信息："+err.Error())
			os.Exit(0)
		}
	}

	if user.CoursesCustom.AutoExam == 2 {
		err := external.CheckApiQueRequest(setting.ApiQueSetting.Url, 3, nil)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), lg.BoldRed, "外置题库不可用，错误信息："+err.Error())
			os.Exit(0)
		}
	}
	//获取作业详细信息
	detailAction, _ := yinghua.WorkDetailAction(userCache, node.Id)
	////{"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
	if len(detailAction) == 0 { //过滤没有作业内容的
		return
	}
	if user.CoursesCustom.AutoExam == 1 {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ", lg.Yellow, "正在AI自动写章节作业...")
	} else if user.CoursesCustom.AutoExam == 2 {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ", lg.Yellow, "正在外置题库自动写章节作业...")
	}

	//开始写作业
	for _, work := range detailAction {
		var err error
		switch user.CoursesCustom.AutoExam {
		case 1:
			err = yinghua.StartWorkAction(userCache, work, setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType, user.CoursesCustom.ExamAutoSubmit)
			break
		case 2:
			err = yinghua.StartWorkForExternalAction(userCache, setting.ApiQueSetting.Url, work, user.CoursesCustom.ExamAutoSubmit)
			break
		}

		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", lg.BoldRed, "该章节作业无法正常执行，服务器返回信息：", err.Error())
			continue
		}
		if user.CoursesCustom.ExamAutoSubmit == 1 {
			//打印最终分数
			s, err1 := yinghua.WorkedFinallyScoreAction(userCache, work)
			if err1 != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", lg.BoldRed, err1)
				continue
			}
			if user.CoursesCustom.AutoExam == 1 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "章节作业AI答题完毕，最高分：", s, "分", " 试卷总分：", fmt.Sprintf("%.2f分", work.Score))
			} else if user.CoursesCustom.AutoExam == 2 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", " 【", node.Name, "】", lg.Green, "章节作业外置题库答题完毕，最高分：", s, "分", " 试卷总分：", fmt.Sprintf("%.2f分", work.Score))
			}

		} else {
			if user.CoursesCustom.AutoExam == 1 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "AI考试完毕,请自行前往主页提交试卷")
			} else if user.CoursesCustom.AutoExam == 2 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "外置题库考试完毕,请自行前往主页提交试卷")
			}
		}
	}

}

// examAction 考试处理逻辑
func examAction(setting config.Setting, user *config.User, userCache *yinghuaApi.YingHuaUserCache, course *yinghua.YingHuaCourse, node yinghua.YingHuaNode) {
	if user.CoursesCustom.AutoExam == 0 { //是否打开了自动考试开关
		return
	}
	if !node.TabExam { //过滤非考试节点
		return
	}

	if user.CoursesCustom.AutoExam == 1 {
		//检测AI可用性
		err := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), lg.BoldRed, fmt.Sprintf("<%s>", setting.AiSetting.AiType), "AI不可用，错误信息："+err.Error())
			os.Exit(0)
		}
	}

	if user.CoursesCustom.AutoExam == 2 {
		err := external.CheckApiQueRequest(setting.ApiQueSetting.Url, 3, nil)
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), lg.BoldRed, "外置题库不可用，错误信息："+err.Error())
			os.Exit(0)
		}
	}

	//获取作业详细信息
	detailAction, _ := yinghua.ExamDetailAction(userCache, node.Id)
	////{"_code":9,"status":false,"msg":"考试测试时间还未开始","result":{}}
	if len(detailAction) == 0 { //过滤没有考试内容的
		return
	}
	//开始考试
	if user.CoursesCustom.AutoExam == 1 {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ", lg.Yellow, "正在AI自动考试...")
	} else if user.CoursesCustom.AutoExam == 2 {
		modelLog.ModelPrint(setting.BasicSetting.LogModel == 0, lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", lg.Default, fmt.Sprintf("【%s】", course.Name), "【"+node.Name+"】 ", lg.Yellow, "正在外置题库自动考试...")
	}

	for _, exam := range detailAction {
		//err := yinghua.StartExamAction(userCache, exam, setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType, user.CoursesCustom.ExamAutoSubmit)
		var err error
		switch user.CoursesCustom.AutoExam {
		case 1:
			err = yinghua.StartExamAction(userCache, exam, setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType, user.CoursesCustom.ExamAutoSubmit)
			break
		case 2:
			err = yinghua.StartExamForExternalAction(userCache, exam, setting.ApiQueSetting.Url, user.CoursesCustom.ExamAutoSubmit)
			break
		}
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", lg.BoldRed, "该考试无法正常执行，服务器返回信息：", err.Error())
			continue
		}

		if user.CoursesCustom.ExamAutoSubmit == 1 {
			//打印最终分数
			s, err1 := yinghua.ExamFinallyScoreAction(userCache, exam)
			if err1 != nil {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】 ", lg.BoldRed, err1.Error())
				continue
			}
			if user.CoursesCustom.AutoExam == 1 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "AI考试完毕,最终分：", s, "分", " 试卷总分：", fmt.Sprintf("%.2f分", exam.Score))
			} else if user.CoursesCustom.AutoExam == 2 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "外置题库考试完毕,最终分：", s, "分", " 试卷总分：", fmt.Sprintf("%.2f分", exam.Score))
			}
		} else {
			if user.CoursesCustom.AutoExam == 1 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "AI考试完毕,请自行前往主页提交试卷")
			} else if user.CoursesCustom.AutoExam == 2 {
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, userCache.Account, lg.Default, "] ", fmt.Sprintf("【%s】", course.Name), "【", node.Name, "】", lg.Green, "外置题库考试完毕,请自行前往主页提交试卷")
			}

		}

	}
}
