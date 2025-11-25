package logic

import (
	"yatori-go-console/entity/pojo"
)

func NewExecutor() *UserActivity {
	return &UserActivity{
		stop: make(chan struct{}),
	}
}

func (e *UserActivity) Stop() {
	close(e.stop)
}

// 用户活动
type UserActivity struct {
	pojo.UserPO      //用户数据
	IsLogin     bool //是否为登录状态
	IsRunning   bool //是否在执行中，执行中的任务无法进行数据更改，需要停止运行才能
	Cache       any  //登录后的缓存数据
	stop        chan struct{}
}
