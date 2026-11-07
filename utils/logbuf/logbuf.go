package logbuf

import (
	"strings"
	"sync"
	"time"
)

const maxLines = 1000

var (
	mu    sync.RWMutex
	store = make(map[string][]string)
)

// Key 生成账号日志的存储 key。
func Key(accountType, account string) string {
	return strings.ToUpper(strings.TrimSpace(accountType)) + ":" + strings.TrimSpace(account)
}

// Append 追加一行账号日志，每个账号最多保留 maxLines 行。
func Append(key, line string) {
	if key == "" || strings.TrimSpace(line) == "" {
		return
	}
	value := time.Now().Format("15:04:05") + " " + line
	mu.Lock()
	defer mu.Unlock()
	lines := append(store[key], value)
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	store[key] = lines
}

// Tail 读取最近 limit 行。
func Tail(key string, limit int) []string {
	if limit <= 0 {
		limit = 200
	}
	if limit > maxLines {
		limit = maxLines
	}
	mu.RLock()
	defer mu.RUnlock()
	lines := store[key]
	if len(lines) == 0 {
		return []string{}
	}
	start := len(lines) - limit
	if start < 0 {
		start = 0
	}
	out := make([]string, len(lines)-start)
	copy(out, lines[start:])
	return out
}

// Clear 清空某账号日志。
func Clear(key string) {
	mu.Lock()
	defer mu.Unlock()
	delete(store, key)
}
