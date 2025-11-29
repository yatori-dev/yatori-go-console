package global

import (
	"yatori-go-console/entity/pojo"
	"yatori-go-console/web/activity"

	"gorm.io/gorm"
)

var GlobalDB *gorm.DB //数据库挂载

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
