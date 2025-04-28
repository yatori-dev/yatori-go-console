//go:build windows
// +build windows

package utils

import (
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/sys/windows/registry"
)

func SetVirtualTerminalLevel() {
	//修改
	// 打开（或创建）指定的注册表键
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Console`, registry.SET_VALUE)
	if err != nil {
		lg.Print(lg.INFO, lg.Red, "打开注册表 Console 失败", err.Error())
	}
	defer key.Close()

	// 设置 VirtualTerminalLevel 的值为 1 (DWORD类型)
	err = key.SetDWordValue("VirtualTerminalLevel", 1)
	if err != nil {
		lg.Print(lg.INFO, lg.Red, "设置注册表 Console/VirtualTerminalLevel = 1 失败", err.Error())
	}
}
