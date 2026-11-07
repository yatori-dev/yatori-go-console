package activity

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/logic/runner"

	"github.com/google/uuid"
	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

// 默认运行配置，由 logic.Lunch 在进入 Web 模式前注入。
var defaultSetting config.Setting

// SetDefaultSetting 设置 Web 模式下账号启动时使用的运行配置。
func SetDefaultSetting(setting config.Setting) {
	defaultSetting = setting
}

// State 账号任务状态。
type State string

const (
	StateIdle      State = "idle"
	StateStarting  State = "starting"
	StateRunning   State = "running"
	StateStopping  State = "stopping"
	StateSucceeded State = "succeeded"
	StateFailed    State = "failed"
	StateCanceled  State = "canceled"
)

// Status 账号运行态快照，供 API 与前端轮询使用。
type Status struct {
	UID        string    `json:"uid"`
	State      State     `json:"state"`
	TaskID     string    `json:"taskId,omitempty"`
	Courses    []string  `json:"courses,omitempty"`
	StartedAt  time.Time `json:"startedAt,omitempty"`
	FinishedAt time.Time `json:"finishedAt,omitempty"`
	LastError  string    `json:"lastError,omitempty"`
}

func (s Status) running() bool {
	return s.State == StateStarting || s.State == StateRunning || s.State == StateStopping
}

// Activity 账号运行态抽象。
type Activity interface {
	Login() error
	Start() error
	Stop() error
	GetUserCache() any
	SetUser(user config.User)
	GetUser() config.User
}

// CourseAbility 暴露课程列表能力，供 Web API 使用。
type CourseAbility interface {
	PullCourseList() ([]yatori.Course, error)
}

// RunningAbility 暴露运行状态。
type RunningAbility interface {
	Running() bool
}

// StatusAbility 暴露完整任务状态。
type StatusAbility interface {
	Status() Status
}

// ErrorAbility 暴露最近一次任务错误，供 Web 端提示。
type ErrorAbility interface {
	LastError() string
}

// CourseRunner 支持本次只执行指定课程（课程名称）。
type CourseRunner interface {
	StartWithCourses(courses []string) error
}

// UserActivityBase 通用账号运行态。
type UserActivityBase struct {
	User      config.User
	UserCache any

	statusMu sync.RWMutex
	status   Status
}

func (u *UserActivityBase) SetUser(user config.User) { u.User = user }
func (u *UserActivityBase) GetUser() config.User     { return u.User }
func (u *UserActivityBase) GetUserCache() any        { return u.UserCache }

// Status 返回运行态快照。
func (u *UserActivityBase) Status() Status {
	u.statusMu.RLock()
	defer u.statusMu.RUnlock()
	out := u.status
	out.Courses = append([]string(nil), u.status.Courses...)
	return out
}

func (u *UserActivityBase) setStatus(state State, taskID string, courses []string, err error) {
	u.statusMu.Lock()
	defer u.statusMu.Unlock()
	now := time.Now()
	u.status.State = state
	if u.status.UID == "" {
		u.status.UID = u.User.Account
	}
	if taskID != "" {
		u.status.TaskID = taskID
	}
	if state == StateStarting {
		u.status.Courses = append([]string(nil), courses...)
	}
	switch state {
	case StateStarting:
		u.status.StartedAt = now
		u.status.FinishedAt = time.Time{}
		u.status.LastError = ""
	case StateRunning:
		u.status.LastError = ""
	case StateSucceeded, StateFailed, StateCanceled:
		u.status.FinishedAt = now
		if err != nil {
			u.status.LastError = err.Error()
		}
	}
}

func (u *UserActivityBase) setStatusError(err error) {
	u.statusMu.Lock()
	defer u.statusMu.Unlock()
	if err == nil {
		u.status.LastError = ""
		return
	}
	u.status.LastError = err.Error()
}

func (u *UserActivityBase) Running() bool {
	return u.Status().running()
}

// LastError 返回最近一次任务错误。
func (u *UserActivityBase) LastError() string {
	return u.Status().LastError
}

// PlatformActivity 基于 yatori-go-core v3 统一 Client 的账号运行态。
type PlatformActivity struct {
	UserActivityBase

	client yatori.Client
	stop   atomic.Bool
	mu     sync.Mutex
}

// BuildUserActivity 根据数据库账号配置构造运行态；平台不支持时返回 nil。
func BuildUserActivity(po pojo.UserPO) Activity {
	user := config.User{}
	if err := json.Unmarshal([]byte(po.UserConfigJson), &user); err != nil {
		return nil
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
	if _, err := runner.ParsePlatform(user.AccountType); err != nil {
		return nil
	}
	activity := &PlatformActivity{
		UserActivityBase: UserActivityBase{
			User: user,
			status: Status{
				UID:   po.Uid,
				State: StateIdle,
			},
		},
	}
	return activity
}

// Login 创建并登录 v3 Client。
func (activity *PlatformActivity) Login() error {
	activity.mu.Lock()
	defer activity.mu.Unlock()
	if activity.client != nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	client, err := runner.NewClient(ctx, defaultSetting, activity.User)
	if err != nil {
		return err
	}
	activity.client = client
	activity.UserCache = client
	return nil
}

// Start 启动该账号的刷课任务。
func (activity *PlatformActivity) Start() error {
	return activity.StartWithCourses(nil)
}

// StartWithCourses 启动任务，courses 非空时只执行这些课程（课程名称）。
func (activity *PlatformActivity) StartWithCourses(courses []string) error {
	if activity.Running() {
		return runnerError("任务已在运行")
	}
	activity.stop.Store(false)

	taskID := uuid.NewString()
	activity.setStatus(StateStarting, taskID, courses, nil)
	if err := activity.Login(); err != nil {
		activity.setStatus(StateFailed, taskID, courses, err)
		return err
	}
	if activity.stop.Load() {
		activity.setStatus(StateCanceled, taskID, courses, nil)
		return nil
	}
	activity.setStatus(StateRunning, taskID, courses, nil)

	user := activity.User
	// Web 端的课程选择是唯一入口：不选课程 = 全部课程，同时忽略旧的 include/exclude 配置。
	user.CoursesCustom.IncludeCourses = append([]string(nil), courses...)
	user.CoursesCustom.ExcludeCourses = nil
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)
	defer cancel()
	err := runner.RunClient(ctx, activity.client, defaultSetting, user, &activity.stop)
	if activity.stop.Load() {
		activity.setStatus(StateCanceled, taskID, courses, nil)
	} else if err != nil {
		activity.setStatus(StateFailed, taskID, courses, err)
	} else {
		activity.setStatus(StateSucceeded, taskID, courses, nil)
	}
	return err
}

// Stop 请求停止任务，执行器会在下一个节点/进度提交前退出。
func (activity *PlatformActivity) Stop() error {
	if !activity.Running() {
		return nil
	}
	activity.setStatus(StateStopping, "", nil, nil)
	activity.stop.Store(true)
	return nil
}

// PullCourseList 拉取账号课程列表。
func (activity *PlatformActivity) PullCourseList() ([]yatori.Course, error) {
	if err := activity.Login(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	return activity.client.Courses(ctx)
}

// Close 释放底层会话。
func (activity *PlatformActivity) Close() error {
	activity.mu.Lock()
	defer activity.mu.Unlock()
	if activity.client == nil {
		return nil
	}
	err := activity.client.Close()
	activity.client = nil
	activity.UserCache = nil
	return err
}

type runnerError string

func (e runnerError) Error() string { return string(e) }

var _ Activity = (*PlatformActivity)(nil)
var _ CourseAbility = (*PlatformActivity)(nil)
var _ RunningAbility = (*PlatformActivity)(nil)
var _ StatusAbility = (*PlatformActivity)(nil)
var _ ErrorAbility = (*PlatformActivity)(nil)
var _ CourseRunner = (*PlatformActivity)(nil)
