package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/entity/vo"
	"yatori-go-console/global"
	"yatori-go-console/logic/progress"
	"yatori-go-console/logic/runner"
	"yatori-go-console/utils"
	"yatori-go-console/utils/logbuf"
	"yatori-go-console/web/activity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

// 拉取账号列表
func UserListService(c *gin.Context) {
	page := parseQueryInt(c.Query("page"), 1)
	pageSize := parseQueryInt(c.Query("pageSize"), 100)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 500 {
		pageSize = 100
	}
	users, total, err := dao.QueryUsers(global.GlobalDB, page, pageSize)
	if err != nil {
		fail(c, err.Error())
		return
	}
	resUserList := make([]map[string]any, 0, len(users))
	for _, user := range users {
		item := utils.StructToMap(user)
		parsed := pojoToUser(user)
		item["config"] = parsed
		item["isRunning"] = false
		item["lastError"] = ""
		item["status"] = activity.Status{UID: user.Uid, State: activity.StateIdle}
		if userActivity := global.GetUserActivity(user); userActivity != nil {
			if running, ok := (*userActivity).(activity.RunningAbility); ok {
				item["isRunning"] = running.Running()
			}
			if state, ok := (*userActivity).(activity.StatusAbility); ok {
				status := state.Status()
				item["status"] = status
				item["lastError"] = status.LastError
			}
		}
		resUserList = append(resUserList, item)
	}
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "拉取账号成功",
		Data: gin.H{
			"users": resUserList,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// 添加账号
func AddUserService(c *gin.Context) {
	var req vo.AddAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "请求参数错误: "+err.Error())
		return
	}
	user := req.User
	if err := normalizeAccount(&user, true); err != nil {
		fail(c, err.Error())
		return
	}
	// 同平台 + 同 URL + 同账号视为重复
	cond := pojo.UserPO{AccountType: user.AccountType, Url: user.URL, Account: user.Account}
	if existing, _ := dao.QueryUser(global.GlobalDB, cond); existing != nil {
		fail(c, "该账号已存在")
		return
	}

	uid, err := uuid.NewV7()
	if err != nil {
		fail(c, "生成账号 ID 失败: "+err.Error())
		return
	}
	userPO := pojo.UserPO{
		Uid:         uid.String(),
		AccountType: user.AccountType,
		Url:         user.URL,
		Account:     user.Account,
		Password:    user.Password,
	}
	if err := writeUserConfig(&userPO, user); err != nil {
		fail(c, "保存账号配置失败: "+err.Error())
		return
	}
	if err := dao.InsertUser(global.GlobalDB, &userPO); err != nil {
		fail(c, err.Error())
		return
	}
	ok(c, "添加账号成功", gin.H{"user": userPO, "config": user})
}

// 删除账号
func DeleteUserService(c *gin.Context) {
	var req vo.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "数据转换失败")
		return
	}
	var user *pojo.UserPO
	var err error
	if req.Uid != "" {
		user, err = dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: req.Uid})
	} else if req.AccountType != "" && req.Account != "" {
		user, err = dao.QueryUser(global.GlobalDB, pojo.UserPO{
			AccountType: req.AccountType,
			Url:         req.Url,
			Account:     req.Account,
		})
	} else {
		fail(c, "缺少 uid 或账号信息")
		return
	}
	if err != nil || user == nil {
		fail(c, "账号不存在")
		return
	}
	// 先停止运行中的任务，再删除数据库记录和内存运行态
	if userActivity := global.GetUserActivity(*user); userActivity != nil {
		_ = (*userActivity).Stop()
		if closer, ok := (*userActivity).(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}
	global.DeleteUserActivity(user.Uid)
	logbuf.Clear(logbuf.Key(user.AccountType, user.Account))
	progress.Clear(progress.AccountKey(user.AccountType, user.Account))
	if err := dao.DeleteUser(global.GlobalDB, &pojo.UserPO{Uid: user.Uid}); err != nil {
		fail(c, err.Error())
		return
	}
	ok(c, "删除成功", nil)
}

// 检查账号登录是否正常
func AccountLoginCheckService(c *gin.Context) {
	var req vo.AccountLoginCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "请求参数错误: "+err.Error())
		return
	}
	var user config.User
	if req.Uid != "" {
		userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: req.Uid})
		if err != nil || userPO == nil {
			fail(c, "该账号不存在")
			return
		}
		user = pojoToUser(*userPO)
	} else {
		user = config.User{
			AccountType: req.AccountType,
			URL:         req.Url,
			Account:     req.Account,
			Password:    req.Password,
		}
	}
	if err := normalizeAccount(&user, true); err != nil {
		fail(c, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client, err := runner.NewClient(ctx, global.Setting, user)
	if err != nil {
		fail(c, "登录失败: "+err.Error())
		return
	}
	platform := client.Platform().String()
	_ = client.Close()
	ok(c, "账号登录正常", gin.H{"platform": platform})
}

// 获取账号配置信息
func GetAccountInformService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "该账号不存在")
		return
	}
	running := false
	lastError := ""
	status := activity.Status{UID: uid, State: activity.StateIdle}
	if userActivity := global.GetUserActivity(*userPO); userActivity != nil {
		if state, ok1 := (*userActivity).(activity.RunningAbility); ok1 {
			running = state.Running()
		}
		if state, ok1 := (*userActivity).(activity.StatusAbility); ok1 {
			status = state.Status()
			lastError = status.LastError
		}
	}
	ok(c, "拉取信息成功", gin.H{"user": userPO, "config": pojoToUser(*userPO), "isRunning": running, "lastError": lastError, "status": status})
}

// AccountStatusService 读取账号任务状态，供前端轮询。
func AccountStatusService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "账号不存在")
		return
	}
	status := activity.Status{UID: uid, State: activity.StateIdle}
	if userActivity := global.GetUserActivity(*userPO); userActivity != nil {
		if state, ok1 := (*userActivity).(activity.StatusAbility); ok1 {
			status = state.Status()
		}
	}
	ok(c, "读取状态成功", gin.H{"status": status})
}

// AccountLogService 读取账号最近的运行日志。
func AccountLogService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "账号不存在")
		return
	}
	limit := parseQueryInt(c.Query("limit"), 300)
	lines := logbuf.Tail(logbuf.Key(userPO.AccountType, userPO.Account), limit)
	ok(c, "读取日志成功", gin.H{"lines": lines})
}

// 登录账号（保留接口，实际登录由任务启动/登录检测完成）
func LoginUserService(c *gin.Context) {
	ok(c, "登录成功", nil)
}

// 更新账号信息
func UpdateUserService(c *gin.Context) {
	var req vo.UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "JSON 解析失败: "+err.Error())
		return
	}
	if req.Uid == "" {
		fail(c, "UID 不能为空")
		return
	}
	existingPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: req.Uid})
	if err != nil || existingPO == nil {
		fail(c, "账号不存在")
		return
	}
	if userActivity := global.GetUserActivity(*existingPO); userActivity != nil {
		if running, ok := (*userActivity).(activity.RunningAbility); ok && running.Running() {
			fail(c, "任务运行中，请先停止任务再修改配置")
			return
		}
		if closer, ok := (*userActivity).(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	}

	existing := pojoToUser(*existingPO)
	merged := mergeUser(existing, req.User)
	if err := normalizeAccount(&merged, false); err != nil {
		fail(c, err.Error())
		return
	}
	if err := writeUserConfig(existingPO, merged); err != nil {
		fail(c, "保存账号配置失败: "+err.Error())
		return
	}
	updateMap := map[string]any{
		"account_type":     merged.AccountType,
		"url":              merged.URL,
		"account":          merged.Account,
		"password":         merged.Password,
		"user_config_json": existingPO.UserConfigJson,
	}
	if err := dao.UpdateUser(global.GlobalDB, req.Uid, updateMap); err != nil {
		fail(c, err.Error())
		return
	}
	// 配置已变化，销毁旧运行态，下次启动重新登录
	global.DeleteUserActivity(req.Uid)
	existingPO.AccountType = merged.AccountType
	existingPO.Url = merged.URL
	existingPO.Account = merged.Account
	existingPO.Password = merged.Password
	ok(c, "更新成功", gin.H{"user": existingPO, "config": merged})
}

// 拉取课程列表
func AccountCourseListService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "该账号不存在")
		return
	}
	userActivity := getOrCreateActivity(userPO)
	if userActivity == nil {
		fail(c, "该平台暂不支持拉取课程")
		return
	}
	courseAbility, ok1 := (*userActivity).(activity.CourseAbility)
	if !ok1 {
		fail(c, "该平台暂不支持拉取课程")
		return
	}
	list, err := courseAbility.PullCourseList()
	if err != nil {
		fail(c, "拉取课程失败: "+err.Error())
		return
	}
	// 平台课程完成率有缓存（刚刷完的节点可能还是旧值），这里用本地节点完成记录向上覆盖
	accountKey := progress.AccountKey(userPO.AccountType, userPO.Account)
	for i := range list {
		done, total, okLocal := progress.Stats(accountKey, list[i].ID)
		if !okLocal || total <= 0 {
			continue
		}
		if local := float64(done) / float64(total) * 100; local > list[i].Progress {
			list[i].Progress = local
		}
	}
	courseList := make([]vo.CourseInformResponse, 0, len(list))
	for _, course := range list {
		courseList = append(courseList, vo.CourseInformResponse{
			CourseId:   course.ID,
			CourseName: course.Name,
			Progress:   float32(course.Progress),
			Instructor: course.Teacher,
		})
	}
	ok(c, "拉取信息成功", gin.H{"courseList": courseList})
}

// 开始刷课
func StartBrushService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "账号不存在")
		return
	}
	userActivity := getOrCreateActivity(userPO)
	if userActivity == nil {
		fail(c, "该平台暂不支持启动任务")
		return
	}
	if running, ok1 := (*userActivity).(activity.RunningAbility); ok1 && running.Running() {
		fail(c, "任务已在运行")
		return
	}

	var req vo.StartTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		fail(c, "请求参数错误: "+err.Error())
		return
	}
	// 同步登录一次，登录失败直接反馈给前端（避免接口先返回成功、任务随后静默退出）
	if err := (*userActivity).Login(); err != nil {
		fail(c, "登录失败: "+err.Error())
		return
	}
	go func() {
		if starter, ok1 := (*userActivity).(activity.CourseRunner); ok1 {
			_ = starter.StartWithCourses(req.Courses)
			return
		}
		_ = (*userActivity).Start()
	}()
	ok(c, "启动成功", gin.H{"courses": req.Courses})
}

// 停止任务
func StopBrushService(c *gin.Context) {
	uid := c.Param("uid")
	userPO, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if err != nil || userPO == nil {
		fail(c, "账号不存在")
		return
	}
	userActivity := global.GetUserActivity(*userPO)
	if userActivity == nil {
		fail(c, "任务未运行")
		return
	}
	if err := (*userActivity).Stop(); err != nil {
		fail(c, "停止失败: "+err.Error())
		return
	}
	ok(c, "停止成功", nil)
}

var activityInitMu sync.Mutex

// getOrCreateActivity 保证同一个账号只创建一个运行态，避免首次打开详情时
// 课程列表/状态/启动请求并发触发多次登录。
func getOrCreateActivity(userPO *pojo.UserPO) *activity.Activity {
	activityInitMu.Lock()
	defer activityInitMu.Unlock()

	if userActivity := global.GetUserActivity(*userPO); userActivity != nil {
		return userActivity
	}
	created := activity.BuildUserActivity(*userPO)
	if created == nil {
		return nil
	}
	userActivity := &created
	global.PutUserActivity(*userPO, userActivity)
	return userActivity
}

func pojoToUser(po pojo.UserPO) config.User {
	user := config.User{}
	if po.UserConfigJson != "" {
		_ = json.Unmarshal([]byte(po.UserConfigJson), &user)
	}
	if user.AccountType == "" {
		user.AccountType = po.AccountType
	}
	if user.URL == "" {
		user.URL = po.Url
	}
	if user.Account == "" {
		user.Account = po.Account
	}
	if user.Password == "" {
		user.Password = po.Password
	}
	user.URL = runner.NormalizeURL(user.URL)
	config.NormalizeUser(&user)
	return user
}

// mergeUser 用请求中的字段覆盖已有账号配置，空密码表示保留原密码。
func mergeUser(existing, incoming config.User) config.User {
	merged := incoming
	if merged.AccountType == "" {
		merged.AccountType = existing.AccountType
	}
	if merged.URL == "" {
		merged.URL = existing.URL
	}
	if merged.Account == "" {
		merged.Account = existing.Account
	}
	if merged.Password == "" {
		merged.Password = existing.Password
	}
	if merged.InformEmails == nil {
		merged.InformEmails = existing.InformEmails
	}
	if reflect.DeepEqual(merged.CoursesCustom, config.CoursesCustom{}) {
		merged.CoursesCustom = existing.CoursesCustom
	}
	return merged
}

func normalizeAccount(user *config.User, requirePassword bool) error {
	platform, err := runner.ParsePlatform(user.AccountType)
	if err != nil {
		return fmt.Errorf("不支持的平台: %s", user.AccountType)
	}
	if platform == yatori.HaiQiKeJi {
		user.AccountType = "HQKJ"
	} else {
		user.AccountType = string(platform)
	}
	if platform == yatori.YingHua || platform == yatori.HaiQiKeJi {
		if strings.TrimSpace(user.URL) == "" {
			return errors.New("该平台必须填写 URL")
		}
		normalized, err := normalizeURL(user.URL)
		if err != nil {
			return err
		}
		user.URL = normalized
	}
	config.NormalizeUser(user)
	if user.Account == "" {
		return errors.New("账号不能为空")
	}
	if requirePassword && user.Password == "" {
		return errors.New("密码/Cookie/Token 不能为空")
	}
	return nil
}

func normalizeURL(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "", errors.New("URL 不能为空")
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return "", errors.New("URL 格式不正确")
	}
	return parsed.Scheme + "://" + parsed.Host, nil
}

func writeUserConfig(po *pojo.UserPO, user config.User) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	po.UserConfigJson = string(data)
	return nil
}

func parseQueryInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	var value int
	if _, err := fmt.Sscanf(raw, "%d", &value); err != nil {
		return fallback
	}
	return value
}

func ok(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, vo.Response{Code: 200, Message: message, Data: data})
}

func fail(c *gin.Context, message string) {
	c.JSON(http.StatusOK, vo.Response{Code: 400, Message: message})
}
