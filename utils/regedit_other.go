//go:build !windows
// +build !windows

package utils

// 在非Windows系统（比如Linux）下，给出一个空的实现，避免出错
func SetVirtualTerminalLevel() {
	// 什么都不做
}
