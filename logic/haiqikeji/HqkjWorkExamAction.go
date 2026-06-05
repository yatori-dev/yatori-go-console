package haiqikeji

// ============================================================================
// 海旗科技（HQKJ）AI 自动答题接入层
//
// 本文件独立于原有视频刷课逻辑（HqkjPart.go），不修改其任何行为，仅新增「作业 + 考试」
// 的 AI 自动答题能力。整体仿照英华(yinghua) StartWorkAction/StartExamAction 的范式：
//
//   work_list / exam_list (核心库已有) → 拿 workId / examId
//        └→ work_detail / exam_detail → 取 paperId（部分作业绑定固定试卷，start 时需带上）
//        └→ work_start / exam_start (考试版核心库已有 PullExamQuestionsApi；作业版本文件对称实现) → 题面 + 选项
//             └→ qentity.Question → aiq 算答案 → qutils 匹配回选项 idx → cache.AnswerApi 提交
//
// ⚠️ 以下为「运行时假设」，已用日志暴露，实跑一次即可校准：
//   1. 提交答案的 recordId 采用 consult 返回中的 wrId；
//   2. 题型编码 mapHqkjType 目前仅确认 type=1 为单选，其余为推测；
//   3. 逐题 AnswerApi 提交后即落库，暂未接入「单独交卷」接口；
//   4. consult_list 需该作业/考试存在答题记录方能返回题目（首次可能需网页进入一次）。
// ============================================================================

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/global"

	"github.com/thedevsaddam/gojsonq"
	"github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	hqkjApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	"github.com/yatori-dev/yatori-go-core/que-core/aiq"
	"github.com/yatori-dev/yatori-go-core/que-core/qentity"
	"github.com/yatori-dev/yatori-go-core/que-core/qtype"
	"github.com/yatori-dev/yatori-go-core/utils/qutils"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

// hqkjQuiz 作业/考试条目（仅取答题所需字段）
type hqkjQuiz struct {
	Id    string // workId 或 examId
	Title string
}

// hqkjTopic 单道题目
type hqkjTopic struct {
	TopicId   string           // 题目 id（提交用）
	RecordId  string           // work_start 的整卷记录 id（考试 exam_answer_add 用）
	WrId      string           // consult 的作答记录 id（作业 work_answer_add 用）
	WaId      string           // consult 的单题答案槽 id（作业 work_answer_add 用）
	CateBid   string           // consult 题目大分类 id（work_answer_add 可能需要）
	CateMid   string           // consult 题目中分类 id（work_answer_add 可能需要）
	TypeInt   int              // 海旗原始题型编码（提交用）
	Question  qentity.Question // 交给 AI 的标准题目结构
	OptionIdx []string         // 与 Question.Options 一一对应的选项 idx（A/B/C/D）
}

// runAutoAnswer 海旗 AI 自动答题总入口（作业 + 考试）。
// 在 nodeListStudy 视频处理之后调用；仅当 AutoExam==1（AI 答题）时生效。
func runAutoAnswer(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, nodeList []haiqikeji.HqkjNode) {
	if user.CoursesCustom.AutoExam != 1 { // 0=不答题，1=AI答题（海旗目前仅接入 AI 答题）
		return
	}
	// 检测 AI 可用性：不可用则跳过答题，但不影响已完成的视频刷课（故不退出进程）
	if err := aiq.AICheck(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.APIKEY, setting.AiSetting.AiType); err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), lg.BoldRed, fmt.Sprintf("<%s>", setting.AiSetting.AiType), "AI不可用，已跳过该课程AI答题，错误信息："+err.Error())
		return
	}
	seen := map[string]bool{} // 同一作业可能挂在多个 node 下被重复拉取，按 workId 去重，避免重复作答/提交
	for _, node := range nodeList {
		workAction(setting, user, cache, course, node, seen)
		examAction(setting, user, cache, course, node, seen) // 考试 AI 答题（是否真正提交由 examDryRun 控制）
	}
}

// workAction 作业 AI 答题。seen 按 workId 去重，避免同一作业在多个 node 下被重复作答。
func workAction(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, node haiqikeji.HqkjNode, seen map[string]bool) {
	if node.TabWork <= 0 { // 过滤非作业节点
		return
	}
	listResult, err := cache.PullWorkInfoApi(course.Id, node.Id, 10, nil)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.BoldRed, "拉取作业列表失败：", err.Error())
		return
	}
	for _, q := range parseQuizList(listResult, "workInfo", "workId") {
		if seen[q.Id] { // 已处理过的作业跳过
			continue
		}
		seen[q.Id] = true
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), "【", course.Name, "】", "【", node.Name, "】 ", lg.Yellow, "正在AI自动写作业：", q.Title)
		answerQuiz(setting, user, cache, course, node, "work", q)
	}
}

// examAction 考试 AI 答题。seen 按 examId 去重（当前未启用，保留以备开启考试时使用）。
func examAction(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, node haiqikeji.HqkjNode, seen map[string]bool) {
	if node.TabExam <= 0 { // 过滤非考试节点
		return
	}
	listResult, err := cache.PullExamListApi(course.Id, node.Id, 10, nil)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.BoldRed, "拉取考试列表失败：", err.Error())
		return
	}
	for _, q := range parseQuizList(listResult, "examInfo", "examId") {
		if seen["exam:"+q.Id] { // 加 exam: 前缀，避免与作业 workId 数值碰撞被误去重
			continue
		}
		seen["exam:"+q.Id] = true
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", fmt.Sprintf("<%s>", setting.AiSetting.AiType), "【", course.Name, "】", "【", node.Name, "】 ", lg.Yellow, "正在AI自动考试：", q.Title)
		answerQuiz(setting, user, cache, course, node, "exam", q)
	}
}

// answerQuiz 对单个作业/考试执行：拉题 → AI 作答 → 逐题提交
func answerQuiz(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, node haiqikeji.HqkjNode, kind string, quiz hqkjQuiz) {
	// 第一步：拉详情，取记录列表（含每条记录的 id / paperId / classId）
	detailResult, derr := pullDetailApi(cache, kind, quiz.Id, course.Id, quiz.Title)
	if derr != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.BoldRed, "拉取作业详情失败：", derr.Error())
		return
	}
	records := parseDetailRecords(detailResult)
	var paperId, classId string
	for _, r := range records {
		if classId == "" {
			classId = r.ClassId
		}
		if paperId == "" {
			paperId = r.PaperId
		}
	}
	// 第二步：拉题。统一走 work_start（正常作答流程：它会创建可写作答会话并返回 recordId）。
	// paperId=0 的老式作业显式传 "0" 试着建会话（传空会被海旗拒"缺少必要参数: paperId"）。
	startPaperId := paperId
	if startPaperId == "" {
		startPaperId = "0"
	}
	rawResp, err := pullStartApi(cache, kind, quiz.Id, course.Id, quiz.Title, startPaperId, classId)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.BoldRed, "拉取题目失败：", err.Error())
		return
	}
	topics := parseStartQuestions(rawResp)
	// work_start 没出题时（老式作业可能不吃 paperId=0），回退 consult_list 拉题
	if len(topics) == 0 {
		for _, r := range records {
			consultResult, cerr := pullConsultApi(cache, kind, r.Id, course.Id)
			if cerr != nil {
				continue
			}
			rawResp = consultResult
			if t2 := parseStartQuestions(consultResult); len(t2) > 0 {
				topics = t2
				break
			}
		}
	}
	if len(topics) == 0 {
		// 没拉到题：作业无客观题、未开始、已截止或已答题等，打印接口原文便于判断
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", "【", node.Name, "】", "【", quiz.Title, "】 ", lg.Yellow, "跳过（未拉到客观题）：", rawResp)
		return
	}
	okCount := 0
	for _, t := range topics {
		answers := aiAnswerTopic(setting, user.Account, t)
		var subResult string
		if kind == "work" { // 作业走 work_answer_add（核心库未提供，console 层对称实现；recordId+wrId+waId 全带）——已实测打通，恒真实提交
			subResult, err = workAnswerApi(cache, course.Id, quiz.Id, t.TopicId, t.RecordId, t.WrId, t.WaId, strconv.Itoa(t.TypeInt), answers)
		} else { // 考试走 console 全字段 examAnswerApi（核心库 AnswerApi 只发 recordId、缺 wrId/waId）
			if examDryRun { // 【考试 dry-run】只打印将提交的 payload，绝不真正提交；有开放考试核对无误后把 examDryRun 改为 false 放开
				lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ",
					lg.Yellow, "【考试DRY-RUN·未提交】", examAnswerPayload(cache, course.Id, quiz.Id, t.TopicId, t.RecordId, t.WrId, t.WaId, strconv.Itoa(t.TypeInt), answers))
				continue
			}
			subResult, err = examAnswerApi(cache, course.Id, quiz.Id, t.TopicId, t.RecordId, t.WrId, t.WaId, strconv.Itoa(t.TypeInt), answers)
		}
		if err != nil {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.BoldRed, "提交答案异常：", err.Error())
			continue
		}
		if code, ok := gojsonq.New().JSONString(subResult).Find("code").(float64); ok && int(code) == 200 {
			okCount++
		} else {
			lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.Red, "提交答案返回异常：", subResult, " workId：", quiz.Id, " topicId：", t.TopicId, " wrId：", t.WrId, " waId：", t.WaId, " 题型：", strconv.Itoa(t.TypeInt), " 答案：", fmt.Sprintf("%v", answers))
		}
	}
	lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", "【", course.Name, "】", "【", node.Name, "】 >>> ", lg.Green, "AI答题完毕，成功提交 ", strconv.Itoa(okCount), "/", strconv.Itoa(len(topics)), " 题")

	// ⚠️假设3：海旗逐题提交(yee_exam_answer_add)后即落库，暂未接入「单独交卷」接口。
	if user.CoursesCustom.ExamAutoSubmit == 1 {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr[user.AccountType]), "[", lg.Green, config.DisplayAccount(user.Account), lg.Default, "] ", lg.DarkGray, "（答案已逐题提交；海旗暂未接入单独交卷接口，如需自动交卷请提供交卷接口）")
	}
}

// aiAnswerTopic 调 AI 作答单题，并把答案转换为提交格式（选择/判断→选项 idx；填空/简答→文本）
func aiAnswerTopic(setting config.Setting, account string, t hqkjTopic) []string {
	aiAnswer, err := aiq.AggregationAIApi(setting.AiSetting.AiUrl, setting.AiSetting.Model, setting.AiSetting.AiType, aiq.BuildAiQuestionMessage(t.Question), setting.AiSetting.APIKEY)
	if err != nil {
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr["HQKJ"]), "[", lg.Green, config.DisplayAccount(account), lg.Default, "] ", lg.BoldRed, "AI回答异常：", err.Error())
	}
	var contents []string
	_ = json.Unmarshal([]byte(aiAnswer), &contents)

	qt := qtype.Index(t.Question.Type)
	var answers []string
	switch qt {
	case qtype.SingleChoice, qtype.MultipleChoice, qtype.TrueOrFalse:
		for _, c := range contents {
			if idx := matchOptionIdx(c, t); idx != "" {
				answers = append(answers, idx)
			}
		}
	default: // 填空、简答、名词解释、论述等：直接用文本作为答案
		answers = append(answers, contents...)
	}

	// 兜底：AI 无法解析时给默认答案，避免漏题（仿英华策略）
	if len(answers) == 0 {
		switch qt {
		case qtype.SingleChoice, qtype.TrueOrFalse:
			if len(t.OptionIdx) > 0 {
				answers = []string{t.OptionIdx[0]}
			}
		case qtype.MultipleChoice:
			if len(t.OptionIdx) >= 2 {
				answers = []string{t.OptionIdx[0], t.OptionIdx[1]}
			} else if len(t.OptionIdx) > 0 {
				answers = []string{t.OptionIdx[0]}
			}
		}
		lg.Print(lg.INFO, fmt.Sprintf("[%s]", global.AccountTypeStr["HQKJ"]), "[", lg.Green, config.DisplayAccount(account), lg.Default, "] ", lg.BoldRed, "AI回答无法解析，题型：", t.Question.Type, "，已采用兜底答案。AI原始返回：", aiAnswer)
	}
	return answers
}

// matchOptionIdx 把 AI 返回的「选项内容」匹配回海旗选项 idx（A/B/C/D）。
// 用编辑距离相似度找最匹配的选项下标，再取其真实 idx（避免依赖 qutils 内部字母表顺序）。
func matchOptionIdx(content string, t hqkjTopic) string {
	if len(t.Question.Options) == 0 {
		return ""
	}
	best, bestScore := 0, -1.0
	for i, opt := range t.Question.Options {
		if s := qutils.Similarity(opt, content); s > bestScore {
			bestScore = s
			best = i
		}
	}
	if best < len(t.OptionIdx) {
		return t.OptionIdx[best]
	}
	return ""
}

// mapHqkjType 海旗题型编码(int) → qtype 中文题型（用于构造喂给 AI 的题目）。
// ⚠️假设2：目前仅确认 type=1 为单选；2/3/4/5 为按常见编码的推测，遇未知类型默认按单选处理。
func mapHqkjType(t int) string {
	switch t {
	case 1:
		return qtype.SingleChoice.String() // 单选题
	case 2:
		return qtype.MultipleChoice.String() // 多选题
	case 3:
		return qtype.TrueOrFalse.String() // 判断题
	case 4:
		return qtype.FillInTheBlank.String() // 填空题
	case 5:
		return qtype.ShortAnswer.String() // 简答题
	default:
		return qtype.SingleChoice.String()
	}
}

// parseQuizList 解析作业/考试列表，提取 id 与标题。
// 作业：arrKey=workInfo、idKey=workId；考试：arrKey=examInfo、idKey=examId。
func parseQuizList(jsonStr, arrKey, idKey string) []hqkjQuiz {
	var res []hqkjQuiz
	arr, ok := gojsonq.New().JSONString(jsonStr).Find("data." + arrKey).([]any)
	if !ok {
		return res
	}
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		var q hqkjQuiz
		if v, ok := m[idKey].(float64); ok {
			q.Id = strconv.Itoa(int(v))
		}
		if v, ok := m["title"].(string); ok {
			q.Title = v
		}
		if q.Id != "" {
			res = append(res, q)
		}
	}
	return res
}

// parseStartQuestions 解析 work_start/exam_start 返回的题目。
// 实测题目数组为 data.workTopics[]，每题字段：id(题目id→提交用 topicId)/recordId(整卷记录id→提交用)/
// type(题型 1单选 3判断…)/topic(题面,含HTML标签)/option[]{idx,answer,scale}。
// ⚠️按用户要求：只取题面+选项喂 AI，scale(=100 即正确答案标记)一律忽略，绝不抄答案。
func parseStartQuestions(jsonStr string) []hqkjTopic {
	var topics []hqkjTopic
	var arr []any
	for _, key := range []string{"data.workTopics", "data.examTopics", "data.workResult"} {
		if v, ok := gojsonq.New().JSONString(jsonStr).Find(key).([]any); ok {
			arr = v
			break
		}
	}
	for _, it := range arr {
		m, ok := it.(map[string]any)
		if !ok {
			continue
		}
		var t hqkjTopic
		if v, ok := m["id"].(float64); ok { // work_start: 题目 id → 提交用 topicId
			t.TopicId = strconv.Itoa(int(v))
		}
		if v, ok := m["topicId"].(float64); ok { // consult_list: topicId → 提交用 topicId
			t.TopicId = strconv.Itoa(int(v))
		}
		if v, ok := m["recordId"].(float64); ok { // work_start: 整卷记录 id
			t.RecordId = strconv.Itoa(int(v))
		}
		if v, ok := m["wrId"].(float64); ok { // consult_list: 作答记录 id
			t.WrId = strconv.Itoa(int(v))
		}
		if v, ok := m["waId"].(float64); ok { // consult_list: 单题答案槽 id
			t.WaId = strconv.Itoa(int(v))
		}
		if v, ok := m["cateBid"].(float64); ok { // consult_list: 题目大分类 id
			t.CateBid = strconv.Itoa(int(v))
		}
		if v, ok := m["cateMid"].(float64); ok { // consult_list: 题目中分类 id
			t.CateMid = strconv.Itoa(int(v))
		}
		if v, ok := m["type"].(float64); ok {
			t.TypeInt = int(v)
		}
		if v, ok := m["topic"].(string); ok {
			t.Question.Content = stripHTML(v)
		}
		t.Question.Type = mapHqkjType(t.TypeInt)
		if opts, ok := m["option"].([]any); ok {
			for _, o := range opts {
				om, ok := o.(map[string]any)
				if !ok {
					continue
				}
				ans, _ := om["answer"].(string)
				idx, _ := om["idx"].(string)
				t.Question.Options = append(t.Question.Options, ans)
				t.OptionIdx = append(t.OptionIdx, idx)
			}
		}
		if t.TopicId != "" && len(t.Question.Options) > 0 {
			topics = append(topics, t)
		}
	}
	return topics
}

// htmlTagRe 匹配 HTML 标签，用于清洗题面。
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// stripHTML 去掉题面里的 HTML 标签与零宽/实体字符，得到纯文本喂给 AI。
func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "&nbsp;", " ")
	s = strings.ReplaceAll(s, "Feff", "")
	return strings.TrimSpace(s)
}

// pullDetailApi 发起作业/考试「详情(拉题)」查询：work_detail / exam_detail。
// 注意：consult_list 是「查看已答记录」接口（未作答返回空 workResult），真正拉题需用 detail 接口。
// 核心库未提供作业 detail，这里在 console 层自行实现，复用 cache 会话信息（Token/SchoolId/UserId/PreUrl/代理）。
func pullDetailApi(cache *hqkjApi.HqkjUserCache, kind, quizId, courseId, title string) (string, error) {
	var path, idKey string
	if kind == "work" {
		path = "/api/user/yee_course_student_work_detail"
		idKey = "workId"
	} else {
		path = "/api/user/yee_course_student_exam_detail"
		idKey = "examId"
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	urlStr := cache.PreUrl + path + "?schoolId=" + cache.SchoolId + "&studentId=" + cache.UserId + "&" + idKey + "=" + quizId + "&courseId=" + courseId + "&title=" + url.QueryEscape(title)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// pullStartApi 调海旗「开始作业/考试」接口正式拉题（work_start / exam_start）。
// 考试版核心库已实现并自带空 paperId 的 yee_course_student_exam_start(PullExamQuestionsApi)，直接复用；
// 作业版核心库未提供，这里按 exam_start 的对称范式实现 yee_course_student_work_start，复用 cache 会话与代理设置。
func pullStartApi(cache *hqkjApi.HqkjUserCache, kind, quizId, courseId, title, paperId, classId string) (string, error) {
	if kind == "exam" {
		return cache.PullExamQuestionsApi(quizId, courseId, title, 10, nil)
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	urlStr := cache.PreUrl + "/api/user/yee_course_student_work_start?schoolId=" + cache.SchoolId +
		"&studentId=" + cache.UserId + "&courseId=" + courseId + "&workId=" + quizId +
		"&platform=pc&createUserId=" + cache.UserId + "&state=1&classId=" + classId + "&paperId=" + paperId +
		"&random=&randData=%257B%257D&randNumber="
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// hqkjRecord work_detail/exam_detail 里的一条答题记录。
type hqkjRecord struct {
	Id      string // 记录 id（paperId=0 的老式作业用它走 consult_list 拉题）
	PaperId string // 固定试卷 id（>0 时走 work_start 拉题）
	ClassId string // 班级 id（work_start 必填）
}

// parseDetailRecords 解析 work_detail/exam_detail 的记录列表（id / paperId>0 / classId>0）。
func parseDetailRecords(jsonStr string) []hqkjRecord {
	var recs []hqkjRecord
	for _, arrKey := range []string{"data.workDetail", "data.examDetail"} {
		arr, ok := gojsonq.New().JSONString(jsonStr).Find(arrKey).([]any)
		if !ok {
			continue
		}
		for _, it := range arr {
			m, ok := it.(map[string]any)
			if !ok {
				continue
			}
			var r hqkjRecord
			if v, ok := m["id"].(float64); ok {
				r.Id = strconv.Itoa(int(v))
			}
			if v, ok := m["paperId"].(float64); ok && int(v) > 0 {
				r.PaperId = strconv.Itoa(int(v))
			}
			if v, ok := m["classId"].(float64); ok && int(v) > 0 {
				r.ClassId = strconv.Itoa(int(v))
			}
			if r.Id != "" {
				recs = append(recs, r)
			}
		}
	}
	return recs
}

// pullConsultApi 调海旗「答题记录查看」接口 consult_list 拉题（用于 paperId=0 的老式作业/考试）。
// 参数位上的 workId/examId 实际传 work_detail 里的记录 id；返回题目在 data.workResult[] 下。
func pullConsultApi(cache *hqkjApi.HqkjUserCache, kind, recordId, courseId string) (string, error) {
	var path, idKey string
	if kind == "work" {
		path = "/api/user/yee_work_record_consult_list"
		idKey = "workId"
	} else {
		path = "/api/user/yee_exam_record_consult_list"
		idKey = "examId"
	}
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	urlStr := cache.PreUrl + path + "?schoolId=" + cache.SchoolId + "&userId=" + cache.UserId + "&" + idKey + "=" + recordId + "&courseId=" + courseId
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// workAnswerApi 提交作业答案。核心库只有考试版 yee_exam_answer_add，这里按对称范式实现作业版
// yee_work_answer_add；payload 在考试版基础上把 examId 换 workId，并补上 consult 拿到的 wrId/waId。
func workAnswerApi(cache *hqkjApi.HqkjUserCache, courseId, workId, topicId, recordId, wrId, waId, qType string, answers []string) (string, error) {
	// recordId 与 wrId 是同一“作答记录id”在 workTopics/workResult 两种返回里的不同字段名，互补取非空值
	if recordId == "" {
		recordId = wrId
	}
	if wrId == "" {
		wrId = recordId
	}
	if recordId == "" { // 兜底，避免拼出非法 JSON
		recordId = "0"
	}
	if wrId == "" {
		wrId = "0"
	}
	if waId == "" {
		waId = "0"
	}
	answersData, _ := json.Marshal(answers)
	// 三次实测铁律：payload 必须带 recordId（缺则服务端“系统异常”），其值=wrId；并保留 wrId/waId 给全字段最大化成功率。
	payload := `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"workId":` + workId + `,"recordId":` + recordId + `,"wrId":` + wrId + `,"waId":` + waId +
		`,"topicId":` + topicId + `,"answer":` + string(answersData) + `,"type":` + qType + `}`
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/yee_work_answer_add", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// examDryRun 考试专用 dry-run 开关：true=考试只打印将提交的 payload、绝不真正发请求（验证用）；false=真实提交。
// 作业已实测打通、恒真实提交，不受此开关影响。考试限时且疑似只能交一次，故先以 dry-run 验证拉题/AI/payload
// 全部无误，待账号有开放考试、核对 payload 无误后再改为 false 放开真实提交。
const examDryRun = true

// examAnswerPayload 构造考试提交 payload。与已实测打通的作业 workAnswerApi 对称（仅 workId→examId、接口路径不同）：
// 考试 consult 返回 workResult 同样给 wrId/waId、无 recordId，故 recordId↔wrId 互补取非空值、并带全字段。
func examAnswerPayload(cache *hqkjApi.HqkjUserCache, courseId, examId, topicId, recordId, wrId, waId, qType string, answers []string) string {
	if recordId == "" {
		recordId = wrId
	}
	if wrId == "" {
		wrId = recordId
	}
	if recordId == "" { // 兜底，避免拼出非法 JSON
		recordId = "0"
	}
	if wrId == "" {
		wrId = "0"
	}
	if waId == "" {
		waId = "0"
	}
	answersData, _ := json.Marshal(answers)
	return `{"schoolId":` + cache.SchoolId + `,"courseId":` + courseId + `,"userId":` + cache.UserId +
		`,"examId":` + examId + `,"recordId":` + recordId + `,"wrId":` + wrId + `,"waId":` + waId +
		`,"topicId":` + topicId + `,"answer":` + string(answersData) + `,"type":` + qType + `}`
}

// examAnswerApi 提交考试答案（yee_exam_answer_add）。核心库 AnswerApi 只发 recordId、缺 wrId/waId，
// 这里仿已验证的作业版 workAnswerApi 带全字段，复用 cache 会话与代理设置。
func examAnswerApi(cache *hqkjApi.HqkjUserCache, courseId, examId, topicId, recordId, wrId, waId, qType string, answers []string) (string, error) {
	payload := examAnswerPayload(cache, courseId, examId, topicId, recordId, wrId, waId, qType, answers)
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	if cache.IpProxySW { // 复用账号的 IP 代理设置
		tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse("http://" + cache.ProxyIP)
		}
	}
	client := &http.Client{Timeout: 30 * time.Second, Transport: tr}
	req, err := http.NewRequest("POST", cache.PreUrl+"/api/user/yee_exam_answer_add", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("authorization", cache.Token)
	req.Header.Set("content-type", "application/json")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
