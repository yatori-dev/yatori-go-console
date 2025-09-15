package logic

import (
	"os"
	"strings"
	"sync"
	"yatori-go-console/config"
	"yatori-go-console/logic/cqie"
	"yatori-go-console/logic/enaea"
	"yatori-go-console/logic/ketangx"
	"yatori-go-console/logic/xuexitong"
	"yatori-go-console/logic/yinghua"
	utils2 "yatori-go-console/utils"

	lg "github.com/yatori-dev/yatori-go-core/utils/log"
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

	// 检查config.yaml是否存在
	if !fileExists("./config.yaml") {
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

		user := config.Users{
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
	//初始化日志配置
	lg.LogInit(lg.StringToLOGLEVEL(configJson.Setting.BasicSetting.LogLevel), configJson.Setting.BasicSetting.LogOutFileSw == 1, configJson.Setting.BasicSetting.ColorLog, "./assets/log")
	//配置文件检查模块
	configJsonCheck(&configJson)
	//是否开启IP代理池
	checkProxyIp()

	//isIpProxy(&configJson)

	brushBlock(&configJson)
	lg.Print(lg.INFO, lg.Red, "Yatori --- ", "所有任务执行完毕")
}

var platformLock sync.WaitGroup //平台锁
// brushBlock 刷课执行块
func brushBlock(configData *config.JSONDataForConfig) {
	//统一登录模块------------------------------------------------------------------
	yingHuaAccount := yinghua.FilterAccount(configData)
	yingHuaOperation := yinghua.UserLoginOperation(yingHuaAccount)
	enaeaAccount := enaea.FilterAccount(configData)
	enaeaOperation := enaea.UserLoginOperation(enaeaAccount)
	cqieAccount := cqie.FilterAccount(configData)
	cqieOpertation := cqie.UserLoginOperation(cqieAccount)
	xueXiTongAccount := xuexitong.FilterAccount(configData)
	xueXiTongOperation := xuexitong.UserLoginOperation(xueXiTongAccount)
	ketangxAccount := ketangx.FilterAccount(configData)
	ketangxOperation := ketangx.UserLoginOperation(ketangxAccount)

	//统一刷课---------------------------------------------------------------------
	//英华
	platformLock.Add(1)
	go func() {
		yinghua.RunBrushOperation(configData.Setting, yingHuaAccount, yingHuaOperation) //英华统一刷课模块
		platformLock.Done()
	}()
	//学习公社
	platformLock.Add(1)
	go func() {
		enaea.RunBrushOperation(configData.Setting, enaeaAccount, enaeaOperation) //学习公社统一刷课模块
		platformLock.Done()
	}()
	platformLock.Add(1)
	go func() {
		cqie.RunBrushOperation(configData.Setting, cqieAccount, cqieOpertation) //重庆工程学院CQIE刷课模块
		platformLock.Done()
	}()
	//学习通
	platformLock.Add(1)
	go func() {
		xuexitong.RunBrushOperation(configData.Setting, xueXiTongAccount, xueXiTongOperation) //英华统一刷课模块
		platformLock.Done()
	}()
	//码上研训
	platformLock.Add(1)
	go func() {
		ketangx.RunBrushOperation(configData.Setting, ketangxAccount, ketangxOperation) //码上研训统一刷课模块
		platformLock.Done()
	}()
	platformLock.Wait()
}

// configJsonCheck 配置文件检测检验
func configJsonCheck(configData *config.JSONDataForConfig) {
	if len(configData.Users) == 0 {
		lg.Print(lg.INFO, lg.BoldRed, "请先在config文件中配置好相应账号")
		os.Exit(0)
	}

	//防止用户填完整url
	for i, v := range configData.Users {

		if v.AccountType == "YINGHUA" {
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
func checkProxyIp() {
	if !utils2.IsProxyFlag {
		return
	}
	lg.Print(lg.INFO, lg.Yellow, "正在开启IP池代理...")
	lg.Print(lg.INFO, lg.Yellow, "正在检查IP池IP可用性...")
	reader, err := utils2.IpFilesReader("./ip.txt")
	if err != nil {
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
		lg.Print(lg.INFO, lg.BoldRed, "无可用IP代理池，若要继续使用请先检查IP代理池文件内的IP可用性，或者在配置文件关闭IP代理功能")
		os.Exit(0)
	}
}
