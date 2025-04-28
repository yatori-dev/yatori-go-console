package xuexitong

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
	"yatori-go-console/config"
	utils2 "yatori-go-console/utils"

	"github.com/thedevsaddam/gojsonq"
	xuexitong "github.com/yatori-dev/yatori-go-core/aggregation/xuexitong"
	"github.com/yatori-dev/yatori-go-core/api/entity"
	xuexitongApi "github.com/yatori-dev/yatori-go-core/api/xuexitong"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
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
		lg.Print(lg.INFO, `[`, courseItem.CourseName, `] `, lg.BoldRed, "拉取章节信息接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())
		log.Fatal()
	}

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

	lg.Print(lg.INFO, "[", lg.Green, courseItem.CourseName, lg.Default, "] ", lg.Purple, "正在学习该课程")
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
			lg.Print(lg.INFO, `[`, courseItem.CourseName, `] `, lg.BoldRed, "课程数据没有需要刷的课，可能接口访问异常。若需要继续可以配置中添加排除此异常课程。")
			log.Fatal()
		}
		// 视屏类型
		if videoDTOs != nil {
			for _, videoDTO := range videoDTOs {
				card, err := xuexitong.PageMobileChapterCardAction(
					userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, courseItem.Cpi)
				if err != nil {
					log.Fatal(err)
				}
				videoDTO.AttachmentsDetection(card)
				ExecuteVideo2(userCache, &videoDTO)
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

	}
	lg.Print(lg.INFO, "[", lg.Green, courseItem.CourseName, lg.Default, "] ", lg.Purple, "课程学习完毕")
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
					lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport)
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
		var flag = true
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
					lg.Print(lg.INFO, `[`, cache.Name, `] `, lg.BoldRed, "提交学时接口访问异常，返回信息：", playReport)
					break
				}
				if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true { //看完了，则直接退出
					lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
					flag = true
					break
				}
				if flag {
					lg.Print(lg.INFO, "过超提交中。。。。")
					flag = false
				}
				//lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", " 【", p.Title, "】 >>> ", "提交状态：", lg.Green, lg.Green, strconv.FormatBool(gojsonq.New().JSONString(playReport).Find("isPassed").(bool)), lg.Default, " ", "观看时间：", strconv.Itoa(p.Duration)+"/"+strconv.Itoa(p.Duration), " ", "观看进度：", fmt.Sprintf("%.2f", float32(p.Duration)/float32(p.Duration)*100), "%")
				time.Sleep(time.Second)
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
