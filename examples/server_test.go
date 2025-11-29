package examples

import (
	"fmt"
	"testing"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/web/activity"
)

func TestTTXActivity(t *testing.T) {
	po := pojo.UserPO{AccountType: "XUEXITONG", Account: "15891657669", Password: "fjm11222324."}

	userActivity := activity.BuildUserActivity(po)

	err := userActivity.Login()
	if err != nil {
		fmt.Println(err)
	}
	err = userActivity.Start()
	if err != nil {
		fmt.Println(err)
	}
	if xxt, ok := userActivity.(activity.XXTAbility); ok {
		list, err := xxt.PullCourseList()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(list)
	}
	select {}
}
