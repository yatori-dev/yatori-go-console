package global

import (
	"sync"

	"yatori-go-console/config"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/web/activity"

	"gorm.io/gorm"
)

var GlobalDB *gorm.DB // 全局数据库

// Setting 当前进程使用的全局设置（AI、邮箱、日志等），由 logic.Lunch 加载。
var Setting config.Setting

var AccountTypeStr = map[string]string{
	"XUEXITONG": "超星学习通",
	"YINGHUA":   "英华学堂",
	"CANGHUI":   "仓辉实训",
	"ENAEA":     "学习公社",
	"CQIE":      "重庆工程学院",
	"KETANGX":   "码上研训",
	"ICVE":      "智慧职教",
	"QSXT":      "青书学堂",
	"WELEARN":   "WeLearn",
	"HQKJ":      "海旗科技",
	"TTCDW":     "团课",
	"WEIBAN":    "安全微伴",
	"CELA":      "CELA",
	"MOOC":      "中国大学MOOC",
	"ZHIHUISHU": "智慧树",
	"GONGXUE":   "工学云",
}

var (
	activityMu      sync.RWMutex
	UserActivityMap = make(map[string]*activity.Activity) // key 为账号 uuid
)

// GetUserActivity 获取账号运行态
func GetUserActivity(user pojo.UserPO) *activity.Activity {
	activityMu.RLock()
	defer activityMu.RUnlock()
	return UserActivityMap[user.Uid]
}

// PutUserActivity 保存账号运行态
func PutUserActivity(user pojo.UserPO, value *activity.Activity) {
	activityMu.Lock()
	defer activityMu.Unlock()
	UserActivityMap[user.Uid] = value
}

// DeleteUserActivity 删除账号运行态（账号更新/删除时调用）
func DeleteUserActivity(uid string) {
	activityMu.Lock()
	defer activityMu.Unlock()
	delete(UserActivityMap, uid)
}
