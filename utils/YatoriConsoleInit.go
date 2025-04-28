package utils

import (
	_ "embed"
	"github.com/yatori-dev/yatori-go-core/utils"
	"log"
	"os"
	"runtime"
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
		SetVirtualTerminalLevel()
	}
}
