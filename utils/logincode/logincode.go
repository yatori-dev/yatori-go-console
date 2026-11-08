// Package logincode 暂存登录过程中的验证码挑战（短信 / 图形），供 Web 端输入。
//
// 学习通 App 登录链路在风控触发时会要求短信验证码，core 通过
// Options.LoginCodeProvider 回调等待上层提供。这个包把挑战挂起，
// 由前端轮询拿到提示并在弹窗里输入，实现"无终端交互"的验证码流程。
package logincode

import (
	"context"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

// Challenge 是一次待处理的验证码挑战。
type Challenge struct {
	Kind    string
	Prompt  string
	Phone   string
	Image   []byte
	SendSMS func(ctx context.Context) error
}

// Snapshot 是暴露给前端的只读状态。
type Snapshot struct {
	Kind       string `json:"kind"`
	Prompt     string `json:"prompt"`
	Phone      string `json:"phone"`
	Image      string `json:"image,omitempty"`
	CanSendSMS bool   `json:"canSendSms"`
}

type entry struct {
	challenge Challenge
	code      chan string
	createdAt time.Time
}

var (
	mu    sync.Mutex
	store = make(map[string]*entry)
)

// Key 生成账号维度的 key。
func Key(accountType, account string) string {
	return accountType + ":" + account
}

// Wait 挂起一个验证码挑战并等待前端提交，超时或 ctx 结束则返回错误。
func Wait(ctx context.Context, key string, challenge Challenge, timeout time.Duration) (string, error) {
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	item := &entry{challenge: challenge, code: make(chan string, 1), createdAt: time.Now()}

	mu.Lock()
	store[key] = item
	mu.Unlock()

	defer func() {
		mu.Lock()
		if store[key] == item {
			delete(store, key)
		}
		mu.Unlock()
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case code := <-item.code:
		return code, nil
	case <-ctx.Done():
		return "", ctx.Err()
	case <-timer.C:
		return "", errors.New("等待验证码超时，请重试")
	}
}

// Submit 提交前端输入的验证码。
func Submit(key, code string) error {
	mu.Lock()
	item := store[key]
	mu.Unlock()
	if item == nil {
		return errors.New("当前没有待输入的验证码")
	}
	select {
	case item.code <- code:
		return nil
	default:
		return errors.New("验证码已提交，请等待登录结果")
	}
}

// SendSMS 触发平台发送短信验证码。
func SendSMS(ctx context.Context, key string) error {
	mu.Lock()
	item := store[key]
	mu.Unlock()
	if item == nil || item.challenge.SendSMS == nil {
		return errors.New("当前没有可发送的短信验证码")
	}
	return item.challenge.SendSMS(ctx)
}

// Pending 返回当前待处理的挑战；没有则返回 nil。
func Pending(key string) *Snapshot {
	mu.Lock()
	defer mu.Unlock()
	item := store[key]
	if item == nil {
		return nil
	}
	out := &Snapshot{
		Kind:       item.challenge.Kind,
		Prompt:     item.challenge.Prompt,
		Phone:      item.challenge.Phone,
		CanSendSMS: item.challenge.SendSMS != nil,
	}
	if len(item.challenge.Image) > 0 {
		out.Image = base64.StdEncoding.EncodeToString(item.challenge.Image)
	}
	return out
}

// Clear 清理某个账号的挂起状态（停止任务时调用）。
func Clear(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(store, key)
}
