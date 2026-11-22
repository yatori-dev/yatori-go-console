// Package loginstate 记录账号的登录方式可用性，避免重复走注定失败的登录链路。
//
// 学习通的 App 登录和网页登录是两套独立会话：部分账号 App 登录成功后
// 拿到的会话访问 Web 课程接口仍会跳登录页，只能再走一次网页登录。
// 这里把“该账号 App 会话不可用”记到本地，后续直接走网页登录，
// 避免每次运行都先弹一次 App 验证码、再弹一次网页验证码。
package loginstate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	stateFile = "assets/login_state.json"
	// 记录 24 小时后失效，core 修好 App 会话后会自动重新尝试。
	stateTTL = 24 * time.Hour
)

type record struct {
	AppUnsupported bool      `json:"appUnsupported"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

var (
	mu     sync.Mutex
	store  = map[string]record{}
	loaded bool
)

func load() {
	if loaded {
		return
	}
	loaded = true
	data, err := os.ReadFile(stateFile)
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &store)
}

func save() {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(stateFile), 0o755)
	_ = os.WriteFile(stateFile, data, 0o644)
}

// AppUnsupported 返回该账号是否已被判定为“App 会话无法访问课程接口”。
func AppUnsupported(key string) bool {
	mu.Lock()
	defer mu.Unlock()
	load()
	item, ok := store[key]
	if !ok {
		return false
	}
	if time.Since(item.UpdatedAt) > stateTTL {
		delete(store, key)
		save()
		return false
	}
	return item.AppUnsupported
}

// MarkAppUnsupported 记录该账号的 App 会话不可用于 Web 课程接口。
func MarkAppUnsupported(key string) {
	mu.Lock()
	defer mu.Unlock()
	load()
	store[key] = record{AppUnsupported: true, UpdatedAt: time.Now()}
	save()
}

// Clear 清除账号记录（删除账号时调用）。
func Clear(key string) {
	mu.Lock()
	defer mu.Unlock()
	load()
	if _, ok := store[key]; !ok {
		return
	}
	delete(store, key)
	save()
}
