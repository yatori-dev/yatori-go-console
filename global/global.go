package global

import (
	"yatori-go-console/entity/pojo"
	"yatori-go-console/logic"

	"gorm.io/gorm"
)

var GlobalDB *gorm.DB //数据库挂载

// key的值为uuid
var UserActivityMap = make(map[string]*logic.UserActivity) //

// 获取UserActivity
func GetUserActivity(user pojo.UserPO) *logic.UserActivity {
	return UserActivityMap[user.Uid]
}

// 添加UserActivity
func PutUserActivity(user pojo.UserPO, activity *logic.UserActivity) {
	UserActivityMap[user.Uid] = activity
}
