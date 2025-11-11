package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

type JSONDataForConfig struct {
	Setting Setting `json:"setting"`
	Users   []Users `json:"users"`
}
type EmailInform struct {
	Sw       int    `json:"sw"`
	SMTPHost string `json:"smtpHost" yaml:"SMTPHost"`
	SMTPPort int    `json:"smtpPort" yaml:"SMTPPort"`
	UserName string `json:"userName" yaml:"userName"`
	Password string `json:"password"`
}
type BasicSetting struct {
	CompletionTone int    `default:"1" json:"completionTone,omitempty" yaml:"completionTone"` //是否开启刷完提示音，0为关闭，1为开启，默认为1
	ColorLog       int    `json:"colorLog,omitempty" yaml:"colorLog"`                         //是否为彩色日志，0为关闭彩色日志，1为开启，默认为1
	LogOutFileSw   int    `json:"logOutFileSw,omitempty" yaml:"logOutFileSw"`                 //是否输出日志文件0代表不输出，1代表输出，默认为1
	LogLevel       string `json:"logLevel,omitempty" yaml:"logLevel"`                         //日志等级，默认INFO，DEBUG为找BUG调式用的，日志内容较详细，默认为INFO
	LogModel       int    `json:"logModel" yaml:"logModel"`                                   //日志模式，0代表以视频提交学时基准打印日志，1代表以一个课程为基准打印信息，默认为0
}
type AiSetting struct {
	AiType ctype.AiType `json:"aiType" yaml:"aiType"`
	AiUrl  string       `json:"aiUrl" yaml:"aiUrl"`
	Model  string       `json:"model"`
	APIKEY string       `json:"API_KEY" yaml:"API_KEY" mapstructure:"API_KEY"`
}

type ApiQueSetting struct {
	Url string `json:"url"`
}

type Setting struct {
	BasicSetting  BasicSetting  `json:"basicSetting" yaml:"basicSetting"`
	EmailInform   EmailInform   `json:"emailInform" yaml:"emailInform"`
	AiSetting     AiSetting     `json:"aiSetting" yaml:"aiSetting"`
	ApiQueSetting ApiQueSetting `json:"apiQueSetting" yaml:"apiQueSetting"`
}
type CoursesSettings struct {
	Name         string   `json:"name"`
	IncludeExams []string `json:"includeExams" yaml:"includeExams"`
	ExcludeExams []string `json:"excludeExams" yaml:"excludeExams"`
}
type CoursesCustom struct {
	WeLearnTime     string            `json:"weLearnTime" yaml:"weLearnTime"`       //WeLearn设置刷学时的时候范围
	CxNode          int               `json:"cxNode" yaml:"cxNode"`                 //学习通多任务点模式下设置同时任务点数量
	ShuffleSw       int               `json:"shuffleSw" yaml:"shuffleSw"`           //是否打乱顺序学习，1为打乱顺序，0为不打乱
	VideoModel      int               `json:"videoModel" yaml:"videoModel"`         //观看视频模式
	AutoExam        int               `json:"autoExam" yaml:"autoExam"`             //是否自动考试
	ExamAutoSubmit  int               `json:"examAutoSubmit" yaml:"examAutoSubmit"` //是否自动提交试卷
	ExcludeCourses  []string          `json:"excludeCourses" yaml:"excludeCourses"`
	IncludeCourses  []string          `json:"includeCourses" yaml:"includeCourses"`
	CoursesSettings []CoursesSettings `json:"coursesSettings" yaml:"coursesSettings"`
}
type Users struct {
	AccountType   string        `json:"accountType" yaml:"accountType"`
	URL           string        `json:"url"`
	Account       string        `json:"account"`
	Password      string        `json:"password"`
	IsProxy       int           `json:"isProxy" yaml:"isProxy"` //是否代理IP
	InformEmails  []string      `json:"informEmails" yaml:"informEmails"`
	CoursesCustom CoursesCustom `json:"coursesCustom" yaml:"coursesCustom"`
}

// 读取json配置文件
func ReadJsonConfig(filePath string) JSONDataForConfig {
	var configJson JSONDataForConfig
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &configJson)
	if err != nil {
		log.Fatal(err)
	}
	return configJson
}

// 自动识别读取配置文件
func ReadConfig(filePath string) JSONDataForConfig {
	var configJson JSONDataForConfig
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		log2.Print(log2.INFO, log2.BoldRed, "找不到配置文件或配置文件内容书写错误")
		log.Fatal(err)
	}
	err = viper.Unmarshal(&configJson)
	//viper.SetTypeByDefaultValue(true)
	viper.SetDefault("setting.basicSetting.logModel", 5)

	if err != nil {
		log2.Print(log2.INFO, log2.BoldRed, "配置文件读取失败，请检查配置文件填写是否正确")
		log.Fatal(err)
	}
	return configJson
}

// CmpCourse 比较是否存在对应课程,匹配上了则true，没有匹配上则是false
func CmpCourse(course string, courseList []string) bool {
	for i := range courseList {
		if courseList[i] == course {
			return true
		}
	}
	return false
}

func GetUserInput(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(text)
}

func StrToInt(s string) int {
	res, err := strconv.Atoi(s)
	if err != nil {
		return 0 // 其他错误处理逻辑
	}
	return res
}
