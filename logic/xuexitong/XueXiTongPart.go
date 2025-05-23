package xuexitong

import (
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	xuexitong "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	"github.com/yatori-dev/yatori-go-core/utils"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"yatori-go-console/config"
	utils2 "yatori-go-console/utils"
)

var videosLock sync.WaitGroup //视频锁
var usersLock sync.WaitGroup  //用户锁

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
			error := xuexitong.XueXiTLoginAction(cache) // 登录
			if error != nil {
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.White, "] ", lg.Red, error.Error())
				log.Fatal(error) //登录失败则直接退出
			}
			// go keepAliveLogin(cache) //携程保活
			UserCaches = append(UserCaches, cache)
		}
	}
	return UserCaches
}

// 开始刷课模块
func RunBrushOperation(setting config.Setting, users []config.Users, userCaches []*xuexitongApi.XueXiTUserCache) {
	for i, user := range userCaches {
		usersLock.Add(1)
		go userBlock(setting, &users[i], user)
	}
	usersLock.Wait()
}

// 加锁，防止同时过多调用音频通知导致BUG,speak自带的没用，所以别改
// 以用户作为刷课单位的基本块
var soundMut sync.Mutex

func userBlock(setting config.Setting, user *config.Users, cache *xuexitongApi.XueXiTUserCache) {
	// list, err := xuexitong.XueXiTCourseDetailForCourseIdAction(cache, "261619055656961")
	courseList, err := xuexitong.XueXiTPullCourseAction(cache)
	if err != nil {
		log.Fatal(err)
	}
	for _, course := range courseList {
		videosLock.Add(1)
		// fmt.Println(course)
		nodeListStudy(setting, user, cache, &course)
	}
	videosLock.Wait()
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
		videosLock.Done()
		return
	}
	//包含指定课程
	if len(user.CoursesCustom.IncludeCourses) != 0 && !config.CmpCourse(courseItem.CourseName, user.CoursesCustom.IncludeCourses) {
		videosLock.Done()
		return
	}
	key, _ := strconv.Atoi(courseItem.Key)
	action, _, err := xuexitong.PullCourseChapterAction(userCache, courseItem.Cpi, key) //获取对应章节信息

	if err != nil {
		if strings.Contains(err.Error(), "课程章节为空") {
			lg.Print(lg.INFO, `[`, courseItem.CourseName, `] `, lg.BoldRed, "该课程章节为空已自动跳过")
			videosLock.Done()
			return
		}
		lg.Print(lg.INFO, `[`, courseItem.CourseName, `] `, lg.BoldRed, "拉取章节信息接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())
		log.Fatal()
	}
	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "获取课程章节成功 (共 ", lg.Yellow, strconv.Itoa(len(action.Knowledge)), lg.Default, " 个) ")

	var nodes []int
	for _, item := range action.Knowledge {
		nodes = append(nodes, item.ID)
	}
	courseId, _ := strconv.Atoi(courseItem.CourseID)
	userId, _ := strconv.Atoi(userCache.UserID)
	// 检测节点完成情况
	pointAction, err := xuexitong.ChapterFetchPointAction(userCache, nodes, &action, key, userId, courseItem.Cpi, courseId)
	if err != nil {
		lg.Print(lg.INFO, `[`, courseItem.CourseName, `] `, lg.BoldRed, "探测节点完成情况接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())
		log.Fatal()
	}
	var isFinished = func(index int) bool {
		if index < 0 || index >= len(pointAction.Knowledge) {
			return false
		}
		i := pointAction.Knowledge[index]
		return i.PointTotal >= 0 && i.PointTotal == i.PointFinished
	}

	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Purple, "正在学习该课程")
	for index, _ := range nodes {
		if isFinished(index) { //如果完成了的那么直接跳过
			continue
		}
		_, fetchCards, err := xuexitong.ChapterFetchCardsAction(userCache, &action, nodes, index, courseId, key, courseItem.Cpi)
		if err != nil {
			log.Fatal(err)
		}
		videoDTOs, workDTOs, documentDTOs := entity.ParsePointDto(fetchCards)
		if videoDTOs == nil && workDTOs == nil && documentDTOs == nil {
			lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "课程数据没有需要刷的课，可能接口访问异常。若需要继续可以配置中添加排除此异常课程。")
			log.Fatal()
		}
		// 视屏类型
		if videoDTOs != nil && user.CoursesCustom.VideoModel != 0 {
			for _, videoDTO := range videoDTOs {
				card, err := xuexitong.PageMobileChapterCardAction(
					userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, courseItem.Cpi)
				if err != nil {
					log.Fatal(err)
				}
				videoDTO.AttachmentsDetection(card)
				switch user.CoursesCustom.VideoModel {
				case 1:
					ExecuteVideo2(userCache, &videoDTO) //普通模式
				case 2:
					ExecuteVideoQuickSpeed(userCache, &videoDTO) // 暴力模式
				}

				time.Sleep(5 * time.Second)
			}
		}
		// 文档类型
		if documentDTOs != nil {
			for _, documentDTO := range documentDTOs {
				card, err := xuexitong.PageMobileChapterCardAction(
					userCache, key, courseId, documentDTO.KnowledgeID, documentDTO.CardIndex, courseItem.Cpi)
				if err != nil {
					log.Fatal(err)
				}
				documentDTO.AttachmentsDetection(card)
				//point.ExecuteDocument(userCache, &documentDTO)
				ExecuteDocument(userCache, &documentDTO)
				time.Sleep(5 * time.Second)
			}
		}

		//作业刷取
		if workDTOs != nil && user.CoursesCustom.AutoExam != 0 {

			//检测AI可用性
			err := utils.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType)
			if err != nil {
				lg.Print(lg.INFO, lg.BoldRed, "<"+setting.AiSetting.AiType+">", "AI不可用，错误信息："+err.Error())
				os.Exit(0)
			}
			for _, workDTO := range workDTOs {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", lg.Default, " 【"+courseItem.CourseName+"】 ", lg.Yellow, "正在AI自动写章节作业...")
				//以手机端拉取章节卡片数据
				mobileCard, _ := xuexitong.PageMobileChapterCardAction(userCache, key, courseId, workDTO.KnowledgeID, workDTO.CardIndex, courseItem.Cpi)
				flag, _ := workDTO.AttachmentsDetection(mobileCard)
				if !flag {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", " 【", courseItem.CourseName, "】", lg.Green, "该作业已完成，已自动跳过")
					continue
				}
				xuexitong.WorkPageFromAction(userCache, &workDTO)
				//for _, input := range fromAction {
				//	fmt.Printf("Name: %s, Value: %s, Type: %s, ID: %s\n", input.Name, input.Value, input.Type, input.ID)
				//}
				questionAction := xuexitong.ParseWorkQuestionAction(userCache, &workDTO)
				fmt.Println(questionAction)
				for i := range questionAction.Choice {
					q := &questionAction.Choice[i] // 获取对应选项
					message := xuexitong.AIProblemMessage(q.Type.String(), entity.ExamTurn{
						ChoiceQue: *q,
					})
					aiSetting := setting.AiSetting //获取AI设置
					q.AnswerAIGet(userCache.UserID, aiSetting.AiUrl, aiSetting.Model, aiSetting.AiType, message, aiSetting.APIKEY)
				}
				var resultStr string
				if user.CoursesCustom.ExamAutoSubmit == 0 {
					resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, false)
				} else if user.CoursesCustom.ExamAutoSubmit == 1 {
					resultStr = xuexitong.WorkNewSubmitAnswerAction(userCache, questionAction, true)
				}
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "<"+setting.AiSetting.AiType+">", " 【", courseItem.CourseName, "】", lg.Green, "章节作业AI答题完毕,服务器返回信息：", resultStr)

			}
		}

	}
	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Purple, "课程学习完毕")
	videosLock.Done()
}

// 常规刷视频逻辑
func ExecuteVideo(cache *xuexitongApi.XueXiTUserCache, p *entity.PointVideoDto) {
	if state, _ := xuexitong.VideoDtoFetchAction(cache, p); state {
		var playingTime = p.PlayTime
		for {
			if p.Duration-playingTime >= 58 {
				playReport, err := cache.VideoDtoPlayReport(p, playingTime, 0, 4, nil)
				if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport, err.Error())
					break
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")
				playingTime = playingTime + 58
				time.Sleep(58 * time.Second)
			} else if p.Duration-playingTime < 58 {
				playReport, err := cache.VideoDtoPlayReport(p, p.Duration, 2, 4, nil)
				if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
					lg.Print(lg.INFO, `[`, cache.Name, `] `, " 【", p.Title, "】 >>> ", lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport)
					break
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			}
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 常规刷视频逻辑
func ExecuteVideo2(cache *xuexitongApi.XueXiTUserCache, p *entity.PointVideoDto) {
	if state, _ := xuexitong.VideoDtoFetchAction(cache, p); state {
		var playingTime = p.PlayTime
		var overTime = 0
		for {
			var playReport string
			var err error
			if playingTime != p.Duration {
				playReport, err = cache.VideoDtoPlayReport(p, playingTime, 0, 4, nil)
			} else {
				playReport, err = cache.VideoDtoPlayReport(p, playingTime, 4, 4, nil) //4代表播放结束
			}

			if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, " 【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport, err.Error())
				break
			}
			//阈值超限提交
			outTimeMsg := gojsonq.New().JSONString(playReport).Find("OutTimeMsg")
			if outTimeMsg != nil {
				if outTimeMsg.(string) == "观看时长超过阈值" {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), "观看时长超过阈值，已直接提交", lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				break
			}
			lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")

			if overTime >= 200 { //过超提交触发
				lg.Print(lg.INFO, lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "过超提交中。。。。")
				break
			}

			if p.Duration-playingTime < 58 && p.Duration != playingTime { //时间小于58s时
				playingTime = p.Duration
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			} else if p.Duration == playingTime { //记录过超提交触发条件
				overTime += 2
				time.Sleep(2 * time.Second)
			} else { //正常计时逻辑
				playingTime = playingTime + 60
				time.Sleep(60 * time.Second)
			}
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 58倍速模式刷视频逻辑
func ExecuteVideoQuickSpeed(cache *xuexitongApi.XueXiTUserCache, p *entity.PointVideoDto) {
	if state, _ := xuexitong.VideoDtoFetchAction(cache, p); state {
		var playingTime = p.PlayTime
		var overTime = 0
		for {
			var playReport string
			var err error
			if playingTime != p.Duration {
				playReport, err = cache.VideoDtoPlayReport(p, playingTime, 0, 4, nil)
			} else {
				playReport, err = cache.VideoDtoPlayReport(p, playingTime, 4, 4, nil) //4代表播放结束
			}

			if gojsonq.New().JSONString(playReport).Find("isPassed") == nil || err != nil {
				lg.Print(lg.INFO, `[`, cache.Name, `] `, " 【", p.Title, "】", lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport, err.Error())
				break
			}
			//阈值超限提交
			outTimeMsg := gojsonq.New().JSONString(playReport).Find("OutTimeMsg")
			if outTimeMsg != nil {
				if outTimeMsg.(string) == "观看时长超过阈值" {
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), "观看时长超过阈值，已直接提交", lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					break
				}
			}
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
				lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				break
			}
			lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(playingTime)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(playingTime)/float32(p.Duration)*100), "%")

			if overTime >= 100 { //过超提交触发
				lg.Print(lg.INFO, lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "过超提交中。。。。")
				break
			}

			if p.Duration-playingTime < 60 && p.Duration != playingTime { //时间小于58s时
				playingTime = p.Duration
				time.Sleep(time.Duration(p.Duration-playingTime) * time.Second)
			} else if p.Duration == playingTime { //记录过超提交触发条件
				overTime += 1
				time.Sleep(1 * time.Second)
			} else { //正常计时逻辑
				playingTime = playingTime + 60
				time.Sleep(1 * time.Second)
			}
		}
	} else {
		log.Fatal("视频解析失败")
	}
}

// 常规刷文档逻辑
func ExecuteDocument(cache *xuexitongApi.XueXiTUserCache, p *entity.PointDocumentDto) {
	report, err := cache.DocumentDtoReadingReport(p)
	if gojsonq.New().JSONString(report).Find("status") == nil || err != nil {
		lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", report, err.Error())
		log.Fatalln(err)
	}
	if gojsonq.New().JSONString(report).Find("status").(bool) {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "文档阅览状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(report).Find("status").(bool)), lg.Default, " ")
	}
}

// 作业处理逻辑
func WorkAction() {

}
