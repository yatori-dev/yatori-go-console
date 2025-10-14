package xuexitong

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"yatori-go-console/config"
	utils2 "yatori-go-console/utils"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/aggregation/xuexitong/point"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/external"
	"github.com/yatori-dev/yatori-go-core/utils"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
)

var nodesLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup //用户锁

// 用于过滤学习通账号
func FilterAccount(configData *config.JSONDataForConfig) []config.Users {
	var users []config.Users //用于收集英华账号
	for _, user := range configData.Users {
		if user.AccountType == "XUEXITONG" {
			users = append(users, user)
		}
	}
	return users
}

// 用户登录模块
func UserLoginOperation(users []config.Users) []*xuexitongApi.XueXiTUserCache {
	var UserCaches []*xuexitongApi.XueXiTUserCache
	for _, user := range users {
		if user.AccountType == "XUEXITONG" {
			cache := &xuexitongApi.XueXiTUserCache{Name: user.Account, Password: user.Password}
			//设置代理IP
			if user.IsProxy == 1 {
				cache.IpProxySW = true
				cache.ProxyIP = "http://" + utils2.RandProxyStr()
			}
			loginError := xuexitong.XueXiTLoginAction(cache) // 登录
			if loginError != nil {
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.White, "] ", lg.Red, loginError.Error())
				os.Exit(0) //登录失败直接退出
			}
			// go keepAliveLogin(cache) //携程保活
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.Users, userCaches []*xuexitongApi.XueXiTUserCache) {
	for i, _ := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], userCaches[i])
	}
	usersLock.Wait()
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

// 用于模式3的
var model3Caches []xuexitongApi.XueXiTUserCache

func userBlock(setting config.Setting, user *config.Users, cache *xuexitongApi.XueXiTUserCache) {
	// list, err := xuexitong.XueXiTCourseDetailForCourseIdAction(cache, "261619055656961")
	courseList, err := xuexitong.XueXiTPullCourseAction(cache)
	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", lg.Red, "拉取课程失败")
		log.Fatal(err)
	}
	//如果是多节点模式
	if user.CoursesCustom.VideoModel == 3 {

		for i := 0; i < 3; i++ {
			if i == 0 {
				model3Caches = append(model3Caches, *cache)
				continue
			}
			model3Caches = append(model3Caches, *cache)
			xuexitong.ReLogin(&model3Caches[i])
			time.Sleep(5 * time.Second) //隔一下，避免登录太快
		}

	}
	for _, course := range courseList {
		nodesLock.Add(1)
		// fmt.Println(course)
		if user.CoursesCustom.VideoModel == 1 {
			nodeListStudy(setting, user, cache, &course)
		} else {
			go func() {
				nodeListStudy(setting, user, cache, &course)
			}()
		}

	}
	nodesLock.Wait()
	lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	if setting.BasicSetting.CompletionTone == 1 { //如果声音提示开启，那么播放
		soundMut.Lock()
		utils2.PlayNoticeSound() //播放提示音
		soundMut.Unlock()
	}
	usersLock.Done()
}

// 课程节点执行
func nodeListStudy(setting config.Setting, user *config.Users, userCache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse) {
	//过滤课程---------------------------------
	//排除指定课程
	if len(user.CoursesCustom.ExcludeCourses) != 0 && config.CmpCourse(courseItem.CourseName, user.CoursesCustom.ExcludeCourses) {
		nodesLock.Done()
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(courseItem.CourseName, user.CoursesCustom.IncludeCourses) {
		nodesLock.Done()
		return
	}
	//如果课程还未开课则直接退出
	if !courseItem.IsStart {
		nodesLock.Done()
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Blue, "该课程还未开课，已自动跳过该课程")
		return
	}
	//如果该课程已经结束
	if courseItem.State == 1 {
		nodesLock.Done()
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Blue, "该课程已经结束，已自动跳过该课程")
		return
	}

	key, _ := strconv.Atoi(courseItem.Key)
	action, _, err := xuexitong.PullCourseChapterAction(userCache, courseItem.Cpi, key) //获取对应章节信息

	if err != nil {
		if strings.Contains(err.Error(), "课程章节为空") {
			lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该课程章节为空已自动跳过")
			nodesLock.Done()
			return
		}
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "拉取章节信息接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())

		nodesLock.Done()
		return
		//log.Fatal()
	}
	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, "获取课程章节成功 (共 ", lg.Yellow, strconv.Itoa(len(action.Knowledge)), lg.Default, " 个) ")

	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(courseItem.CourseID)
	userId, _ := strconv.Atoi(userCache.UserID)
	// 检测节点完成情况
	pointAction, err := xuexitong.ChapterFetchPointAction(userCache, nodes, &action, key, userId, courseItem.Cpi, courseId)
	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "探测节点完成情况接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())
		//log.Fatal()
		nodesLock.Done()
		return
	}
	var isFinished = func(index int) bool {
		if index < 0 || index >= len(pointAction.Knowledge) {
			return false
		}
		i := pointAction.Knowledge[index]
		if i.PointTotal == 0 && i.PointFinished == 0 {
			//如果是0任务点，则直接浏览一遍主页面即可完成任务，不必继续下去
			err2 := xuexitong.EnterChapterForwardCallAction(userCache, strconv.Itoa(courseId), strconv.Itoa(key), strconv.Itoa(pointAction.Knowledge[index].ID), strconv.Itoa(courseItem.Cpi))
			if err2 != nil {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "零任务点遍历失败。返回信息：", err2.Error())
			}
		}
		return i.PointTotal >= 0 && i.PointTotal == i.PointFinished
	}

	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Purple, "正在学习该课程")
	nodeLock := sync.WaitGroup{}

	//初始化模式3用的队列
	queue := make(chan int, len(model3Caches))
	for i := 0; i < len(model3Caches); i++ {
		queue <- i
	}
	//遍历结点
	for index := range nodes {
		if isFinished(index) { //如果完成了的那么直接跳过
			continue
		}

		if user.CoursesCustom.VideoModel == 3 {

			//如果队列为空则堵塞
			for {
				if len(queue) != 0 {
					break
				}
			}
			// 从队列中取一个资源（如果空则会自动阻塞）
			idx := <-queue
			go func(idx int, index int) {
				nodeLock.Add(1)
				//分配Cookie
				nodeRun(setting, user, &model3Caches[idx], courseItem, pointAction, action, nodes, index, key, courseId)
				// 执行完毕后归还资源
				queue <- idx
				nodeLock.Done()
			}(idx, index)
		} else {
			nodeRun(setting, user, userCache, courseItem, pointAction, action, nodes, index, key, courseId)
		}
	}
	nodeLock.Wait() //等待节点刷完
	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Purple, "课程学习完毕")
	nodesLock.Done()
}

// 任务点分流运行
func nodeRun(setting config.Setting, user *config.Users, userCache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse,
	pointAction xuexitong.ChaptersList, action xuexitong.ChaptersList, nodes []int, index int, key int, courseId int) {
	_, fetchCards, err1 := xuexitong.ChapterFetchCardsAction(userCache, &action, nodes, index, courseId, key, courseItem.Cpi)
	if err1 != nil {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "无法正常拉取卡片信息，请联系作者查明情况,报错信息：", err1.Error())
		//log.Fatal(err1)
		nodesLock.Done()
		return
	}
	videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, bbsDTOs := entity.ParsePointDto(fetchCards)
	if videoDTOs == nil && workDTOs == nil && documentDTOs == nil && hyperlinkDTOs == nil && liveDTOs == nil && bbsDTOs == nil {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, `[`, pointAction.Knowledge[index].Name, `] `, lg.Blue, "课程对应章节无任何任务节点，已自动跳过")
		return
	}
	// 视屏类型
	if videoDTOs != nil && user.CoursesCustom.VideoModel != 0 {
		for _, videoDTO := range videoDTOs {
			card, enc, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				log.Fatal(err2)
			}
			videoDTO.AttachmentsDetection(card)

			if !videoDTO.IsJob {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, `[`, pointAction.Knowledge[index].Name, `] `, lg.Blue, "该视屏非任务点或已完成，已自动跳过")
				continue
			}
			videoDTO.Enc = enc                                        //赋值enc值
			if videoDTO.IsPassed == true && videoDTO.IsJob == false { //如果已经通过了，那么直接跳过
				continue
			} else if videoDTO.IsPassed == false && videoDTO.Attachment == nil && videoDTO.JobID == "" && videoDTO.Duration <= videoDTO.PlayTime { //非任务点如果完成了
				continue
			}
			switch user.CoursesCustom.VideoModel {
			case 1:
				ExecuteVideo2(userCache, courseItem, pointAction.Knowledge[index], &videoDTO, key, courseItem.Cpi) //普通模式
			case 2:
				ExecuteVideo2(userCache, courseItem, pointAction.Knowledge[index], &videoDTO, key, courseItem.Cpi) //多课程模式
			case 3:
				ExecuteVideo2(userCache, courseItem, pointAction.Knowledge[index], &videoDTO, key, courseItem.Cpi) //多任务点模式
			}

			time.Sleep(10 * time.Second)
		}
	}
	// 文档类型
	if documentDTOs != nil {
		for _, documentDTO := range documentDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, documentDTO.KnowledgeID, documentDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				log.Fatal(err2)
			}
			documentDTO.AttachmentsDetection(card)
			//如果不是任务或者说该任务已完成，那么直接跳过
			if !documentDTO.IsJob {
				continue
			}
			ExecuteDocument(userCache, courseItem, pointAction.Knowledge[index], &documentDTO)
			time.Sleep(5 * time.Second)
		}
	}

	//作业刷取
	if workDTOs != nil && user.CoursesCustom.AutoExam != 0 {

		if user.CoursesCustom.AutoExam == 1 { //检测AI可用性
			err2 := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType)
			if err2 != nil {
				lg.Print(lg.INFO, lg.BoldRed, "<"+setting.AiSetting.AiType+">", "AI不可用，错误信息："+err2.Error())
				os.Exit(0)
			}
		} else if user.CoursesCustom.AutoExam == 2 { // 检测外挂题库可用性
			err2 := external.CheckApiQueRequest(setting.ApiQueSetting.Url, 3, nil)
			if err2 != nil {
				lg.Print(lg.INFO, lg.BoldRed, "外挂题库不可用，错误信息："+err2.Error())
				os.Exit(0)
			}
		}

		for _, workDTO := range workDTOs {
			//以手机端拉取章节卡片数据
			mobileCard, _, _ := xuexitong.PageMobileChapterCardAction(userCache, key, courseId, workDTO.KnowledgeID, workDTO.CardIndex, courseItem.Cpi)
			flag, _ := workDTO.AttachmentsDetection(mobileCard)
			questionAction, err2 := xuexitong.ParseWorkQuestionAction(userCache, &workDTO)
			if err2 != nil && strings.Contains(err2.Error(), "已截止，不能作答") {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", questionAction.Title, "】", lg.Yellow, "该试卷已到截止时间，已自动跳过")
				continue
			}
			if !flag {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", questionAction.Title, "】", lg.Green, "该作业已完成，已自动跳过")
				continue
			}
			if len(questionAction.Short) == 0 && len(questionAction.Choice) == 0 &&
				len(questionAction.Judge) == 0 && len(questionAction.Fill) == 0 &&
				len(questionAction.TermExplanation) == 0 && len(questionAction.Essay) == 0 {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", questionAction.Title, "】", lg.Yellow, "该作业任务点无题目，已自动跳过")
				continue
			}
			//if !strings.Contains(questionAction.Title, "2.1小节测验") {
			//	continue
			//}
			WorkAction(userCache, user, setting, courseItem, pointAction.Knowledge[index], questionAction)
		}
	}

	//外链任务点刷取
	if hyperlinkDTOs != nil {
		for _, hyperlinkDTO := range hyperlinkDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, hyperlinkDTO.KnowledgeID, hyperlinkDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				log.Fatal(err2)
			}
			hyperlinkDTO.AttachmentsDetection(card)

			ExecuteHyperlink(userCache, courseItem, pointAction.Knowledge[index], &hyperlinkDTO)
			time.Sleep(5 * time.Second)
		}
	}
	// 直播任务点刷取
	if liveDTOs != nil {
		for _, liveDTO := range liveDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, liveDTO.KnowledgeID, liveDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				log.Fatal(err2)
			}
			liveDTO.AttachmentsDetection(card)
			if !liveDTO.IsJob { //不是任务点或者已经是完成的任务点直接退出
				continue
			}
			ExecuteLive(userCache, courseItem, pointAction.Knowledge[index], &liveDTO)
			time.Sleep(5 * time.Second)
		}
	}

	//讨论任务点刷取
	if bbsDTOs != nil && user.CoursesCustom.AutoExam != 0 {
		if user.CoursesCustom.AutoExam == 1 { //检测AI可用性
			err2 := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType)
			if err2 != nil {
				lg.Print(lg.INFO, lg.BoldRed, "<"+setting.AiSetting.AiType+">", "AI不可用，错误信息："+err2.Error())
				os.Exit(0)
			}
		} else if user.CoursesCustom.AutoExam == 2 { // 检测外挂题库可用性
			err2 := external.CheckApiQueRequest(setting.ApiQueSetting.Url, 3, nil)
			if err2 != nil {
				lg.Print(lg.INFO, lg.BoldRed, "外挂题库不可用，错误信息："+err2.Error())
				os.Exit(0)
			}
		}
		for _, bbsDTO := range bbsDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, bbsDTO.KnowledgeID, bbsDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				log.Fatal(err2)
			}
			bbsDTO.AttachmentsDetection(card)
			if !bbsDTO.IsJob { //不是任务点或者已经是完成的任务点直接退出
				continue
			}
			ExecuteBBS(userCache, setting, courseItem, pointAction.Knowledge[index], &bbsDTO)
			time.Sleep(5 * time.Second)
		}
	}
}

// 检测答题是否有留空
func CheckAnswerIsAvoid(choices []entity.ChoiceQue, judges []entity.JudgeQue, fills []entity.FillQue, shorts []entity.ShortQue) bool {
	for _, choice := range choices {
		resStatus := true
		if choice.Answers != nil {
			candidateSelects := []string{} //待选
			for _, option := range choice.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, answer := range choice.Answers {
				var sortArray []qutils.Co = qutils.SimilarityArrayAndSort(answer, candidateSelects)
				if sortArray[0].Score >= 0.9 {
					resStatus = false
				}
			}

			//fmt.Sprintf("D -> 以上A B C正确。")
			if resStatus { //如果当前题目为留空态
				return true
			}
		} else {
			return true
		}
	}
	for _, judge := range judges {
		resStatus := true
		if judge.Answers != nil {
			for _, answer := range judge.Answers {
				for _, option := range judge.Options {
					if answer == option || answer == "错误" || answer == "正确" { //只需检测能够映射对应选项即可
						resStatus = false
					}
				}
			}
			if resStatus { //如果当前题目为留空态
				return true
			}
		} else {
			return true
		}
	}
	for _, fill := range fills {
		if fill.OpFromAnswer == nil || len(fill.OpFromAnswer) <= 0 {
			return true
		}
	}
	for _, short := range shorts {
		if short.OpFromAnswer == nil || len(short.OpFromAnswer) <= 0 {
			return true
		}
	}
	return false
}

// 答案修正匹配
func AnswerFixedPattern(choices []entity.ChoiceQue, judges []entity.JudgeQue, fills []entity.FillQue, shorts []entity.ShortQue) {
	//选择题修正
	for i, choice := range choices {
		if choice.Answers != nil {
			candidateSelects := []string{} //待选
			selectAnswers := []string{}
			for _, option := range choice.Options {
				candidateSelects = append(candidateSelects, option)
			}
			for _, answer := range choice.Answers {
				selectAnswers = append(selectAnswers, qutils.SimilarityArrayAnswer(answer, candidateSelects))
			}
			if selectAnswers != nil {
				choices[i].Answers = selectAnswers
			}
		}
	}
	for i, judge := range judges {
		if judge.Answers != nil {
			selectAnswer := []string{}
			for _, answer := range judge.Answers {
				answer = strings.ReplaceAll(answer, "对", "正确")
				selectAnswer = append(selectAnswer, qutils.SimilarityArrayAnswer(answer, []string{"正确", "错误"}))
			}
			judges[i].Answers = selectAnswer
		}
	}
}

// 常规刷视频逻辑
func ExecuteVideo2(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointVideoDto, key, courseCpi int) {

	if state, _ := xuexitong.VideoDtoFetchAction(cache, p); state {

		var playingTime = p.PlayTime
		if p.IsPassed == false && p.PlayTime == p.Duration {
			playingTime = 0
		}
		var overTime = 0
		//secList := []int{58} //停滞时间随机表
		selectSec := 58                     //默认60s
		extendSec := 5                      //过超提交停留时间
		limitTime := max(500, p.Duration/2) //过超时间最大限制
		mode := 1                           //0为Web模式，1为手机模式
		//flag := 0
		for {
			var playReport string
			var err error
			//selectSec = secList[rand.Intn(len(secList))] //随机选择时间
			if playingTime != p.Duration {

				if playingTime == p.PlayTime {
					playReport, err = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 3)
				} else {
					playReport, err = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 0)
				}
			} else {
				playReport, err = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 0)
			}
			if err != nil {
				//若报错500并且已经过超，那么可能是视屏有问题，所以最好直接跳过进行下一个视频
				if strings.Contains(err.Error(), "failed to fetch video, status code: 500") {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，触发风控500，重登次数过多已自动跳到下一任务点。", "，返回信息：", playReport, err.Error())
					break
				}
				//当报错无权限的时候尝试人脸
				if strings.Contains(err.Error(), "failed to fetch video, status code: 403") { //触发403立即使用人脸检测
					if mode == 1 {
						mode = 0
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Yellow, "检测到手机端触发403正在切换为Web端...")
						continue
					}
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Yellow, "触发403正在尝试绕过人脸识别...")
					//上传人脸
					pullJson, img, err2 := cache.GetHistoryFaceImg("")
					if err2 != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.BoldRed, "上传人脸失败，已自动跳过该视屏", pullJson, err2)
						return
						//os.Exit(0)
					}
					disturbImage := utils.ImageRGBDisturb(img)
					//uuid,qrEnc,ObjectId,successEnc
					_, _, _, _, errPass := xuexitong.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
					if errPass != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, "绕过人脸失败", errPass.Error(), "请在学习通客户端上确保最近一次人脸识别是正确的，yatori会自动拉取最近一次识别的人脸数据进行")
					} else {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Green, "绕过人脸成功")
					}

					//cid, _ := strconv.Atoi(p.CourseID)
					time.Sleep(5 * time.Second) //不要删！！！！一定要等待一小段时间才能请求PageMobile
					//card, enc, err2 := xuexitong.PageMobileChapterCardAction(
					//	cache, key, cid, p.KnowledgeID, p.CardIndex, courseCpi)
					//if err2 != nil {
					//	lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.BoldRed, "绕过人脸后获取卡片数据失败，已自动绕过该视屏", err2)
					//	//log.Fatal(err2)
					//	return
					//}
					//p.Enc = enc
					//p.AttachmentsDetection(card)
					//time.Sleep(5 * time.Second)
					//每次人脸过后都需要先进行isdrag=3的提交
					var startPlay string
					var startErr error
					playReport, startErr = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 0)

					if startErr != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.BoldRed, startPlay, startErr.Error())
						if playingTime-selectSec >= 0 {
							playingTime = playingTime - selectSec
						}
					}
					continue
				}
				if strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发404
					time.Sleep(10 * time.Second)
					continue
				}
			}
			if err != nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，返回信息：", err.Error())
				break
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed") == nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport)
				break
			}
			//阈值超限提交
			outTimeMsg := gojsonq.New().JSONString(playReport).Find("OutTimeMsg")
			if outTimeMsg != nil {
				if outTimeMsg.(string) == "观看时长超过阈值" {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), "观看时长超过阈值，已直接提交", lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
				if overTime == 0 {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				} else {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "过超时间：", strconv.Itoa(overTime)+"/"+strconv.Itoa(limitTime), " ", lg.Green, "过超提交成功", lg.Default, " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				}
				break
			}

			if overTime == 0 { //正常提交
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")
			} else { //过超提交
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "过超时间：", strconv.Itoa(overTime)+"/"+strconv.Itoa(limitTime), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")
			}
			if overTime >= limitTime { //过超提交触发
				lg.Print(lg.INFO, lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, "过超提交失败，自动进行下一任务...")
				break
			}

			if p.Duration-playingTime < selectSec && p.Duration != playingTime { //时间小于58s时
				playingTime = p.Duration
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			} else if p.Duration == playingTime { //记录过超提交触发条件
				//判断是否为任务点，如果为任务点那么就不累计过超提交
				if p.JobID == "" && p.Attachment == nil {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "该视频为非任务点看完后直接跳入下一视频")
					break
				} else {
					overTime += extendSec
				}
				time.Sleep(time.Duration(extendSec) * time.Second)
			} else { //正常计时逻辑
				playingTime = playingTime + selectSec
				time.Sleep(time.Duration(selectSec) * time.Second)
			}
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 58倍速模式刷视频逻辑
func ExecuteVideoQuickSpeed(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointVideoDto, key, courseCpi int) {
	if state, _ := xuexitong.VideoDtoFetchAction(cache, p); state {
		var playingTime = p.PlayTime
		if p.IsPassed == false && p.PlayTime == p.Duration {
			playingTime = 0
		}
		var overTime = 0
		selectSec := 58
		mode := 1 //0为web模式，1为手机模式
		for {
			var playReport string
			var err error
			if playingTime != p.Duration {
				playReport, err = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 0)
			} else {
				playReport, err = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 0)
			}
			if err != nil {
				//若报错500并且已经过超，那么可能是视屏有问题，所以最好直接跳过进行下一个视频
				if strings.Contains(err.Error(), "failed to fetch video, status code: 500") {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，触发风控500，重登次数过多已自动跳到下一任务点。", "，返回信息：", playReport, err.Error())
					break

				}
				//当报错无权限的时候尝试人脸
				if strings.Contains(err.Error(), "failed to fetch video, status code: 403") { //触发403立即使用人脸检测
					if mode == 1 {
						mode = 0
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Yellow, "检测到手机端触发403正在切换为Web端...")
						continue
					}
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Yellow, "触发403正在尝试绕过人脸识别...")
					//上传人脸
					faceImg, err := utils.GetFaceBase64()
					disturbImage := utils.ImageRGBDisturb(faceImg)
					if err != nil {
						fmt.Println(err)
					}
					_, _, _, successEnc, errPass := xuexitong.PassFaceAction3(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
					if errPass != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, "绕过人脸失败", errPass.Error(), "请在学习通客户端上确保最近一次人脸识别是正确的，yatori会自动拉取最近一次识别的人脸数据进行")
					} else {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Green, "绕过人脸成功")
					}
					cid, _ := strconv.Atoi(p.CourseID)
					time.Sleep(3 * time.Second)
					card, enc, err := xuexitong.PageMobileChapterCardAction(
						cache, key, cid, p.KnowledgeID, p.CardIndex, courseCpi)
					if err != nil {
						log.Fatal(err)
					}
					p.AttachmentsDetection(card)
					p.Enc = enc
					p.VideoFaceCaptureEnc = successEnc
					time.Sleep(5 * time.Second)
					//每次人脸过后都需要先进行isdrag=3的提交

					var startPlay string
					var startErr error
					playReport, startErr = xuexitong.VideoSubmitStudyTimeAction(cache, p, playingTime, mode, 3)
					if startErr != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, startPlay, startErr.Error())
						if playingTime-selectSec >= 0 {
							playingTime = playingTime - selectSec
						}
					}
					continue
				}
				if strings.Contains(err.Error(), "failed to fetch video, status code: 404") { //触发202立即使用人脸检测
					time.Sleep(10 * time.Second)
					continue
				}
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport, err.Error())
				break
			}
			//阈值超限提交
			outTimeMsg := gojsonq.New().JSONString(playReport).Find("OutTimeMsg")
			if outTimeMsg != nil {
				if outTimeMsg.(string) == "观看时长超过阈值" {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), "观看时长超过阈值，已直接提交", lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				break
			}
			lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")

			if overTime >= 150 { //过超提交触发
				lg.Print(lg.INFO, lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "过超提交中。。。。")
				break
			}

			if p.Duration-playingTime < selectSec && p.Duration != playingTime { //时间小于58s时
				playingTime = p.Duration
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			} else if p.Duration == playingTime { //记录过超提交触发条件
				overTime += 1
				time.Sleep(1 * time.Second)
			} else { //正常计时逻辑
				playingTime = playingTime + selectSec
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 常规刷文档逻辑
func ExecuteDocument(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointDocumentDto) {
	report, err := point.ExecuteDocument(cache, p)
	if gojsonq.New().JSONString(report).Find("status") == nil || err != nil || gojsonq.New().JSONString(report).Find("status") == false {
		if err == nil {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "文档学习提交接口访问异常（可能是因为该文档不是任务点导致的），返回信息：", report)
		} else {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "文档学习提交接口访问异常（可能是因为该文档不是任务点导致的），返回信息：", report, err.Error())
		}

		//log.Fatalln(err)
	}

	if gojsonq.New().JSONString(report).Find("status").(bool) {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "文档阅览状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(report).Find("status").(bool)), lg.Default, " ")
	}
}

// 常规外链任务处理
func ExecuteHyperlink(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointHyperlinkDto) {
	report, err := point.ExecuteHyperlink(cache, p)
	if gojsonq.New().JSONString(report).Find("status") == nil || err != nil || gojsonq.New().JSONString(report).Find("status") == false {
		if err == nil {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "外链任务点学习提交接口访问异常，返回信息：", report)
		} else {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "外链任务点学习提交接口访问异常，返回信息：", report, err.Error())
		}
	}

	if gojsonq.New().JSONString(report).Find("status").(bool) {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "外链任务点状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(report).Find("status").(bool)), lg.Default, " ")
	}
}

// 常规直播任务处理
func ExecuteLive(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointLiveDto) {
	point.PullLiveInfoAction(cache, p)
	var passValue float64 = 90

	//如果该直播还未开播
	if p.LiveStatusCode == 0 {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Yellow, "该直播任务点还未开播，已自动跳过")
		return
	}
	relationReport, err2 := point.LiveCreateRelationAction(cache, p)
	if err2 != nil {
		lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "直播任务点建立联系接口访问异常，返回信息：", relationReport, err2.Error())
	} else {
		lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.Green, "直播任务点建立联系成功，返回信息：", relationReport)
	}

	for {
		report, err := point.ExecuteLive(cache, p)

		point.PullLiveInfoAction(cache, p) //更新直播节点结构体进度
		if err != nil {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "直播任务点学习提交接口访问异常，返回信息：", report, err.Error())
		}

		if strings.Contains(report, "@success") {
			lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", "直播任务点状态：", lg.Green, report, lg.Default, "，直播观看进度：", lg.Green, fmt.Sprintf("%.2f", p.VideoCompletePercent), "%")
		} else {
			if err != nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "直播任务点学习提交接口访问异常，返回信息：", report, err.Error())
			} else {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】", lg.BoldRed, "直播任务点学习提交接口访问异常，返回信息：", report)
			}
		}
		if p.VideoCompletePercent >= passValue {
			lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Green, "直播任务点已完成")
			return
		}
		time.Sleep(30 * time.Second)
	}
}

// 常规讨论任务处理
func ExecuteBBS(cache *xuexitongApi.XueXiTUserCache, setting config.Setting, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, bbsDto *entity.PointBBsDto) {
	bbsTopic, err := point.PullBbsInfoAction(cache, bbsDto) //拉取相关数据
	if err != nil {
		fmt.Println(err)
	}
	report, err := bbsTopic.AIAnswer(cache, bbsDto, setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.AiType, setting.AiSetting.APIKEY)
	if gojsonq.New().JSONString(report).Find("status") == nil || err != nil || gojsonq.New().JSONString(report).Find("status") == false {
		if err == nil {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", bbsTopic.Title, "】", lg.BoldRed, "外链任务点学习提交接口访问异常，返回信息：", report)
		} else {
			lg.Print(lg.INFO, `[`, cache.Name, `] `, "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", bbsTopic.Title, "】", lg.BoldRed, "外链任务点学习提交接口访问异常，返回信息：", report, err.Error())
		}
	}

	if gojsonq.New().JSONString(report).Find("status").(bool) {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", bbsTopic.Title, "】 >>> ", "讨论任务点状态：", lg.Green, lg.Green, gojsonq.New().JSONString(report).Find("msg").(string), lg.Default, " ")
	}
}

// 作业处理逻辑
func WorkAction(userCache *xuexitongApi.XueXiTUserCache, user *config.Users, setting config.Setting, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, questionAction entity.Question) {
	if user.CoursesCustom.AutoExam == 1 {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", lg.Default, "【"+courseItem.CourseName+"】 ", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.Yellow, "正在AI自动写章节作业...")
	} else {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", lg.Default, "【"+courseItem.CourseName+"】 ", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.Yellow, "正在外挂题库自动写章节作业...")
	}

	//选择题
	for i := range questionAction.Choice {
		q := &questionAction.Choice[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXChoiceQue: *q,
			})

			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}
	//判断题
	for i := range questionAction.Judge {
		q := &questionAction.Judge[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXJudgeQue: *q,
			})

			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}
	//填空题
	for i := range questionAction.Fill {
		q := &questionAction.Fill[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXFillQue: *q,
			})
			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}
	//简答题
	for i := range questionAction.Short {
		q := &questionAction.Short[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXShortQue: *q,
			})
			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}
	//名词解释
	for i := range questionAction.TermExplanation {
		q := &questionAction.TermExplanation[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXTermExplanationQue: *q,
			})
			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}
	//论述题
	for i := range questionAction.Essay {
		q := &questionAction.Essay[i] // 获取对应选项
		switch user.CoursesCustom.AutoExam {
		case 1:
			message := xuexitong.AIProblemMessage(questionAction.Title, q.Type.String(), entity.ExamTurn{
				XueXEssayQue: *q,
			})
			aiSetting := setting.AiSetting //获取AI设置
			q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
		case 2:
			q.AnswerExternalGet(setting.ApiQueSetting.Url)
		}

	}

	var resultStr string
	if user.CoursesCustom.ExamAutoSubmit == 0 {
		AnswerFixedPattern(questionAction.Choice, questionAction.Judge, questionAction.Fill, questionAction.Short)
		resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, false)
	} else if user.CoursesCustom.ExamAutoSubmit == 1 {
		AnswerFixedPattern(questionAction.Choice, questionAction.Judge, questionAction.Fill, questionAction.Short)
		resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, true)
	} else if user.CoursesCustom.ExamAutoSubmit == 2 {
		AnswerFixedPattern(questionAction.Choice, questionAction.Judge, questionAction.Fill, questionAction.Short)
		if CheckAnswerIsAvoid(questionAction.Choice, questionAction.Judge, questionAction.Fill, questionAction.Short) {
			resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, false) //留空了，只保存
			//如果提交失败那么直接输出AI答题的文本
			if gojsonq.New().JSONString(resultStr).Find("status") == false {
				if user.CoursesCustom.AutoExam == 1 {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.BoldRed, "AI答题保存失败,返回信息："+resultStr, " AI答题信息：", questionAction.String())
				} else if user.CoursesCustom.AutoExam == 2 {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.BoldRed, "外挂题库答题保存失败,返回信息："+resultStr, " 外挂题库答题信息：", questionAction.String())
				}

			}
		} else {
			//AnswerFixedPattern(questionAction.Choice, questionAction.Judge, questionAction.Fill, questionAction.Short)
			resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, true) //没有留空则提交
			if gojsonq.New().JSONString(resultStr).Find("status") == false {
				if user.CoursesCustom.AutoExam == 1 {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.BoldRed, "AI答题保存失败,返回信息："+resultStr, " AI答题信息：", questionAction.String())
				} else if user.CoursesCustom.AutoExam == 2 {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.BoldRed, "外挂题库答题保存失败,返回信息："+resultStr, " 外挂题库答题信息：", questionAction.String())
				}
			}
		}
	}
	//提交试卷成功的话{"msg":"success!","stuStatus":4,"backUrl":"","url":"/mooc-ans/api/work?courseid=250215285&workId=b63d4e7466624ace9c382cd112c9c95a&clazzId=125521307&knowledgeid=951783044&ut=s&type=&submit=true&jobid=work-6967802218b44f4dace8e3a8755cf3d9&enc=db5c2413ac1367c5ed28b4cfa5194318&ktoken=c0bf3b45e0b3e625e377cae3b77e1cfa&mooc2=0&skipHeader=true&originJobId=null","status":true}
	//提交作业失败的话{"msg" : "作业提交失败！","status" : false}
	if user.CoursesCustom.AutoExam == 1 {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.Green, "章节作业AI答题完毕,服务器返回信息：", resultStr)
	} else if user.CoursesCustom.AutoExam == 2 {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", questionAction.Title, "】", lg.Green, "章节作业外挂题库答题完毕,服务器返回信息：", resultStr)
	}

}
