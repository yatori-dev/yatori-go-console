package utils

import (
	_ "embed"
	lg "github.com/yatori-dev/yatori-go-core/utils/log"
	"golang.org/x/sys/windows/registry"
	"log"
	"os"
	"runtime"

	"github.com/yatori-dev/yatori-go-core/utils"
)

//go:embed finishNotice.mp3
var noticeSound []byte

// 初始化YatoriConsole
func YatoriConsoleInit() {
	utils.YatoriCoreInit() //初始化Core核心

	f1, _ := utils.PathExists("./assets/sound/finishNotice.mp3")
	if !f1 {
		writeDLLToDisk() //确保文件都加载了
	}
	initConsole()
}

// 将必要文件复制到当前目录下
func writeDLLToDisk() {
	utils.PathExistForCreate("./assets/sound")
	noticePath := "./assets/sound/finishNotice.mp3"
	f1 := os.WriteFile(noticePath, noticeSound, 0644)
	if f1 != nil {
		log.Fatal(f1)
	}
}

// 初始化控制台配置
func initConsole() {
	sysType := runtime.GOOS
	if sysType == "windows" {
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
}
