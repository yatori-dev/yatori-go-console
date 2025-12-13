package activity

import (
	"encoding/json"
	"yatori-go-console/config"
	"yatori-go-console/entity/pojo"
)

// Activity 统一接口
type Activity interface {
	Login() error      //登录统一接口
	Start() error      //启动
	Stop() error       //停止任务统一接口
	GetUserCache() any //获取Cache
	SetUser(user config.User)
	GetUser() config.User
}

// 用户活动
// 基础用户活动信息
type UserActivityBase struct {
	User      config.User //配置文件
	IsRunning bool
	UserCache any
}

// 设置User
func (u *UserActivityBase) SetUser(user config.User) {
	u.User = user
}

// 获取User
func (u *UserActivityBase) GetUser() config.User {
	return u.User
}

// 获取UserCache
func (u *UserActivityBase) GetUserCache() any {
	return u.UserCache
}

// 构建活动
func BuildUserActivity(po pojo.UserPO) Activity {
	switch po.AccountType {
	case "XUEXITONG":
		user := config.User{}
		err2 := json.Unmarshal([]byte(po.UserConfigJson), &user)
		if err2 != nil {
			return nil
		}
		return &XXTActivity{
			UserActivityBase: UserActivityBase{
				User:      user,
				IsRunning: false,
				UserCache: nil,
			},
		}
	case "YINGHUA":
		user := config.User{}
		err2 := json.Unmarshal([]byte(po.UserConfigJson), &user)
		if err2 != nil {
			return nil
		}
		return &YingHuaActivity{
			UserActivityBase: UserActivityBase{
				User:      user,
				IsRunning: false,
				UserCache: nil,
			},
		}

	default:
		return nil
	}
}
