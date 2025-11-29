package main

import (
	"fmt"
	"yatori-go-console/config"
	utils2 "yatori-go-console/init"
	"yatori-go-console/logic"
	"yatori-go-console/utils"
)

func main() {
	utils2.YatoriConsoleInit()       //初始化yatori-console
	fmt.Println(config.YaotirLogo()) //打印LOGO
	utils.ShowAnnouncement()         //用于显示公告
	logic.Lunch()                    //启动yatori-console
}
