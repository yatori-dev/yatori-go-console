package activity

import (
	"yatori-go-console/config"
	"yatori-go-console/entity/pojo"
)

// Activity 统一接口
type Activity interface {
	Login() error      //登录统一接口
	Start() error      //启动
	Stop() error       //停止任务统一接口
	GetUserCache() any //获取Cache
	SetUser(po config.User)
	GetUser() config.User
}

// 用户活动
// 基础用户活动信息
type UserActivityBase struct {
	User      config.User
	IsRunning bool
	UserCache any
}

func (u *UserActivityBase) SetUser(po config.User) {
	u.User = po
}

func (u *UserActivityBase) GetUser() config.User {
	return u.User
}

func (u *UserActivityBase) GetUserCache() any {
	return u.UserCache
}
func BuildUserActivity(po pojo.UserPO) Activity {
	switch po.AccountType {
	case "XUEXITONG":
		return &XXTActivity{
			UserActivityBase: UserActivityBase{
				User:      config.User{AccountType: po.AccountType, Account: po.Account, Password: po.Password},
				IsRunning: false,
				UserCache: nil,
			},
		}
	case "YINGHUA":
		return &YingHuaActivity{
			UserActivityBase: UserActivityBase{
				User:      config.User{AccountType: po.AccountType, Account: po.Account, Password: po.Password},
				IsRunning: false,
				UserCache: nil,
			},
		}

	// case "ZHIHUIZHIJIAO":
	//	  return &zhihuizhijiao.ZJYActivity{...}

	default:
		return nil
	}
}
