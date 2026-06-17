package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/yatori-dev/yatori-go-core/models/ctype"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
)

type JSONDataForConfig struct {
	Setting Setting `json:"setting"`
	Users   []User  `json:"users"`
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
	WebModel       int    `json:"webModel" yaml:"webModel"`
	// 以下为 Web 模式的可选安全/网络配置，均可留空保持原有行为（向后兼容）
	WebHost       string   `json:"webHost,omitempty" yaml:"webHost,omitempty"`             //Web 服务监听地址，默认 0.0.0.0
	WebPort       int      `json:"webPort,omitempty" yaml:"webPort,omitempty"`             //Web 服务监听端口，默认 8080
	AdminPassword string   `json:"adminPassword,omitempty" yaml:"adminPassword,omitempty"` //若非空，则 /api/v1 接口需通过 X-Admin-Pass 头或 admin_pass 查询参数鉴权；为空则不鉴权（默认）
	AllowOrigins  []string `json:"allowOrigins,omitempty" yaml:"allowOrigins,omitempty"`   //CORS 允许的来源白名单；为空时默认放行 *（不带 Credentials）
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
	StudyTime       string            `json:"studyTime" yaml:"studyTime"`             //WeLearn设置刷学时的时候范围
	CxNode          *int              `json:"cxNode" yaml:"cxNode"`                   //学习通多任务点模式下设置同时任务点数量
	CxChapterTestSw *int              `json:"cxChapterTestSw" yaml:"cxChapterTestSw"` //学习通是否开启章测
	CxWorkSw        *int              `json:"cxWorkSw" yaml:"cxWorkSw"`               //学习通是否开启作业
	CxExamSw        *int              `json:"cxExamSw" yaml:"cxExamSw"`               //学习通是否开启考试
	ShuffleSw       int               `json:"shuffleSw" yaml:"shuffleSw"`             //是否打乱顺序学习，1为打乱顺序，0为不打乱
	VideoModel      int               `json:"videoModel" yaml:"videoModel"`           //观看视频模式
	AutoExam        int               `json:"autoExam" yaml:"autoExam"`               //是否自动考试
	ExamAutoSubmit  int               `json:"examAutoSubmit" yaml:"examAutoSubmit"`   //是否自动提交试卷
	ExcludeCourses  []string          `json:"excludeCourses" yaml:"excludeCourses"`
	IncludeCourses  []string          `json:"includeCourses" yaml:"includeCourses"`
	CoursesSettings []CoursesSettings `json:"coursesSettings" yaml:"coursesSettings"`
}
type User struct {
	AccountType   string        `json:"accountType" yaml:"accountType"`
	URL           string        `json:"url"`
	RemarkName    string        `json:"remarkName,omitempty" yaml:"remarkName,omitempty" mapstructure:"remarkName"` //可选备注名
	Account       string        `json:"account"`
	Password      string        `json:"password"`
	IsProxy       int           `json:"isProxy" yaml:"isProxy"` //是否代理IP
	InformEmails  []string      `json:"informEmails" yaml:"informEmails"`
	CoursesCustom CoursesCustom `json:"coursesCustom" yaml:"coursesCustom"`
}

var (
	remarkNameLock sync.RWMutex
	remarkNames    = map[string]string{}
)

func DisplayAccount(account string) string {
	remarkNameLock.RLock()
	defer remarkNameLock.RUnlock()
	if remarkName, ok := remarkNames[account]; ok {
		return remarkName
	}
	return account
}

func registerRemarkNames(config *JSONDataForConfig) {
	remarkNameLock.Lock()
	defer remarkNameLock.Unlock()
	remarkNames = map[string]string{}
	ambiguousAccounts := map[string]struct{}{}
	for _, user := range config.Users {
		remarkName := strings.TrimSpace(user.RemarkName)
		if user.Account != "" && remarkName != "" {
			if existingRemarkName, ok := remarkNames[user.Account]; ok && existingRemarkName != remarkName {
				delete(remarkNames, user.Account)
				ambiguousAccounts[user.Account] = struct{}{}
				continue
			}
			if _, ok := ambiguousAccounts[user.Account]; ok {
				continue
			}
			remarkNames[user.Account] = remarkName
		}
	}
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
	registerRemarkNames(&configJson)
	return configJson
}

// 默认值处理
func defaultValue(config *JSONDataForConfig) {
	for i := range config.Users {
		v := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		if config.Users[i].CoursesCustom.CxNode == nil {
			(&config.Users[i].CoursesCustom).CxNode = &v[3]
		}
		if config.Users[i].CoursesCustom.CxChapterTestSw == nil {
			(&config.Users[i].CoursesCustom).CxChapterTestSw = &v[1]
		}
		if config.Users[i].CoursesCustom.CxChapterTestSw == nil {
			(&config.Users[i].CoursesCustom).CxChapterTestSw = &v[1]
		}
		if config.Users[i].CoursesCustom.CxWorkSw == nil {
			(&config.Users[i].CoursesCustom).CxWorkSw = &v[1]
		}
		if config.Users[i].CoursesCustom.CxExamSw == nil {
			(&config.Users[i].CoursesCustom).CxExamSw = &v[1]
		}
	}
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
	defaultValue(&configJson) //设置默认值
	registerRemarkNames(&configJson)

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

var configWriteLock sync.Mutex

// SaveRawConfigAtomic 以原子方式写入配置文件，避免并发写或写入中途崩溃导致 config.yaml 被截断/损坏。
// 先写入同目录下的临时文件，fsync 后再 rename 覆盖目标文件；全程持有进程级互斥锁。
func SaveRawConfigAtomic(path string, data []byte) error {
	configWriteLock.Lock()
	defer configWriteLock.Unlock()

	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".config-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // rename 成功后此文件已不存在，Remove 仅在出错路径生效

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
