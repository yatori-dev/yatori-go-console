package activity

import (
	"fmt"
	"log"
	"math/rand"
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

	"github.com/yatori-dev/yatori-go-core/utils"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

type XXTActivity struct {
	UserActivityBase
}

// 学习通扩展能力
type XXTAbility interface {
	PullCourseList() ([]xuexitong.XueXiTCourse, error) //拉取课程
}

// 用户登录模块
func (activity *XXTActivity) Login() error {
	//TODO implement me
	cache := &xuexitongApi.XueXiTUserCache{Name: activity.User.Account, Password: activity.User.Password}
	//设置代理IP
	if activity.User.IsProxy == 1 {
		cache.IpProxySW = true
		cache.ProxyIP = "http://" + utils2.RandProxyStr()
	}
	loginError := xuexitong.XueXiTLoginAction(cache) // 登录
	if loginError != nil {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.White, "] ", lg.Red, loginError.Error())
		//os.Exit(0) //登录失败直接退出
		return loginError
	}
	activity.UserCache = cache //赋值寄存
	return nil
}

// 启动
func (activity *XXTActivity) Start() error {
	//TODO implement me
	activity.IsRunning = true
	activity.userBlock() //开刷
	return nil
}

// 暂停
func (activity *XXTActivity) Stop() error {
	//TODO implement me
	activity.IsRunning = false
	return nil
}

// 拉取课程列表
func (user *XXTActivity) PullCourseList() ([]xuexitong.XueXiTCourse, error) {
	cache := user.UserCache.(*xuexitongApi.XueXiTUserCache)
	courseList, err := xuexitong.XueXiTPullCourseAction(cache)
	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", lg.Red, "拉取课程失败")
		log.Fatal(err)
	}
	return courseList, nil
}

// 直接刷课
func (activity *XXTActivity) userBlock() {
	cache := activity.UserCache.(*xuexitongApi.XueXiTUserCache)
	courseList, err := xuexitong.XueXiTPullCourseAction(cache)
	if !activity.IsRunning { //打断
		return
	}
	if err != nil {
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", lg.Red, "拉取课程失败")
		log.Fatal(err)
	}

	var nodesLock sync.WaitGroup //视频锁
	for _, course := range courseList {
		nodesLock.Add(1)
		// fmt.Println(course) //创建当前循环变量的独立副本

		activity.nodeListStudy(&activity.User, cache, &course)
		nodesLock.Done()
		if !activity.IsRunning { //打断
			return
		}
	}
	if !activity.IsRunning { //打断
		return
	}
	nodesLock.Wait()
	lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", lg.Purple, "所有待学习课程学习完毕")
	//如果开启了邮箱通知

}

func (activity *XXTActivity) nodeListStudy(user *config.User, userCache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse) {
	//过滤课程---------------------------------
	for i, uid := range user.CoursesCustom.IncludeCourses {
		if uid == courseItem.Key {
			break
		}
		if i+1 == len(user.CoursesCustom.IncludeCourses) {
			return
		}
	}

	if !activity.IsRunning { //打断
		return
	}
	//如果课程还未开课则直接退出
	if !courseItem.IsStart {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Blue, "该课程还未开课，已自动跳过该课程")
		return
	}
	//如果该课程已经结束
	if courseItem.State == 1 {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Blue, "该课程已经结束，已自动跳过该课程")
		return
	}

	key, _ := strconv.Atoi(courseItem.Key)
	action, _, err := xuexitong.PullCourseChapterAction(userCache, courseItem.Cpi, key) //获取对应章节信息
	if !activity.IsRunning {                                                            //打断
		return
	}
	if err != nil {
		if strings.Contains(err.Error(), "课程章节为空") {
			lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该课程章节为空已自动跳过")
			return
		}
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "拉取章节信息接口访问异常，若需要继续可以配置中添加排除此异常课程。返回信息：", err.Error())
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

	var nodeLock sync.WaitGroup
	//遍历结点
	for index := range nodes {
		if isFinished(index) { //如果完成了的那么直接跳过
			continue
		}
		activity.nodeRun(userCache, courseItem, pointAction, action, nodes, index, key, courseId)
	}
	nodeLock.Wait()
	lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", "[", courseItem.CourseName, "] ", lg.Purple, "课程学习完毕")
}

// 任务点分流运行
func (activity *XXTActivity) nodeRun(userCache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse,
	pointAction xuexitong.ChaptersList, action xuexitong.ChaptersList, nodes []int, index int, key int, courseId int) {
	_, fetchCards, err1 := xuexitong.ChapterFetchCardsAction(userCache, &action, nodes, index, courseId, key, courseItem.Cpi)
	if err1 != nil {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "无法正常拉取卡片信息，请联系作者查明情况,报错信息：", err1.Error())
		return
	}
	videoDTOs, workDTOs, documentDTOs, hyperlinkDTOs, liveDTOs, bbsDTOs := entity.ParsePointDto(fetchCards)
	if !activity.IsRunning { //打断
		return
	}
	if videoDTOs == nil && workDTOs == nil && documentDTOs == nil && hyperlinkDTOs == nil && liveDTOs == nil && bbsDTOs == nil {
		lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, `[`, pointAction.Knowledge[index].Name, `] `, lg.Blue, "课程对应章节无任何任务节点，已自动跳过")
		return
	}
	// 视屏类型

	for _, videoDTO := range videoDTOs {
		if !activity.IsRunning { //打断
			return
		}
		card, enc, err2 := xuexitong.PageMobileChapterCardAction(
			userCache, key, courseId, videoDTO.KnowledgeID, videoDTO.CardIndex, courseItem.Cpi)
		if err2 != nil {
			if strings.Contains(err2.Error(), "章节未开放") {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该章节未开放，可能是因为前面章节有任务点未学完导致后续任务点未开放，已自动跳过该任务点")
				break
			}
			if strings.Contains(err2.Error(), "没有历史人脸") {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "过人脸失败，该账号可能从未进行过人脸识别，请先进行一次人脸识别后再试")
				os.Exit(0)
			}
			if strings.Contains(err2.Error(), "活体检测失败") {
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "过人脸失败，该账号所录入的人脸可能并不规范，请自行拍摄人脸放到assets/face/账号名称.jpg路径下再重试")
				os.Exit(0)
			}
			lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.Red, err2.Error())
			os.Exit(0)
		}
		_, err := videoDTO.AttachmentsDetection(card)
		if err != nil {
		}

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

		activity.ExecuteVideo2(userCache, courseItem, pointAction.Knowledge[index], &videoDTO, key, courseItem.Cpi) //普通模式
		randSleepTime := rand.Intn(51) + 10
		time.Sleep(time.Duration(randSleepTime) * time.Second)
	}
	// 文档类型
	if documentDTOs != nil {
		for _, documentDTO := range documentDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, documentDTO.KnowledgeID, documentDTO.CardIndex, courseItem.Cpi)
			if err2 != nil {
				if strings.Contains(err2.Error(), "章节未开放") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该章节未开放，可能是因为前面章节有任务点未学完导致后续任务点未开放，已自动跳过该任务点")
					break
				}
				if strings.Contains(err2.Error(), "没有历史人脸") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "过人脸失败，该账号可能从未进行过人脸识别，请先进行一次人脸识别后再试")
					os.Exit(0)
				}
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.Red, err2.Error())
				os.Exit(0)
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

	//外链任务点刷取
	if hyperlinkDTOs != nil {
		for _, hyperlinkDTO := range hyperlinkDTOs {
			card, _, err2 := xuexitong.PageMobileChapterCardAction(
				userCache, key, courseId, hyperlinkDTO.KnowledgeID, hyperlinkDTO.CardIndex, courseItem.Cpi)

			if err2 != nil {
				if strings.Contains(err2.Error(), "章节未开放") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该章节未开放，可能是因为前面章节有任务点未学完导致后续任务点未开放，已自动跳过该任务点")
					return
				}
				if strings.Contains(err2.Error(), "没有历史人脸") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "过人脸失败，该账号可能从未进行过人脸识别，请先进行一次人脸识别后再试")
					os.Exit(0)
				}
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.Red, err2.Error())
				os.Exit(0)
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
				if strings.Contains(err2.Error(), "章节未开放") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "该章节未开放，可能是因为前面章节有任务点未学完导致后续任务点未开放，已自动跳过该任务点")
					return
				}
				if strings.Contains(err2.Error(), "没有历史人脸") {
					lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.BoldRed, "过人脸失败，该账号可能从未进行过人脸识别，请先进行一次人脸识别后再试")
					os.Exit(0)
				}
				lg.Print(lg.INFO, "[", lg.Green, userCache.Name, lg.Default, "] ", `[`, courseItem.CourseName, `] `, lg.Red, err2.Error())
				os.Exit(0)
			}
			liveDTO.AttachmentsDetection(card)
			if !liveDTO.IsJob { //不是任务点或者已经是完成的任务点直接退出
				continue
			}
			ExecuteLive(userCache, courseItem, pointAction.Knowledge[index], &liveDTO)
			time.Sleep(5 * time.Second)
		}
	}

}

// 常规刷视频逻辑
func (activity *XXTActivity) ExecuteVideo2(cache *xuexitongApi.XueXiTUserCache, courseItem *xuexitong.XueXiTCourse, knowledgeItem xuexitong.KnowledgeItem, p *entity.PointVideoDto, key, courseCpi int) {

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
			if !activity.IsRunning { //打断
				return
			}
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
					_, _, _, _, errPass := xuexitong.PassFacePCAction(cache, p.CourseID, p.ClassID, p.Cpi, fmt.Sprintf("%d", p.KnowledgeID), p.Enc, p.JobID, p.ObjectID, p.Mid, p.RandomCaptureTime, disturbImage)
					if errPass != nil {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, "绕过人脸失败", errPass.Error(), "请在学习通客户端上确保最近一次人脸识别是正确的，yatori会自动拉取最近一次识别的人脸数据进行")
					} else {
						lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Green, "绕过人脸成功")
					}
					time.Sleep(5 * time.Second) //不要删！！！！一定要等待一小段时间才能请求PageMobile
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
			if gojsonq.New().JSONString(playReport).Find("isPassed").(bool) == true && playingTime >= p.Duration { //看完了，则直接退出
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
		lg.Print(lg.INFO, "[", lg.Green, cache.Name, lg.Default, "] ", "【", courseItem.CourseName, "】", "【", knowledgeItem.Label, " ", knowledgeItem.Name, "】", "【", p.Title, "】 >>> ", lg.Red, "该视屏任务点解析失败，可能是任务点视屏本身问题，已自动跳过")
		//log.Fatal("视频解析失败")
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
