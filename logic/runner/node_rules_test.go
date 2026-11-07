package runner

import (
	"testing"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

func TestIsNodeFinished(t *testing.T) {
	cases := []struct {
		name string
		node yatori.Node
		want bool
	}{
		{"平台标记完成", yatori.Node{Kind: yatori.NodeVideo, Completed: true, Progress: 62}, true},
		{"进度 100", yatori.Node{Kind: yatori.NodeVideo, Progress: 100}, true},
		{"学习通常见的 99.x", yatori.Node{Kind: yatori.NodeVideo, Progress: 99.4}, true},
		{"98.7 仍需补齐", yatori.Node{Kind: yatori.NodeVideo, Progress: 98.7}, false},
		{"作业不参与视频判定", yatori.Node{Kind: yatori.NodeWork, Progress: 100}, false},
		{"章节不参与视频判定", yatori.Node{Kind: yatori.NodeChapter, Progress: 100}, false},
	}
	for _, item := range cases {
		if got := isNodeFinished(item.node); got != item.want {
			t.Fatalf("%s: isNodeFinished=%v want %v", item.name, got, item.want)
		}
	}
}

func TestIsResumeNode(t *testing.T) {
	if !isResumeNode(yatori.Node{Kind: yatori.NodeVideo, Progress: 98.7}) {
		t.Fatal("98.7% 的视频应该走补齐流程")
	}
	if isResumeNode(yatori.Node{Kind: yatori.NodeVideo, Progress: 43.7}) {
		t.Fatal("43.7% 的视频不应该走补齐流程")
	}
	if isResumeNode(yatori.Node{Kind: yatori.NodeDocument, Progress: 98.7}) {
		t.Fatal("非视频节点不应该走补齐流程")
	}
	if isResumeNode(yatori.Node{Kind: yatori.NodeVideo, Progress: 100}) {
		t.Fatal("已完成节点不应该走补齐流程")
	}
}
