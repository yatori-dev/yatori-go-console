package runner

import (
	"testing"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

func TestBuildPendingOrderedKeepsTreeOrder(t *testing.T) {
	nodes := []yatori.Node{
		{ID: "v1", Kind: yatori.NodeVideo},
		{ID: "c1", Kind: yatori.NodeWork},
		{ID: "v2", Kind: yatori.NodeVideo},
		{ID: "c2", Kind: yatori.NodeWork},
	}

	got := nodeIDs(buildPendingOrdered(nodes, map[string]bool{}, true))
	want := []string{"v1", "c1", "v2", "c2"}
	if !sameIDs(got, want) {
		t.Fatalf("章节树顺序被改变: got=%v want=%v", got, want)
	}

	got = nodeIDs(buildPendingOrdered(nodes, map[string]bool{}, false))
	want = []string{"v1", "v2"}
	if !sameIDs(got, want) {
		t.Fatalf("未开启章测时应只保留视频: got=%v want=%v", got, want)
	}

	got = nodeIDs(buildPendingOrdered(nodes, map[string]bool{"c1": true}, true))
	want = []string{"v1", "v2", "c2"}
	if !sameIDs(got, want) {
		t.Fatalf("本地已完成章测应被跳过: got=%v want=%v", got, want)
	}

	completed := append([]yatori.Node(nil), nodes...)
	completed[0].Completed = true
	got = nodeIDs(buildPendingOrdered(completed, map[string]bool{}, true))
	want = []string{"c1", "v2", "c2"}
	if !sameIDs(got, want) {
		t.Fatalf("平台已完成视频应被跳过: got=%v want=%v", got, want)
	}
}

func nodeIDs(nodes []yatori.Node) []string {
	out := make([]string, len(nodes))
	for i, node := range nodes {
		out[i] = node.ID
	}
	return out
}

func sameIDs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
