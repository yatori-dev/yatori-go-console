// Package runner 是 console 基于 yatori-go-core v3 统一 SDK 的刷课执行器。
//
// v3 把各平台的 aggregation/api 全部收敛到了 internal，对外只暴露
// github.com/yatori-dev/yatori-go-core/v3 的 yatori.Client。console 不再按平台
// 维护各自的登录/刷课循环，而是统一走：
//
//	New -> Login -> Courses -> Tree -> Study -> Works/DoWork -> Exams/DoExam
//
// 平台差异由 core 的 adapter 处理，console 只负责配置映射、节奏控制和日志输出。
//
// 已知差异（v3 公共接口暂未覆盖的旧 console 能力）：
//   - 学习通章测：v3 暂未暴露章测接口，遇到时跳过并提示；
//   - 学习通内置 AI：v3 暂未暴露，按普通 AI 答题处理；
//   - 英华去红模式、代理定时轮换、学习通多任务点并发：v3 统一 Client 不再支持，按普通/串行处理。
package runner

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/logic/progress"
	utils2 "yatori-go-console/utils"
	"yatori-go-console/utils/logbuf"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
	lg "github.com/yatori-dev/yatori-go-core/v3/utils/log"
)

var soundMut sync.Mutex

// NewClient 按 console 账号配置创建并登录一个 v3 Client。
func NewClient(ctx context.Context, setting config.Setting, user config.User) (yatori.Client, error) {
	platform, err := ParsePlatform(user.AccountType)
	if err != nil {
		return nil, err
	}
	client, err := yatori.New(yatori.Config{
		Platform: platform,
		Account: yatori.Account{
			Username: user.Account,
			Password: user.Password,
			Cookie:   strings.TrimSpace(user.Cookie),
			URL:      NormalizeURL(user.URL),
		},
		Options: buildOptions(setting, user),
	})
	if err != nil {
		return nil, err
	}
	if err := login(ctx, client, user); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func login(ctx context.Context, client yatori.Client, user config.User) error {
	cookie := resolveCookie(user)
	doLogin := func() error {
		if yatori.Supports(client, yatori.CapCookieLogin) {
			if cookie != "" {
				return client.LoginWithCookie(ctx, cookie)
			}
			if looksLikeCookie(user.Password) {
				return client.LoginWithCookie(ctx, user.Password)
			}
		}
		return client.Login(ctx)
	}
	err := callWithOCRRecover(doLogin)
	if errors.Is(err, errOCRConflict) {
		// core 里 YatoriCoreInit 与 driverutil.EnsureOCR 各持一个 sync.Once，
		// 但 onnxruntime 只能初始化一次；第二次会 panic。
		// sync.Once 在 panic 后依然标记为已完成，所以这里直接重试即可成功。
		logInfo(user, "OCR 运行时已初始化，重试登录")
		return doLogin()
	}
	return err
}

// resolveCookie 读取 cookie 字段或 cookieFile 文件；都没有时返回空串。
func resolveCookie(user config.User) string {
	if cookie := strings.TrimSpace(user.Cookie); cookie != "" {
		return cookie
	}
	file := strings.TrimSpace(user.CookieFile)
	if file == "" {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// errOCRConflict 表示 core 侧 OCR 被重复初始化（onnxruntime 只能初始化一次）。
var errOCRConflict = errors.New("ocr runtime already initialized")

// callWithOCRRecover 捕获 core 中 OCR 重复初始化抛出的 panic，转成可识别的错误。
// 其它 panic 继续向上抛，交给上层的恢复逻辑。
func callWithOCRRecover(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			message := ""
			switch value := r.(type) {
			case error:
				message = value.Error()
			case string:
				message = value
			default:
				message = fmt.Sprint(value)
			}
			if strings.Contains(strings.ToLower(message), "onnxruntime") &&
				strings.Contains(strings.ToLower(message), "already been initialized") {
				err = errOCRConflict
				return
			}
			panic(r)
		}
	}()
	return fn()
}

// looksLikeCookie 兼容旧版 console 把 Cookie / Token 直接填在 password 字段的用法。
func looksLikeCookie(value string) bool {
	if len(value) < 50 {
		return false
	}
	return strings.Contains(value, "=") || strings.Contains(value, ";")
}

func buildOptions(setting config.Setting, user config.User) yatori.Options {
	options := yatori.Options{
		Timeout: 2 * time.Minute,
	}
	if user.IsProxy == 1 && len(utils2.IPProxyPool) > 0 {
		options.ProxyURL = "http://" + utils2.RandProxyStr()
	}
	ai := setting.AiSetting
	if ai.APIKEY != "" || ai.Model != "" || ai.AiUrl != "" {
		options.AI = &yatori.AIConfig{
			Type:   yatori.AIType(ai.AiType),
			URL:    ai.AiUrl,
			Model:  ai.Model,
			APIKey: ai.APIKEY,
		}
	}
	if url := strings.TrimSpace(setting.ApiQueSetting.Url); url != "" {
		options.QuestionBank = &yatori.QuestionBankConfig{URL: url}
	}
	if sign := buildFaceSignature(user.FaceSignature); sign != nil {
		options.FaceSignature = sign
	}
	return options
}

// buildFaceSignature 把 console 配置映射成 core 的人脸签名参数。
// KeyPEM 支持内联 PEM 或从 keyPemFile 读取，避免把私钥直接写进配置文件。
func buildFaceSignature(in *config.FaceSignature) *yatori.FaceSignature {
	if in == nil {
		return nil
	}
	keyPEM := strings.TrimSpace(in.KeyPEM)
	if keyPEM == "" {
		file := strings.TrimSpace(in.KeyPEMFile)
		if file != "" {
			if data, err := os.ReadFile(file); err == nil {
				keyPEM = string(data)
			}
		}
	}
	out := &yatori.FaceSignature{
		Cxcid:  strings.TrimSpace(in.Cxcid),
		Sc:     strings.TrimSpace(in.Sc),
		KeyPEM: keyPEM,
	}
	if in.Chapter != nil {
		out.Chapter = buildFaceSignature(in.Chapter)
	}
	if out.Cxcid == "" && out.Sc == "" && out.KeyPEM == "" && out.Chapter == nil {
		return nil
	}
	return out
}

// RunUser 执行一个账号的全部课程。
func RunUser(setting config.Setting, user config.User) {
	RunUserWithStop(setting, user, nil)
}

// RunUserWithStop 执行一个账号的全部课程，stop 非空时可在节点间中断。
func RunUserWithStop(setting config.Setting, user config.User, stop *atomic.Bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()

	client, err := NewClient(ctx, setting, user)
	if err != nil {
		logError(user, "登录失败：", err.Error())
		return
	}
	defer client.Close()
	logInfo(user, lg.Green, "登录成功，平台：", client.Platform().Description())
	_ = RunClient(ctx, client, setting, user, stop)
}

// RunClient 使用一个已登录的 Client 执行该账号的全部课程。
func RunClient(ctx context.Context, client yatori.Client, setting config.Setting, user config.User, stop *atomic.Bool) error {
	courses, err := client.Courses(ctx)
	if err != nil {
		logError(user, "拉取课程列表失败：", err.Error())
		return err
	}
	logInfo(user, lg.Purple, "课程列表拉取完毕，共 ", fmt.Sprintf("%d", len(courses)), " 门课程")
	logInfo(user, "本次视频模式：", videoModeDescription(client.Platform(), user.CoursesCustom.VideoModel))
	if client.Platform() == yatori.XueXiTong && user.CoursesCustom.CxChapterTestSw != nil && *user.CoursesCustom.CxChapterTestSw == 1 {
		logInfo(user, lg.Yellow, "提示：yatori-go-core v3 暂未提供学习通章测接口，本次将跳过章测")
	}
	if client.Platform() == yatori.XueXiTong && user.CoursesCustom.AutoExam == 3 {
		logInfo(user, lg.Yellow, "提示：v3 SDK 暂未提供学习通内置AI答题入口，本次按普通AI答题处理")
	}

	// 多任务点模式：整轮任务只创建一次并发会话池，各课程复用
	var nodePool []yatori.Client
	if concurrentNodes(client.Platform(), user.CoursesCustom.VideoModel) {
		nodePool = buildWorkerClients(ctx, client, setting, user, nodeWorkerCount(user))
		defer closeWorkerClients(nodePool, client)
		logInfo(user, lg.Purple, "多任务点模式：", strconv.Itoa(len(nodePool)), " 个任务点同时推进")
	}

	var firstErr error
	if concurrentCourses(client.Platform(), user.CoursesCustom.VideoModel) {
		firstErr = runCoursesConcurrent(ctx, client, setting, user, courses, stop)
	} else {
		for _, course := range courses {
			if stopRequested(stop) {
				return firstErr
			}
			if !courseAllowed(user, course) {
				continue
			}
			if !courseDateAllowed(course) {
				logInfo(user, "【", course.Name, "】 >>> ", lg.Yellow, "未在开课时间内，已跳过")
				continue
			}
			if err := runCourse(ctx, client, setting, user, course, stop, nodePool); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	logInfo(user, lg.Purple, "所有待学习课程学习完毕")
	if setting.EmailInform.Sw == 1 && len(user.InformEmails) > 0 {
		utils2.SendMail(
			setting.EmailInform.SMTPHost,
			setting.EmailInform.SMTPPort,
			setting.EmailInform.UserName,
			setting.EmailInform.Password,
			user.InformEmails,
			fmt.Sprintf("账号：[%s]</br>平台：[%s]</br>通知：所有课程已执行完毕", user.Account, platformName(user)),
		)
	}
	if setting.BasicSetting.CompletionTone == 1 {
		soundMut.Lock()
		utils2.PlayNoticeSound()
		soundMut.Unlock()
	}
	return firstErr
}

const (
	// nodeFinishedProgress 视作“平台已经学完”的进度阈值。
	// 学习通等平台常把看完的视频回报成 99.x%，留 1% 容差避免每次重新开刷。
	nodeFinishedProgress = 99.0
	// nodeResumeProgress 以上的视频节点只补齐剩余进度，不再走完整的 start -> progress 流程。
	nodeResumeProgress = 95.0
)

// fastStudy 判断该平台该档位是否走“单次提交完成”。
//
// 老版行为：英华 2 暴力、海旗 2 秒刷、学习公社/重庆工程 2 暴力秒刷、WeLearn 2 刷完成度；
// 智慧职教/码上研训本身就是一次提交。学习通的 2/3 是多课程/多任务点，不是秒刷。
func fastStudy(platform yatori.Platform, mode int) bool {
	switch platform {
	case yatori.YingHua:
		return mode == 2
	case yatori.HaiQiKeJi, yatori.ENAAE, yatori.CQIE, yatori.WeLearn:
		return mode == 2
	}
	return singleSubmitPlatform(platform)
}

// isSerialStudyMode 判断节点之间是否需要按老版“普通模式”的节奏等待。
// 只有学习通普通模式需要，其它平台/档位都是并发或快刷。
func isSerialStudyMode(platform yatori.Platform, mode int) bool {
	return platform == yatori.XueXiTong && mode == 1
}

// studySucceeded 判断一次提交结果是否代表节点已经完成。
func studySucceeded(result *yatori.StudyResult) bool {
	if result == nil {
		return false
	}
	return result.Completed || result.Progress >= nodeFinishedProgress
}

// isNodeSkipped 不是学习节点（作业/考试/章节），由作业与考试流程单独处理。
func isNodeSkipped(node yatori.Node) bool {
	switch node.Kind {
	case yatori.NodeWork, yatori.NodeExam, yatori.NodeChapter:
		return true
	}
	return false
}

// isNodeFinished 判断学习节点是否已经完成，避免“学完的章节下次又从头来”。
func isNodeFinished(node yatori.Node) bool {
	if isNodeSkipped(node) {
		return false
	}
	return node.Completed || node.Progress >= nodeFinishedProgress
}

// isResumeNode 表示节点已经学了大半，只需要补齐剩余进度。
func isResumeNode(node yatori.Node) bool {
	if node.Kind != yatori.NodeVideo && node.Kind != yatori.NodeAudio {
		return false
	}
	return node.Progress >= nodeResumeProgress && node.Progress < nodeFinishedProgress
}

const (
	maxCourseWorkers = 3
	maxNodeWorkers   = 5
)

// concurrentCourses 对应老版“多课程模式”：多门课程同时推进（学习通 2）。
func concurrentCourses(platform yatori.Platform, mode int) bool {
	return platform == yatori.XueXiTong && mode == 2
}

// concurrentNodes 对应老版“多任务点模式”：同一课程内多个任务点同时推进（学习通 3）。
// 其它平台（含英华暴力模式）老版是单会话并发请求，v3 的单 Client 带锁，无法真正并行，
// 这里保持串行快刷，避免多开登录把会话挤掉。
func concurrentNodes(platform yatori.Platform, mode int) bool {
	return platform == yatori.XueXiTong && mode == 3
}

// singleSubmitPlatform 部分平台一次提交就完成，不需要 start -> progress -> finish 三段式。
func singleSubmitPlatform(platform yatori.Platform) bool {
	return platform == yatori.ICVE || platform == yatori.Ketangx
}

func courseWorkerCount(total int) int {
	n := maxCourseWorkers
	if total < n {
		n = total
	}
	if n < 1 {
		n = 1
	}
	return n
}

// nodeWorkerCount 对应老版的 cxNode：同时学习的任务点数量，默认 3。
func nodeWorkerCount(user config.User) int {
	n := 3
	if user.CoursesCustom.CxNode != nil && *user.CoursesCustom.CxNode > 0 {
		n = *user.CoursesCustom.CxNode
	}
	if n > maxNodeWorkers {
		n = maxNodeWorkers
	}
	return n
}

// buildWorkerClients 用已登录会话克隆并发客户端。
//
// v3 的 Client 每个方法内部都会加锁，单客户端无法真正并发，所以多课程 / 多任务点模式
// 按老版思路“多开”：每个 worker 一个独立 Client。
//   - 支持 Cookie 登录的平台（学习通等）直接复用会话，开销最小；
//   - 不支持 Cookie 登录的平台（英华等）退回一次完整登录，和老版多开一致。
func buildWorkerClients(ctx context.Context, base yatori.Client, setting config.Setting, user config.User, count int) []yatori.Client {
	if count <= 1 {
		return []yatori.Client{base}
	}
	// 没有可用的账号凭据时无法“多开”，直接按单客户端串行
	if strings.TrimSpace(user.Password) == "" {
		return []yatori.Client{base}
	}
	platform, err := ParsePlatform(user.AccountType)
	if err != nil {
		return []yatori.Client{base}
	}
	cookie := strings.TrimSpace(base.Session().Cookie)
	clients := []yatori.Client{base}
	for i := 1; i < count; i++ {
		clone, err := yatori.New(yatori.Config{
			Platform: platform,
			Account: yatori.Account{
				Username: user.Account,
				Password: user.Password,
				URL:      NormalizeURL(user.URL),
			},
			Options: buildOptions(setting, user),
		})
		if err != nil {
			logInfo(user, lg.Yellow, "创建并发客户端失败：", err.Error(), "，并发数降为 ", strconv.Itoa(len(clients)))
			break
		}
		var loginErr error
		if cookie != "" && yatori.Supports(clone, yatori.CapCookieLogin) {
			loginErr = clone.LoginWithCookie(ctx, cookie)
		} else {
			loginErr = clone.Login(ctx)
		}
		if loginErr != nil {
			_ = clone.Close()
			logInfo(user, lg.Yellow, "并发客户端登录失败：", loginErr.Error(), "，并发数降为 ", strconv.Itoa(len(clients)))
			break
		}
		// 学习通的 Cookie 登录不会带出 UserID / 课程缓存，拉一次课程列表把这些补齐，
		// 否则后续 Study 会拿到空 userID。
		if platform == yatori.XueXiTong {
			if _, err := clone.Courses(ctx); err != nil {
				_ = clone.Close()
				logInfo(user, lg.Yellow, "并发客户端初始化失败：", err.Error(), "，并发数降为 ", strconv.Itoa(len(clients)))
				break
			}
		}
		clients = append(clients, clone)
	}
	return clients
}

func closeWorkerClients(clients []yatori.Client, base yatori.Client) {
	for _, c := range clients {
		if c == nil || c == base {
			continue
		}
		_ = c.Close()
	}
}

// runCoursesConcurrent 多课程模式：所有课程并发推进，每个 worker 一个独立客户端。
func runCoursesConcurrent(ctx context.Context, base yatori.Client, setting config.Setting, user config.User, courses []yatori.Course, stop *atomic.Bool) error {
	targets := make([]yatori.Course, 0, len(courses))
	for _, course := range courses {
		if !courseAllowed(user, course) {
			continue
		}
		if !courseDateAllowed(course) {
			logInfo(user, "【", course.Name, "】 >>> ", lg.Yellow, "未在开课时间内，已跳过")
			continue
		}
		targets = append(targets, course)
	}
	if len(targets) == 0 {
		return nil
	}

	workers := buildWorkerClients(ctx, base, setting, user, courseWorkerCount(len(targets)))
	defer closeWorkerClients(workers, base)
	logInfo(user, lg.Purple, "多课程模式：", strconv.Itoa(len(workers)), " 门课程同时推进（共 ", strconv.Itoa(len(targets)), " 门）")

	jobs := make(chan yatori.Course)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for _, worker := range workers {
		wg.Add(1)
		go func(c yatori.Client) {
			defer wg.Done()
			for course := range jobs {
				if stopRequested(stop) {
					continue
				}
				if err := runCourse(ctx, c, setting, user, course, stop, nil); err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}
		}(worker)
	}
	for _, course := range targets {
		if stopRequested(stop) {
			break
		}
		jobs <- course
	}
	close(jobs)
	wg.Wait()
	return firstErr
}

// studyNodesConcurrent 多任务点模式：同一课程内的多个任务点并发推进。
func studyNodesConcurrent(ctx context.Context, pool []yatori.Client, user config.User, course yatori.Course, nodes []yatori.Node, stop *atomic.Bool) error {
	if len(pool) <= 1 {
		return nil
	}
	logInfo(user, "【", course.Name, "】 多任务点模式：", strconv.Itoa(len(pool)), " 个任务点同时推进，本次待学习 ",
		strconv.Itoa(len(nodes)), " 个")

	accountKey := progress.AccountKey(user.AccountType, user.Account)
	var index int64
	jobs := make(chan yatori.Node)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	for _, worker := range pool {
		wg.Add(1)
		go func(c yatori.Client) {
			defer wg.Done()
			for node := range jobs {
				if stopRequested(stop) {
					continue
				}
				current := atomic.AddInt64(&index, 1)
				logInfo(user, "【", course.Name, "】 节点 ", fmt.Sprintf("%d/%d", current, len(nodes)), "：", node.Name)
				completed, err := studyNode(ctx, c, user, course, node, stop)
				if completed {
					progress.MarkDone(accountKey, course.ID, node.ID)
				}
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
				}
			}
		}(worker)
	}
	for _, node := range nodes {
		if stopRequested(stop) {
			break
		}
		jobs <- node
	}
	close(jobs)
	wg.Wait()
	return firstErr
}

// yingHuaRedDelay 英华暴力模式跑完后等待一段时间再执行去红（老版是 1 分钟）。
const yingHuaRedDelay = 20 * time.Second

// clearYingHuaRedFlags 英华去红模式。
//
// 老版按节点上的 ErrorMessage（“检测到可能使用并行播放刷课”）筛选，v3 的 Node 没有暴露该字段，
// 这里按“有进度但未完成”的节点处理：重复提交当前观看位置、不推进时间，用于消除红标。
func clearYingHuaRedFlags(ctx context.Context, client yatori.Client, user config.User, course yatori.Course, stop *atomic.Bool) error {
	tree, err := client.Tree(ctx, course)
	if err != nil {
		logError(user, "【", course.Name, "】 去红：拉取章节失败：", err.Error())
		return err
	}
	var firstErr error
	handled := 0
	for _, node := range yatori.Flatten(tree) {
		if stopRequested(stop) {
			break
		}
		if node.Kind != yatori.NodeVideo && node.Kind != yatori.NodeAudio {
			continue
		}
		if node.Progress <= 0 || node.Progress >= nodeFinishedProgress {
			continue
		}
		position := time.Duration(float64(node.Duration) * clamp(node.Progress, 0, 100) / 100)
		handled++
		logInfo(user, "【", course.Name, "】 去红：", node.Name, "（当前进度 ", fmt.Sprintf("%.1f%%", node.Progress), "）")
		for i := 0; i < 3; i++ {
			if stopRequested(stop) {
				break
			}
			if _, err := studyWithRetry(ctx, client, yatori.StudyRequest{
				Course:   course,
				Node:     node,
				Action:   yatori.StudyActionProgress,
				Mode:     yatori.StudyModeNormal,
				Position: position,
			}, 2); err != nil {
				logError(user, "【", node.Name, "】 去红提交失败：", err.Error())
				if firstErr == nil {
					firstErr = err
				}
				break
			}
			time.Sleep(8 * time.Second)
		}
	}
	if handled == 0 {
		logInfo(user, "【", course.Name, "】 去红：没有发现需要处理的未完成任务点")
	} else {
		logInfo(user, "【", course.Name, "】 去红：已处理 ", strconv.Itoa(handled), " 个任务点")
	}
	return firstErr
}

func runCourse(ctx context.Context, client yatori.Client, setting config.Setting, user config.User, course yatori.Course, stop *atomic.Bool, nodePool []yatori.Client) error {
	logInfo(user, "正在学习课程：", lg.Yellow, "【"+course.Name+"】")

	// 英华去红模式：老版只处理被标记“并行播放”的视频，不学习新节点
	if client.Platform() == yatori.YingHua && user.CoursesCustom.VideoModel == 3 {
		return clearYingHuaRedFlags(ctx, client, user, course, stop)
	}

	tree, err := client.Tree(ctx, course)
	if err != nil {
		logError(user, "【", course.Name, "】 >>> 拉取章节失败：", err.Error())
		return err
	}
	nodes := yatori.Flatten(tree)
	if user.CoursesCustom.ShuffleSw == 1 {
		rand.Shuffle(len(nodes), func(i, j int) {
			nodes[i], nodes[j] = nodes[j], nodes[i]
		})
	}

	var firstErr error
	if user.CoursesCustom.VideoModel != 0 {
		// 平台的课程完成率/卡片状态有缓存，这里把平台已完成的节点和我们提交成功过的
		// 节点合并成一份本地进度，既用于跳过，也用于课程进度显示。
		accountKey := progress.AccountKey(user.AccountType, user.Account)
		studyNodes := make([]yatori.Node, 0, len(nodes))
		platformDone := make([]string, 0, len(nodes))
		for _, node := range nodes {
			if isNodeSkipped(node) {
				continue
			}
			studyNodes = append(studyNodes, node)
			if isNodeFinished(node) {
				platformDone = append(platformDone, node.ID)
			}
		}
		doneNodes := progress.MergeDone(accountKey, course.ID, len(studyNodes), platformDone)

		studyTotal := 0
		finished := 0
		for _, node := range studyNodes {
			if isNodeFinished(node) || doneNodes[node.ID] {
				finished++
				continue
			}
			studyTotal++
		}
		if finished > 0 {
			logInfo(user, "【", course.Name, "】 已跳过 ", strconv.Itoa(finished), " 个已完成节点，本次待学习 ",
				strconv.Itoa(studyTotal), " 个")
		}
		studyIndex := 0
		pending := make([]yatori.Node, 0, studyTotal)
		for _, node := range studyNodes {
			// 平台或本地记录已经完成的节点不再重复提交
			if isNodeFinished(node) || doneNodes[node.ID] {
				continue
			}
			pending = append(pending, node)
		}

		if concurrentNodes(client.Platform(), user.CoursesCustom.VideoModel) && len(nodePool) > 1 && len(pending) > 1 {
			if nodeErr := studyNodesConcurrent(ctx, nodePool, user, course, pending, stop); nodeErr != nil && firstErr == nil {
				firstErr = nodeErr
			}
		} else {
			for _, node := range pending {
				if stopRequested(stop) {
					return firstErr
				}
				studyIndex++
				logInfo(user, "【", course.Name, "】 节点 ", fmt.Sprintf("%d/%d", studyIndex, studyTotal), "：", node.Name)
				completed, nodeErr := studyNode(ctx, client, user, course, node, stop)
				if completed {
					progress.MarkDone(accountKey, course.ID, node.ID)
					doneNodes[node.ID] = true
				}
				if nodeErr != nil && firstErr == nil {
					firstErr = nodeErr
				}
				// 普通模式在节点之间随机等待，避免看起来“一口气把所有节点同时跑了”
				if isSerialStudyMode(client.Platform(), user.CoursesCustom.VideoModel) && studyIndex < studyTotal {
					waitSeconds := 10 + rand.Intn(51)
					logInfo(user, "普通模式：等待 ", strconv.Itoa(waitSeconds), " 秒后继续下一个节点")
					time.Sleep(time.Duration(waitSeconds) * time.Second)
				} else if nodeErr != nil {
					time.Sleep(5 * time.Second)
				}
			}
		}

		// 英华暴力模式：老版跑完会延迟一分钟自动执行一次去红
		if client.Platform() == yatori.YingHua && user.CoursesCustom.VideoModel == 2 && studyTotal > 0 && !stopRequested(stop) {
			time.Sleep(yingHuaRedDelay)
			if err := clearYingHuaRedFlags(ctx, client, user, course, stop); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	if user.CoursesCustom.AutoExam != 0 {
		processWorks(ctx, client, user, course, stop)
		processExams(ctx, client, user, course, stop)
	}

	logInfo(user, "【", course.Name, "】 >>> ", lg.Green, "课程处理完毕")
	return firstErr
}

func studyNode(ctx context.Context, client yatori.Client, user config.User, course yatori.Course, node yatori.Node, stop *atomic.Bool) (bool, error) {
	mode := user.CoursesCustom.VideoModel
	if mode == 0 {
		return false, nil
	}
	resume := isResumeNode(node)
	if resume {
		logInfo(user, "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Yellow,
			"继续补齐剩余进度，已完成 ", fmt.Sprintf("%.1f%%", node.Progress), "，时长：", node.Duration.String())
	} else {
		logInfo(user, "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Yellow, "开始学习，节点类型：", string(node.Kind),
			"，平台初始进度：", fmt.Sprintf("%.1f%%", node.Progress), "，时长：", node.Duration.String())
	}

	// 学习通的 2/3 是老版本的「多课程 / 多任务点」，不是秒刷，仍然走正常学习流程。
	// 其他平台的 2 档是暴力/秒刷；ICVE、码上研训一次提交即完成，也直接走单次提交。
	fast := fastStudy(client.Platform(), mode)
	if resume || fast || (node.Kind != yatori.NodeVideo && node.Kind != yatori.NodeAudio) {
		studyMode := yatori.StudyModeNormal
		if fast {
			studyMode = yatori.StudyModeFast
		}
		result, err := studyWithRetry(ctx, client, yatori.StudyRequest{
			Course:   course,
			Node:     node,
			Action:   yatori.StudyActionFinish,
			Mode:     studyMode,
			Position: node.Duration,
		}, 3)
		if err != nil {
			if yatori.IsUnsupported(err) {
				logInfo(user, "【", node.Name, "】 ", lg.Yellow, "该节点类型暂不支持自动学习，已跳过")
				return false, nil
			}
			logError(user, "【", node.Name, "】 学习失败：", err.Error())
			return false, err
		}
		logStudyResult(user, node, result)
		return studySucceeded(result), nil
	}

	// 普通模式：start -> progress -> finish。
	startResult, startErr := studyWithRetry(ctx, client, yatori.StudyRequest{
		Course: course,
		Node:   node,
		Action: yatori.StudyActionStart,
		Mode:   yatori.StudyModeNormal,
	}, 2)
	if startErr != nil && !yatori.IsUnsupported(startErr) {
		logError(user, "【", node.Name, "】 初始化学习会话失败：", startErr.Error())
		return false, startErr
	}
	if startResult != nil && strings.Contains(startResult.Message, "验证码") {
		logInfo(user, "【", node.Name, "】 ", startResult.Message)
		return false, nil
	}
	if node.Duration <= 0 && startResult != nil && startResult.Duration > 0 {
		node.Duration = startResult.Duration
		logInfo(user, "【", node.Name, "】 已从平台补齐视频时长：", node.Duration.String())
	}

	if node.Duration <= 0 {
		// 普通模式拿不到时长时不能直接 finish，否则会在没有真正学习的情况下把节点标记成 100%。
		// 常见于学习通视频仍在转码（status=waiting）或卡片没有返回 duration。
		err := fmt.Errorf("未获取到视频时长（可能平台仍在转码），普通模式下不直接标记完成")
		logError(user, "【", node.Name, "】 ", err.Error())
		return false, err
	}

	position := time.Duration(float64(node.Duration) * clamp(node.Progress, 0, 100) / 100)
	step := 30 * time.Second
	if step > node.Duration {
		step = node.Duration
	}
	for attempts := 0; attempts < 500 && position < node.Duration; attempts++ {
		if stopRequested(stop) {
			return false, nil
		}
		position += step
		if position > node.Duration {
			position = node.Duration
		}
		result, err := studyWithRetry(ctx, client, yatori.StudyRequest{
			Course:   course,
			Node:     node,
			Action:   yatori.StudyActionProgress,
			Mode:     yatori.StudyModeNormal,
			Position: position,
		}, 3)
		if err != nil {
			logError(user, "【", node.Name, "】 提交进度失败：", err.Error())
			return false, err
		}
		logStudyResult(user, node, result)
		if result.Completed || result.Progress >= 99.9 || position >= node.Duration {
			break
		}
		time.Sleep(30 * time.Second)
	}

	result, err := studyWithRetry(ctx, client, yatori.StudyRequest{
		Course:   course,
		Node:     node,
		Action:   yatori.StudyActionFinish,
		Mode:     yatori.StudyModeNormal,
		Position: node.Duration,
	}, 3)
	if err != nil {
		if !yatori.IsUnsupported(err) {
			logError(user, "【", node.Name, "】 结束学习会话失败：", err.Error())
			return false, err
		}
		return false, nil
	}
	logStudyResult(user, node, result)
	return studySucceeded(result), nil
}

// studyWithRetry 对“视频还没转码完成/临时 5xx/超时”这类可恢复错误做短重试。
// 学习通常见 status=waiting，稍等后重新拉取 dtoken 有机会成功。
func studyWithRetry(ctx context.Context, client yatori.Client, req yatori.StudyRequest, attempts int) (*yatori.StudyResult, error) {
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		result, err := client.Study(ctx, req)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if !isRetryableStudyError(err) {
			return nil, err
		}
		if i < attempts-1 {
			time.Sleep(5 * time.Second)
		}
	}
	return nil, lastErr
}

func isRetryableStudyError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "fetch failed") ||
		strings.Contains(text, "waiting") ||
		strings.Contains(text, "status code: 5") ||
		strings.Contains(text, "timeout") ||
		strings.Contains(text, "deadline exceeded")
}

func logStudyResult(user config.User, node yatori.Node, result *yatori.StudyResult) {
	if result == nil {
		return
	}
	message := result.Message
	if message == "" {
		message = "ok"
	}
	logInfo(user, "【", node.Name, "】 ", "学习进度：", lg.Green, fmt.Sprintf("%.1f%%", result.Progress), lg.Default, " ", message)
}

func processWorks(ctx context.Context, client yatori.Client, user config.User, course yatori.Course, stop *atomic.Bool) {
	if !yatori.Supports(client, yatori.CapWorks) {
		return
	}
	if client.Platform() == yatori.XueXiTong && user.CoursesCustom.CxWorkSw != nil && *user.CoursesCustom.CxWorkSw == 0 {
		return
	}
	works, err := client.Works(ctx, course)
	if err != nil {
		if !yatori.IsUnsupported(err) {
			logError(user, "【", course.Name, "】 拉取作业列表失败：", err.Error())
		}
		return
	}
	if len(works) == 0 {
		return
	}
	source := answerSource(client.Platform(), user)
	// 海旗的 DoWork 会在一次调用内处理整门课程的全部作业。
	if client.Platform() == yatori.HaiQiKeJi {
		if stopRequested(stop) {
			return
		}
		result, err := client.DoWork(ctx, yatori.WorkRequest{Course: course, Work: works[0], Source: source, Submit: true})
		logAnswerResult(user, course, "作业", "", result, err)
		return
	}
	for _, work := range works {
		if stopRequested(stop) {
			return
		}
		if isFinishedStatus(work.Status) {
			continue
		}
		result, err := client.DoWork(ctx, yatori.WorkRequest{Course: course, Work: work, Source: source, Submit: true})
		logAnswerResult(user, course, "作业", work.Name, result, err)
	}
}

func processExams(ctx context.Context, client yatori.Client, user config.User, course yatori.Course, stop *atomic.Bool) {
	if !yatori.Supports(client, yatori.CapExams) {
		return
	}
	if client.Platform() == yatori.XueXiTong && user.CoursesCustom.CxExamSw != nil && *user.CoursesCustom.CxExamSw == 0 {
		return
	}
	exams, err := client.Exams(ctx, course)
	if err != nil {
		if !yatori.IsUnsupported(err) {
			logError(user, "【", course.Name, "】 拉取考试列表失败：", err.Error())
		}
		return
	}
	if len(exams) == 0 {
		return
	}
	source := answerSource(client.Platform(), user)
	submit := user.CoursesCustom.ExamAutoSubmit == 1
	if client.Platform() == yatori.HaiQiKeJi {
		if stopRequested(stop) {
			return
		}
		result, err := client.DoExam(ctx, yatori.ExamRequest{Course: course, Exam: exams[0], Source: source, Submit: submit})
		logAnswerResult(user, course, "考试", "", result, err)
		return
	}
	for _, exam := range exams {
		if stopRequested(stop) {
			return
		}
		if !examAllowed(user, course.Name, exam.Name) || isFinishedStatus(exam.Status) {
			continue
		}
		result, err := client.DoExam(ctx, yatori.ExamRequest{Course: course, Exam: exam, Source: source, Submit: submit})
		logAnswerResult(user, course, "考试", exam.Name, result, err)
	}
}

func answerSource(platform yatori.Platform, user config.User) yatori.AnswerSource {
	switch user.CoursesCustom.AutoExam {
	case 1:
		return yatori.AnswerSourceAI
	case 2:
		if platform == yatori.HaiQiKeJi {
			return yatori.AnswerSourceBuiltin
		}
		return yatori.AnswerSourceBank
	case 3:
		// v3 SDK 暂未暴露学习通内置 AI 答题入口，按普通 AI 处理。
		return yatori.AnswerSourceAI
	default:
		return yatori.AnswerSourceManual
	}
}

func logAnswerResult(user config.User, course yatori.Course, kind, title string, result *yatori.AnswerResult, err error) {
	if err != nil {
		logError(user, "【", course.Name, "】", "【", title, "】 ", kind, "处理失败：", err.Error())
		return
	}
	if result == nil {
		logInfo(user, "【", course.Name, "】", "【", title, "】 ", kind, "处理完成")
		return
	}
	message := result.Message
	if message == "" {
		message = "ok"
	}
	logInfo(user, "【", course.Name, "】", "【", result.Title, "】 ", lg.Green, kind, "处理完成", lg.Default, " ", message)
}

func courseAllowed(user config.User, course yatori.Course) bool {
	if len(user.CoursesCustom.ExcludeCourses) > 0 && matchCourse(course, user.CoursesCustom.ExcludeCourses) {
		return false
	}
	if len(user.CoursesCustom.IncludeCourses) > 0 && !matchCourse(course, user.CoursesCustom.IncludeCourses) {
		return false
	}
	return true
}

// matchCourse 兼容 ENAEA 的 "项目名-->课程标签" 写法。
func matchCourse(course yatori.Course, patterns []string) bool {
	for _, pattern := range patterns {
		parts := strings.SplitN(pattern, "-->", 2)
		project := strings.TrimSpace(parts[0])
		if project == "" {
			continue
		}
		if len(parts) == 2 {
			tag := strings.TrimSpace(parts[1])
			if course.ExtraValue("titleTag") == tag &&
				(course.ExtraValue("projectName") == project || course.Name == project) {
				return true
			}
			continue
		}
		if course.Name == project || course.ExtraValue("projectName") == project {
			return true
		}
	}
	return false
}

func courseDateAllowed(course yatori.Course) bool {
	start := course.ExtraValue("startDate")
	end := course.ExtraValue("endDate")
	now := time.Now()
	if start != "" {
		if t, err := time.ParseInLocation("2006-01-02", start, time.Local); err == nil && now.Before(t) {
			return false
		}
	}
	if end != "" {
		if t, err := time.ParseInLocation("2006-01-02", end, time.Local); err == nil && now.After(t.Add(24*time.Hour)) {
			return false
		}
	}
	return true
}

func examAllowed(user config.User, courseName, examName string) bool {
	for _, setting := range user.CoursesCustom.CoursesSettings {
		if setting.Name != courseName {
			continue
		}
		if len(setting.ExcludeExams) > 0 && contains(setting.ExcludeExams, examName) {
			return false
		}
		if len(setting.IncludeExams) > 0 && !contains(setting.IncludeExams, examName) {
			return false
		}
	}
	return true
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if strings.TrimSpace(value) == target {
			return true
		}
	}
	return false
}

func isFinishedStatus(status string) bool {
	value := strings.ToLower(strings.TrimSpace(status))
	if value == "" {
		return false
	}
	if strings.Contains(value, "未完成") || strings.Contains(value, "未交") || strings.Contains(value, "incomplete") {
		return false
	}
	return strings.HasPrefix(value, "finish") ||
		strings.HasPrefix(value, "complete") ||
		strings.Contains(value, "已完成") ||
		strings.Contains(value, "已交")
}

func stopRequested(stop *atomic.Bool) bool {
	return stop != nil && stop.Load()
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func videoModeDescription(platform yatori.Platform, mode int) string {
	if mode == 0 {
		return "不刷（只处理作业/考试）"
	}
	if platform == yatori.XueXiTong {
		switch mode {
		case 1:
			return "普通模式（按章节顺序逐个任务点学习，节点间隔 10-60 秒）"
		case 2:
			return "多课程模式（最多 3 门课程同时推进，每门课一个独立会话）"
		case 3:
			return "多任务点模式（同一课程内多个任务点同时推进，并发数取 cxNode，默认 3）"
		}
	}
	if platform == yatori.YingHua {
		switch mode {
		case 1:
			return "普通模式（顺序学习）"
		case 2:
			return "暴力模式（快速提交任务点，跑完后自动去红）"
		case 3:
			return "去红模式（对已有进度未完成的任务点重复提交当前进度，清理红标）"
		}
	}
	switch mode {
	case 1:
		return "普通模式（顺序学习，节点间隔 10-60 秒）"
	case 2:
		switch platform {
		case yatori.HaiQiKeJi:
			return "秒刷模式（每个任务点单次提交完成）"
		case yatori.ENAAE:
			return "暴力模式（直接提交学习时长）"
		case yatori.CQIE:
			return "暴力模式（秒刷，每个任务点单次提交完成）"
		case yatori.WeLearn:
			return "刷完成度（按任务点提交完成度）"
		}
		return "快速模式（每个任务点单次提交完成）"
	case 3:
		if platform == yatori.XueXiTong {
			return "多任务点模式（v3 单客户端下按任务点依次补齐）"
		}
		return "去红模式（按普通模式兼容）"
	default:
		return fmt.Sprintf("未知模式(%d)", mode)
	}
}

func platformName(user config.User) string {
	if platform, err := ParsePlatform(user.AccountType); err == nil {
		return platform.Description()
	}
	return user.AccountType
}

func logInfo(user config.User, args ...any) {
	prefix := []any{
		fmt.Sprintf("[%s]", platformName(user)),
		"[", lg.Green, config.DisplayName(user), lg.Default, "] ",
	}
	lg.Print(lg.INFO, append(prefix, args...)...)
	appendAccountLog(user, "", args...)
}

func logError(user config.User, args ...any) {
	prefix := []any{
		fmt.Sprintf("[%s]", platformName(user)),
		"[", lg.Green, config.DisplayName(user), lg.Default, "] ",
		lg.BoldRed,
	}
	lg.Print(lg.INFO, append(prefix, args...)...)
	appendAccountLog(user, "ERROR ", args...)
}

func appendAccountLog(user config.User, level string, args ...any) {
	var builder strings.Builder
	for _, arg := range args {
		if _, ok := arg.(int); ok { // 颜色常量是 int，不写入文本日志
			continue
		}
		builder.WriteString(fmt.Sprint(arg))
	}
	line := fmt.Sprintf("[%s][%s] %s%s", platformName(user), config.DisplayName(user), level, strings.TrimSpace(builder.String()))
	logbuf.Append(logbuf.Key(user.AccountType, user.Account), line)
}
