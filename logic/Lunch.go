package logic

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"yatori-go-console/config"
	"yatori-go-console/global"
	"yatori-go-console/logic/runner"
	utils2 "yatori-go-console/utils"
	"yatori-go-console/web"
	"yatori-go-console/web/activity"

	"github.com/gofrs/flock"
	lg "github.com/yatori-dev/yatori-go-core/v3/utils/log"
	"gopkg.in/yaml.v3"
)

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func Lunch() {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		lg.Print(lg.INFO, lg.BoldRed, "启动失败，无法获取系统缓存目录：", err.Error())
		os.Exit(1)
	}
	lockDir := filepath.Join(cacheDir, "yatori-go-console")
	if err = os.MkdirAll(lockDir, 0700); err != nil {
		lg.Print(lg.INFO, lg.BoldRed, "启动失败，无法创建运行目录：", err.Error())
		os.Exit(1)
	}
	instanceLock := flock.New(filepath.Join(lockDir, "instance.lock"))
	locked, err := instanceLock.TryLock()
	if err != nil {
		lg.Print(lg.INFO, lg.BoldRed, "启动失败，无法检查程序运行状态：", err.Error())
		os.Exit(1)
	}
	if !locked {
		lg.Print(lg.INFO, lg.BoldRed, "Yatori-go-console 已在运行，请勿重复启动")
		os.Exit(1)
	}
	defer instanceLock.Unlock()

	// --web / -web：直接进入 Web 管理模式，账号从数据库读取
	webModeFlag := hasFlag("--web") || hasFlag("-web")
	if !fileExists("./config.yaml") && webModeFlag {
		createDefaultWebConfig()
	}

	// 检查config.yaml是否存在
	if !fileExists("./config.yaml") {
		lg.Print(lg.INFO, `
程序未检测到config.yaml配置文件，

如果你用的配置文件生成器你确定你文件放对位置了？
以及请不要放个config（1）.yaml，config（2）.yaml这样子的的文件，要的是config.yaml。

如果使用的是控制台配置生成器的同学可以忽略此条信息。
`)
		// 不存在使用生成方式建立
		setConfig := config.JSONDataForConfig{}
		// 设置基本设置
		setConfig.Setting.BasicSetting.CompletionTone = 1
		setConfig.Setting.BasicSetting.ColorLog = 1
		setConfig.Setting.BasicSetting.LogOutFileSw = 1
		setConfig.Setting.BasicSetting.LogLevel = "INFO"
		setConfig.Setting.BasicSetting.LogModel = 0
		//setConfig.Setting.BasicSetting.IpProxySw = 0

		setConfig.Setting.AiSetting.AiType = "TONGYI"
		setConfig.Setting.ApiQueSetting.Url = "http://localhost:8083"

		accountType := config.GetUserInput("请输入平台类型 (如 YINGHUA)(全大写): ")
		url := config.GetUserInput("请输入平台的URL链接 (可留空): ")
		account := config.GetUserInput("请输入账号: ")
		password := config.GetUserInput("请输入密码: ")

		videoModel := config.GetUserInput("请输入刷视频模式 (0-不刷, 1-普通模式, 2-暴力模式, 3-去红模式): ")
		autoExam := config.GetUserInput("是否自动考试? (0-不考试, 1-AI考试, 2-外部题库对接考试): ")
		examAutoSubmit := config.GetUserInput("考完试是否自动提交试卷? (0-否, 1-是): ")
		includeCourses := config.GetUserInput("请输入需要包含的课程名称，多个用(英文逗号)分隔(可留空): ")
		excludeCourses := config.GetUserInput("请输入需要排除的课程名称，多个用(英文逗号)分隔(可留空): ")

		cleanStringSlice := func(s string) []string {
			if s == "" {
				return []string{}
			}
			parts := strings.Split(s, ",")
			var result []string
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}

		user := config.User{
			AccountType: accountType,
			URL:         url,
			Account:     account,
			Password:    password,
			CoursesCustom: config.CoursesCustom{
				VideoModel:     config.StrToInt(videoModel),
				AutoExam:       config.StrToInt(autoExam),
				ExamAutoSubmit: config.StrToInt(examAutoSubmit),
				IncludeCourses: cleanStringSlice(includeCourses),
				ExcludeCourses: cleanStringSlice(excludeCourses),
			},
		}
		setConfig.Users = append(setConfig.Users, user)

		data, err := yaml.Marshal(&setConfig)
		if err != nil {
			panic(err)
		}

		err = os.WriteFile("./config.yaml", data, 0644)
		if err != nil {
			panic(err)
		}
	}
	//读取配置文件
	configJson := config.ReadConfig("./config.yaml")
	if webModeFlag {
		configJson.Setting.BasicSetting.WebModel = 1
	}
	//全局设置（Web 模式下的 DB 账号也会复用 AI、邮箱等配置）
	global.Setting = configJson.Setting
	//初始化日志配置
	lg.LogInit(lg.StringToLOGLEVEL(configJson.Setting.BasicSetting.LogLevel), configJson.Setting.BasicSetting.LogOutFileSw == 1, configJson.Setting.BasicSetting.ColorLog, "./assets/log")
	//配置文件检查模块
	configJsonCheck(&configJson)
	//是否开启IP代理池（Web 模式允许 config.yaml 不配置账号）
	checkProxyIp(configJson.Setting.BasicSetting.WebModel == 1)

	//isIpProxy(&configJson)
	//如果开启了Web模式，则直接切换Web模式
	if configJson.Setting.BasicSetting.WebModel == 1 {
		activity.SetDefaultSetting(configJson.Setting)
		web.ServiceInit()
	}
	brushBlock(&configJson)
	lg.Print(lg.INFO, lg.Red, "Yatori --- ", "所有任务执行完毕")
}

func hasFlag(flag string) bool {
	for _, arg := range os.Args[1:] {
		if arg == flag {
			return true
		}
	}
	return false
}

// createDefaultWebConfig 在 --web 首次启动且没有 config.yaml 时生成一份不含账号的基础配置。
func createDefaultWebConfig() {
	setConfig := config.JSONDataForConfig{}
	setConfig.Setting.BasicSetting.CompletionTone = 1
	setConfig.Setting.BasicSetting.ColorLog = 1
	setConfig.Setting.BasicSetting.LogOutFileSw = 1
	setConfig.Setting.BasicSetting.LogLevel = "INFO"
	setConfig.Setting.BasicSetting.LogModel = 0
	setConfig.Setting.BasicSetting.WebModel = 1
	setConfig.Setting.AiSetting.AiType = "TONGYI"
	setConfig.Setting.ApiQueSetting.Url = "http://localhost:8083"

	data, err := yaml.Marshal(&setConfig)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("./config.yaml", data, 0644); err != nil {
		panic(err)
	}
	lg.Print(lg.INFO, lg.Green, "已生成 Web 模式基础配置 config.yaml（不含账号，账号请在网页中添加）")
}

var platformLock sync.WaitGroup //用户锁

// brushBlock 刷课执行块：v3 统一 SDK 下不再按平台拆分，账号之间并发执行。
func brushBlock(configData *config.JSONDataForConfig) {
	for i := range configData.Users {
		platformLock.Add(1)
		go func(user config.User) {
			defer platformLock.Done()
			runner.RunUser(configData.Setting, user)
		}(configData.Users[i])
	}
	platformLock.Wait()
}

// configJsonCheck 配置文件检测检验
func configJsonCheck(configData *config.JSONDataForConfig) {
	webMode := configData.Setting.BasicSetting.WebModel == 1
	// Web 模式的账号从数据库读取，config.yaml 可以不再配置 users
	if len(configData.Users) == 0 {
		if webMode {
			lg.Print(lg.INFO, lg.Yellow, "Web 模式：config.yaml 未配置账号，将从数据库读取账号")
			return
		}
		lg.Print(lg.INFO, lg.BoldRed, "请先在config文件中配置好相应账号")
		os.Exit(0)
	}
	if len(configData.Users) > 5 && !webMode {
		os.Exit(0)
	}
	//防止用户填完整url
	for i, v := range configData.Users {

		if v.AccountType == "YINGHUA" || v.AccountType == "HQKJ" || v.AccountType == "ZXCPKS" {
			if !strings.HasPrefix(v.URL, "http") {
				lg.Print(lg.INFO, lg.BoldRed, "账号", v.Account, "未配置正确url，请先在config文件中配置好相应账号信息")
				os.Exit(0)
			}
			split := strings.Split(v.URL, "/")
			(*configData).Users[i].URL = split[0] + "/" + split[1] + "/" + split[2]
		}

		//如果有账号开启代理，那么标记Flag就未true
		if v.IsProxy == 1 {
			utils2.IsProxyFlag = true
		}
	}
}

// 检查代理IP是否为正常
func checkProxyIp(webMode bool) {
	if !utils2.IsProxyFlag && !webMode {
		return
	}
	if webMode {
		// Web 模式账号在数据库里，允许没有 ip.txt（不使用代理的账号可以正常跑）
		if _, err := os.Stat("./ip.txt"); os.IsNotExist(err) {
			return
		}
	}
	lg.Print(lg.INFO, lg.Yellow, "正在开启IP池代理...")
	lg.Print(lg.INFO, lg.Yellow, "正在检查IP池IP可用性...")
	reader, err := utils2.IpFilesReader("./ip.txt")
	if err != nil {
		if webMode {
			lg.Print(lg.INFO, lg.BoldRed, "IP代理池文件ip.txt读取失败，Web 模式下将忽略代理池")
			return
		}
		lg.Print(lg.INFO, lg.BoldRed, "IP代理池文件ip.txt读取失败，请确认文件格式或者内容是否正确")
		os.Exit(0)
	}
	for _, v := range reader {
		_, state, err := utils2.CheckProxyIp(v)
		if err != nil {
			lg.Print(lg.INFO, " ["+v+"] ", lg.BoldRed, "该IP代理不可用，错误信息：", err.Error())
			continue
		}
		lg.Print(lg.INFO, " ["+v+"] ", lg.Green, "检测通过，状态：", state)
		utils2.IPProxyPool = append(utils2.IPProxyPool, v) //添加到IP代理池里面
	}
	lg.Print(lg.INFO, lg.BoldGreen, "IP检查完毕")
	//若无可用IP代理则直接退出
	if len(utils2.IPProxyPool) == 0 {
		if webMode {
			lg.Print(lg.INFO, lg.BoldRed, "无可用IP代理池，Web 模式下未开启代理的账号仍可正常运行")
			return
		}
		lg.Print(lg.INFO, lg.BoldRed, "无可用IP代理池，若要继续使用请先检查IP代理池文件内的IP可用性，或者在配置文件关闭IP代理功能")
		os.Exit(0)
	}
}
