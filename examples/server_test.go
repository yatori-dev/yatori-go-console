package examples

import (
	"fmt"
	"testing"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/web/activity"
)

func TestTTXActivity(t *testing.T) {
	po := pojo.UserPO{AccountType: "XUEXITONG", Account: "15891657669", Password: "fjm11222324.", UserConfigJson: `{"accountType":"XUEXITONG","url":"","account":"15891657669","password":"fjm11222324.","isProxy":0,"informEmails":null,"coursesCustom":{"studyTime":"","cxNode":null,"cxChapterTestSw":null,"cxWorkSw":null,"cxExamSw":null,"shuffleSw":0,"videoModel":0,"autoExam":0,"examAutoSubmit":0,"excludeCourses":null,"includeCourses":null,"coursesSettings":null}}`}

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
