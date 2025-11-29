package activity

import (
	"log"

	"github.com/yatori-dev/yatori-go-core/aggregation/yinghua"
	yinghuaApi "github.com/yatori-dev/yatori-go-core/api/yinghua"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
)

type YingHuaActivity struct {
	UserActivityBase
}

// YingHuaAbility 英华相关接口
type YingHuaAbility interface {
	PullCourseList() ([]yinghua.YingHuaCourse, error) //拉取课程
}

func (activity *YingHuaActivity) Start() error {
	//TODO implement me
	activity.IsRunning = true
	return nil
}

func (activity *YingHuaActivity) Stop() error {
	//TODO implement me
	activity.IsRunning = false
	return nil
}

// Login 登录
func (activity *YingHuaActivity) Login() error {

	cache := &yinghuaApi.YingHuaUserCache{PreUrl: activity.User.URL, Account: activity.User.Account, Password: activity.User.Password}

	err1 := yinghua.YingHuaLoginAction(cache) // 登录
	if err1 != nil {
		lg.Print(lg.INFO, "[", lg.Green, cache.Account, lg.White, "] ", lg.Red, err1.Error())
		log.Fatal(err1) //登录失败则直接退出
	}
	activity.UserCache = cache
	return nil
}

// PullCourseList 拉取课程列表
func (activity *YingHuaActivity) PullCourseList() ([]yinghua.YingHuaCourse, error) {
	cache := activity.UserCache.(*yinghuaApi.YingHuaUserCache)
	courseList, err := yinghua.CourseListAction(cache) //拉取课程列表
	if err != nil {
		return nil, err
	}
	return courseList, nil
}
