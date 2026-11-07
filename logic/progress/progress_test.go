package progress

import "testing"

func TestMergeDoneAndStats(t *testing.T) {
	account := AccountKey(" XUEXITONG ", "probe-user")
	if account != "XUEXITONG:probe-user" {
		t.Fatalf("unexpected account key: %s", account)
	}

	done := MergeDone(account, "course-a", 4, []string{"n1", "n2"})
	if len(done) != 2 {
		t.Fatalf("want 2 done nodes, got %d", len(done))
	}
	if got, total, ok := Stats(account, "course-a"); !ok || got != 2 || total != 4 {
		t.Fatalf("stats=%d/%d ok=%v", got, total, ok)
	}

	MarkDone(account, "course-a", "n3")
	if got, _, _ := Stats(account, "course-a"); got != 3 {
		t.Fatalf("want 3 done nodes, got %d", got)
	}

	// 平台再次回报已完成节点时不应重复计数
	MergeDone(account, "course-a", 4, []string{"n1", "n2"})
	if got, total, _ := Stats(account, "course-a"); got != 3 || total != 4 {
		t.Fatalf("after merge stats=%d/%d", got, total)
	}

	// 课程新增节点时总数被更新
	MergeDone(account, "course-a", 6, nil)
	if _, total, _ := Stats(account, "course-a"); total != 6 {
		t.Fatalf("want total 6, got %d", total)
	}

	Clear(account)
	if _, _, ok := Stats(account, "course-a"); ok {
		t.Fatal("progress should be cleared")
	}
}

func TestStatsWithoutData(t *testing.T) {
	if _, _, ok := Stats("XUEXITONG:missing", "course"); ok {
		t.Fatal("no local progress expected")
	}
}
