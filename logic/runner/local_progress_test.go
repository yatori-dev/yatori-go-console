package runner

import (
	"context"
	"testing"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/logic/progress"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

// 平台还没更新、但本地记录已经完成的节点应直接跳过，不再重复刷。
func TestRunSkipsLocallyCompletedNodes(t *testing.T) {
	user := config.User{
		AccountType:   "XUEXITONG",
		Account:       "local-progress-skip",
		CoursesCustom: config.CoursesCustom{VideoModel: 2},
	}
	accountKey := progress.AccountKey(user.AccountType, user.Account)
	defer progress.Clear(accountKey)
	progress.MergeDone(accountKey, "c1", 1, []string{"n1"})

	client := &fakeClient{
		platform: yatori.XueXiTong,
		courses:  []yatori.Course{{ID: "c1", Name: "课程1"}},
		nodes: []yatori.Node{
			{ID: "n1", CourseID: "c1", Name: "视频1", Kind: yatori.NodeVideo, Duration: 30 * time.Second},
		},
	}

	RunClient(context.Background(), client, config.Setting{}, user, nil)
	if client.studyCalls != 0 {
		t.Fatalf("已完成节点不应重复提交，实际调用 %d 次", client.studyCalls)
	}
}

// 刷完的节点要写入本地进度，供进度显示与下次跳过使用。
func TestRunMarksNodeDone(t *testing.T) {
	user := config.User{
		AccountType:   "XUEXITONG",
		Account:       "local-progress-mark",
		CoursesCustom: config.CoursesCustom{VideoModel: 2},
	}
	accountKey := progress.AccountKey(user.AccountType, user.Account)
	defer progress.Clear(accountKey)

	client := &fakeClient{
		platform: yatori.XueXiTong,
		courses:  []yatori.Course{{ID: "c2", Name: "课程2"}},
		nodes: []yatori.Node{
			{ID: "n2", CourseID: "c2", Name: "视频2", Kind: yatori.NodeVideo, Duration: 30 * time.Second},
		},
	}

	RunClient(context.Background(), client, config.Setting{}, user, nil)
	if client.studyCalls == 0 {
		t.Fatal("未完成节点应该被提交")
	}
	done, total, ok := progress.Stats(accountKey, "c2")
	if !ok || done != 1 || total != 1 {
		t.Fatalf("本地进度应为 1/1，实际 %d/%d ok=%v", done, total, ok)
	}
}
