package runner

import (
	"context"
	"os"
	"testing"
	"time"

	"yatori-go-console/config"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

// 真实账号联调用例：只在显式设置 YATORI_IT=1 时执行，避免日常测试连平台。
// 验证“用已登录会话克隆并发客户端”这条链路是否可用。
func TestCloneClientsIntegration(t *testing.T) {
	if os.Getenv("YATORI_IT") != "1" {
		t.Skip("set YATORI_IT=1 to run integration test")
	}
	user := config.User{
		AccountType:   "XUEXITONG",
		Account:       os.Getenv("YATORI_IT_ACCOUNT"),
		Password:      os.Getenv("YATORI_IT_PASSWORD"),
		CoursesCustom: config.CoursesCustom{VideoModel: 2},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	base, err := NewClient(ctx, config.Setting{}, user)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	defer base.Close()

	pool := buildWorkerClients(ctx, base, config.Setting{}, user, 2)
	defer closeWorkerClients(pool, base)
	if len(pool) < 2 {
		t.Fatalf("expected 2 clients, got %d", len(pool))
	}

	courses, err := base.Courses(ctx)
	if err != nil {
		t.Fatalf("courses failed: %v", err)
	}
	if len(courses) == 0 {
		t.Skip("no courses")
	}

	// 有些课程本身没有章节，先找一个 base 能拉到章节的课程，再用 clone 拉一次对比
	var picked *yatori.Course
	for i := range courses {
		if _, err := base.Tree(ctx, courses[i]); err == nil {
			picked = &courses[i]
			break
		}
	}
	if picked == nil {
		t.Skip("no course with chapters")
	}

	nodes, err := pool[1].Tree(ctx, *picked)
	if err != nil {
		t.Fatalf("clone tree failed: %v", err)
	}
	t.Logf("clone OK: course=%s tree=%d", picked.Name, len(nodes))
}
