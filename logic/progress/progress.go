// Package progress 维护学习节点的本地完成记录。
//
// 学习通等平台的课程完成率、章节卡片是否通过都存在缓存：刚刷完的节点，
// 平台接口可能还要过一段时间才更新。只依赖平台数据会导致
//   - 已经 100% 的节点下次又被当成未完成重新刷；
//   - 课程进度一直停在旧值（例如明明学完了还显示 11%）。
//
// 这里把“平台返回已完成”与“我们提交成功过”的节点合并落库，作为本地视图。
package progress

import (
	"encoding/json"
	"strings"
	"sync"
	"time"

	"yatori-go-console/dao"
	"yatori-go-console/entity/pojo"

	"gorm.io/gorm"
)

var (
	mu    sync.Mutex
	db    *gorm.DB
	store = make(map[string]*pojo.StudyProgressPO)
)

// Init 注入数据库句柄；未初始化时退化为纯内存模式（单元测试可用）。
func Init(gdb *gorm.DB) {
	mu.Lock()
	defer mu.Unlock()
	db = gdb
}

// AccountKey 生成进度归属的账号 key，与账号类型一起区分同号不同平台。
func AccountKey(accountType, account string) string {
	return strings.ToUpper(strings.TrimSpace(accountType)) + ":" + strings.TrimSpace(account)
}

func key(account, courseID string) string {
	return account + "|" + courseID
}

// Snapshot 返回某门课当前的本地完成集合。
func Snapshot(account, courseID string) map[string]bool {
	mu.Lock()
	defer mu.Unlock()
	po := loadLocked(account, courseID)
	out := make(map[string]bool)
	if po == nil {
		return out
	}
	for _, id := range decode(po.DoneIDs) {
		out[id] = true
	}
	return out
}

// MergeDone 把平台报告的已完成节点并入本地记录，并写入最新的节点总数。
// 返回合并后的完成集合，调用方可以直接用它来判断节点是否要跳过。
func MergeDone(account, courseID string, total int, platformDone []string) map[string]bool {
	mu.Lock()
	defer mu.Unlock()

	po := loadLocked(account, courseID)
	if po == nil {
		po = &pojo.StudyProgressPO{Account: account, CourseID: courseID}
	}
	done := make(map[string]bool)
	for _, id := range decode(po.DoneIDs) {
		done[id] = true
	}
	for _, id := range platformDone {
		if id != "" {
			done[id] = true
		}
	}
	if total > 0 {
		po.Total = total
	}
	po.DoneIDs = encode(done)
	po.UpdatedAt = time.Now()
	saveLocked(po)
	return done
}

// MarkDone 记录一个节点已经完成。
func MarkDone(account, courseID, nodeID string) {
	if nodeID == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	po := loadLocked(account, courseID)
	if po == nil {
		po = &pojo.StudyProgressPO{Account: account, CourseID: courseID}
	}
	done := make(map[string]bool)
	for _, id := range decode(po.DoneIDs) {
		done[id] = true
	}
	if done[nodeID] {
		return
	}
	done[nodeID] = true
	po.DoneIDs = encode(done)
	po.UpdatedAt = time.Now()
	saveLocked(po)
}

// Stats 返回本地记录的完成节点数与节点总数；total 为 0 表示还没有本地数据。
func Stats(account, courseID string) (done int, total int, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	po := loadLocked(account, courseID)
	if po == nil || po.Total <= 0 {
		return 0, 0, false
	}
	return len(decode(po.DoneIDs)), po.Total, true
}

// Clear 清空某个账号的本地进度（删除账号时调用）。
func Clear(account string) {
	mu.Lock()
	defer mu.Unlock()
	for k := range store {
		if strings.HasPrefix(k, account+"|") {
			delete(store, k)
		}
	}
	if db != nil {
		_ = dao.DeleteStudyProgress(db, account)
	}
}

func loadLocked(account, courseID string) *pojo.StudyProgressPO {
	if po, exists := store[key(account, courseID)]; exists {
		return po
	}
	if db == nil {
		return nil
	}
	po, err := dao.QueryStudyProgress(db, account, courseID)
	if err != nil || po == nil {
		return nil
	}
	store[key(account, courseID)] = po
	return po
}

func saveLocked(po *pojo.StudyProgressPO) {
	store[key(po.Account, po.CourseID)] = po
	if db == nil {
		return
	}
	_ = dao.SaveStudyProgress(db, po)
}

func encode(done map[string]bool) string {
	if len(done) == 0 {
		return "[]"
	}
	ids := make([]string, 0, len(done))
	for id := range done {
		ids = append(ids, id)
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decode(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var ids []string
	if err := json.Unmarshal([]byte(raw), &ids); err != nil {
		return nil
	}
	return ids
}
