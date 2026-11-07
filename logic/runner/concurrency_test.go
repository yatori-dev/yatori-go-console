package runner

import (
	"context"
	"testing"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/logic/progress"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

func TestConcurrentModeHelpers(t *testing.T) {
	if !concurrentCourses(yatori.XueXiTong, 2) {
		t.Fatal("学习通 2 应该是多课程模式")
	}
	if !concurrentNodes(yatori.XueXiTong, 3) {
		t.Fatal("学习通 3 应该是多任务点模式")
	}
	if concurrentCourses(yatori.XueXiTong, 3) || concurrentCourses(yatori.YingHua, 2) {
		t.Fatal("多课程模式只对学习通 2 生效")
	}
	if concurrentNodes(yatori.XueXiTong, 2) || concurrentNodes(yatori.YingHua, 2) {
		t.Fatal("多任务点并发只对学习通 3 生效")
	}
	if !singleSubmitPlatform(yatori.ICVE) || !singleSubmitPlatform(yatori.Ketangx) {
		t.Fatal("智慧职教 / 码上研训 应走单次提交")
	}
	if singleSubmitPlatform(yatori.XueXiTong) {
		t.Fatal("学习通不属于单次提交平台")
	}
}

func TestFastStudyMapping(t *testing.T) {
	cases := []struct {
		platform yatori.Platform
		mode     int
		want     bool
	}{
		{yatori.YingHua, 1, false},
		{yatori.YingHua, 2, true},
		{yatori.YingHua, 3, false},
		{yatori.HaiQiKeJi, 2, true},
		{yatori.ENAAE, 2, true},
		{yatori.CQIE, 2, true},
		{yatori.WeLearn, 2, true},
		{yatori.ICVE, 1, true},
		{yatori.Ketangx, 1, true},
		{yatori.XueXiTong, 1, false},
		{yatori.XueXiTong, 2, false},
		{yatori.XueXiTong, 3, false},
		{yatori.QSXT, 1, false},
	}
	for _, item := range cases {
		if got := fastStudy(item.platform, item.mode); got != item.want {
			t.Fatalf("fastStudy(%v,%d)=%v want %v", item.platform, item.mode, got, item.want)
		}
	}
}

func TestSerialStudyMode(t *testing.T) {
	if !isSerialStudyMode(yatori.XueXiTong, 1) {
		t.Fatal("学习通普通模式需要节点间隔")
	}
	if isSerialStudyMode(yatori.XueXiTong, 2) || isSerialStudyMode(yatori.XueXiTong, 3) {
		t.Fatal("学习通多课程 / 多任务点不应再有节点间隔")
	}
	if isSerialStudyMode(yatori.YingHua, 1) {
		t.Fatal("英华普通模式不需要 10-60 秒间隔")
	}
}

func TestWorkerCountBounds(t *testing.T) {
	if got := courseWorkerCount(1); got != 1 {
		t.Fatalf("1 门课并发数应为 1，实际 %d", got)
	}
	if got := courseWorkerCount(10); got != maxCourseWorkers {
		t.Fatalf("并发数应被限制在 %d，实际 %d", maxCourseWorkers, got)
	}

	v := 3
	user := config.User{CoursesCustom: config.CoursesCustom{CxNode: &v}}
	if got := nodeWorkerCount(user); got != 3 {
		t.Fatalf("cxNode=3 时并发数应为 3，实际 %d", got)
	}
	big := 99
	user.CoursesCustom.CxNode = &big
	if got := nodeWorkerCount(user); got != maxNodeWorkers {
		t.Fatalf("cxNode 超限时应限制在 %d，实际 %d", maxNodeWorkers, got)
	}
	user.CoursesCustom.CxNode = nil
	if got := nodeWorkerCount(user); got != 3 {
		t.Fatalf("cxNode 为空时应默认 3，实际 %d", got)
	}
}

// 没有 Cookie 会话时并发池应自动退化为单客户端，任务仍然照常执行。
func TestRunCoursesConcurrentFallsBackToSerial(t *testing.T) {
	user := config.User{
		AccountType:   "XUEXITONG",
		Account:       "concurrency-fallback",
		CoursesCustom: config.CoursesCustom{VideoModel: 2},
	}
	defer progress.Clear(progress.AccountKey(user.AccountType, user.Account))

	client := &fakeClient{
		platform: yatori.XueXiTong,
		courses:  []yatori.Course{{ID: "c1", Name: "课程1"}, {ID: "c2", Name: "课程2"}},
		nodes: []yatori.Node{
			{ID: "n1", CourseID: "c1", Name: "视频1", Kind: yatori.NodeVideo, Duration: 30 * time.Second},
		},
	}

	RunClient(context.Background(), client, config.Setting{}, user, nil)
	if client.studyCalls == 0 {
		t.Fatal("降级为串行后仍应执行学习任务")
	}
}
