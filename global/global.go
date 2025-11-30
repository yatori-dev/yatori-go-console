package global

import (
	"yatori-go-console/entity/pojo"
	"yatori-go-console/web/activity"

	"gorm.io/gorm"
)

var GlobalDB *gorm.DB //数据库挂载

var AccountTypeStr = map[string]string{
	"XUEXITONG": "学习通",
	"YINGHUA":   "英华学堂",
	"CANGHUI":   "仓辉实训",
	"ENAEA":     "学习公社",
	"CQIE":      "重庆工业学院",
	"KETANGX":   "码上研训",
	"ICVE":      "智慧职教",
	"QSXT":      "青书学堂",
	"WELEARN":   "WeLearn",
}

// key的值为uuid
var UserActivityMap = make(map[string]*activity.UserActivityBase) //

// 获取UserActivity
func GetUserActivity(user pojo.UserPO) *activity.UserActivityBase {
	return UserActivityMap[user.Uid]
}

// 添加UserActivity
func PutUserActivity(user pojo.UserPO, activity *activity.UserActivityBase) {
	UserActivityMap[user.Uid] = activity
}
