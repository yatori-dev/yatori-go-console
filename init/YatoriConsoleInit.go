package init

import (
	_ "embed"
	"log"
	"os"
	"runtime"
	utils2 "yatori-go-console/utils"
	//"yatori-go-console/web"

	"github.com/yatori-dev/yatori-go-core/utils"
)

//go:embed assets/finishNotice.mp3
var noticeSound []byte

// 初始化YatoriConsole
func YatoriConsoleInit() {
	utils.YatoriCoreInit() //初始化Core核心
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
		utils2.SetVirtualTerminalLevel()
	}
}
