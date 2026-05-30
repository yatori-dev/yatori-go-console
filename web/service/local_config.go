package service

import (
	"bufio"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/dto"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/global"

	"github.com/google/uuid"
	"golang.org/x/crypto/scrypt"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigFilePath = "./config.yaml"
	defaultLogDirPath     = "./logs"
)

var localConfigMu sync.Mutex

type localConfigRoot struct {
	Setting config.Setting    `yaml:"setting"`
	Users   []localConfigUser `yaml:"users"`
}

type localConfigUser struct {
	Uid           string               `json:"uid" yaml:"uid"`
	AccountType   string               `json:"accountType" yaml:"accountType"`
	URL           string               `json:"url" yaml:"url"`
	RemarkName    string               `json:"remarkName,omitempty" yaml:"remarkName,omitempty"`
	Account       string               `json:"account" yaml:"account"`
	Password      string               `json:"password" yaml:"password"`
	IsProxy       int                  `json:"isProxy" yaml:"isProxy"`
	InformEmails  []string             `json:"informEmails" yaml:"informEmails"`
	CoursesCustom config.CoursesCustom `json:"coursesCustom" yaml:"coursesCustom"`
	Deletable     bool                 `json:"deletable,omitempty" yaml:"deletable,omitempty"`
}

func getConfigFilePath() string {
	if value := strings.TrimSpace(os.Getenv("CONFIG_FILE_PATH")); value != "" {
		return value
	}
	return defaultConfigFilePath
}

func getLogDirPath() string {
	if value := strings.TrimSpace(os.Getenv("YATORI_LOG_DIR")); value != "" {
		return value
	}
	return defaultLogDirPath
}

func loadLocalConfig() (*localConfigRoot, error) {
	content, err := os.ReadFile(getConfigFilePath())
	if err != nil {
		return nil, err
	}

	cfg := &localConfigRoot{}
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}
	if cfg.Users == nil {
		cfg.Users = []localConfigUser{}
	}
	return cfg, nil
}

func saveLocalConfig(cfg *localConfigRoot) error {
	content, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigFilePath(), content, 0644)
}

func ensureLocalUserUIDs(cfg *localConfigRoot) bool {
	changed := false
	seen := make(map[string]struct{}, len(cfg.Users))
	for i := range cfg.Users {
		uid := strings.TrimSpace(cfg.Users[i].Uid)
		if uid == "" {
			cfg.Users[i].Uid = uuid.NewString()
			changed = true
			continue
		}
		if _, exists := seen[uid]; exists {
			cfg.Users[i].Uid = uuid.NewString()
			changed = true
			continue
		}
		seen[cfg.Users[i].Uid] = struct{}{}
	}
	return changed
}

func localConfigUserToDTO(user localConfigUser) dto.ConfigManagerUser {
	return dto.ConfigManagerUser{
		Uid:           user.Uid,
		AccountType:   user.AccountType,
		URL:           user.URL,
		RemarkName:    user.RemarkName,
		Account:       user.Account,
		Password:      user.Password,
		IsProxy:       user.IsProxy,
		InformEmails:  user.InformEmails,
		CoursesCustom: user.CoursesCustom,
		Deletable:     user.Deletable,
	}
}

func dtoToLocalConfigUser(user dto.ConfigManagerUser) localConfigUser {
	return localConfigUser{
		Uid:           user.Uid,
		AccountType:   user.AccountType,
		URL:           user.URL,
		RemarkName:    user.RemarkName,
		Account:       user.Account,
		Password:      user.Password,
		IsProxy:       user.IsProxy,
		InformEmails:  user.InformEmails,
		CoursesCustom: user.CoursesCustom,
		Deletable:     user.Deletable,
	}
}

func fetchConfigManagerUsers() ([]dto.ConfigManagerUser, error) {
	localConfigMu.Lock()
	defer localConfigMu.Unlock()

	cfg, err := loadLocalConfig()
	if err != nil {
		return nil, err
	}

	changed := ensureLocalUserUIDs(cfg)
	users := make([]dto.ConfigManagerUser, 0, len(cfg.Users))
	for _, user := range cfg.Users {
		users = append(users, localConfigUserToDTO(user))
	}

	if changed {
		if err := saveLocalConfig(cfg); err != nil {
			return nil, err
		}
	}

	return users, nil
}

func getLocalConfigUserByUID(uid string) (*dto.ConfigManagerUser, error) {
	users, err := fetchConfigManagerUsers()
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.Uid == uid {
			result := user
			return &result, nil
		}
	}
	return nil, errors.New("用户不存在")
}

func upsertLocalConfigUser(user dto.ConfigManagerUser) (*dto.ConfigManagerUser, error) {
	localConfigMu.Lock()
	defer localConfigMu.Unlock()

	cfg, err := loadLocalConfig()
	if err != nil {
		return nil, err
	}
	ensureLocalUserUIDs(cfg)

	if strings.TrimSpace(user.Uid) == "" {
		user.Uid = uuid.NewString()
	}
	if user.InformEmails == nil {
		user.InformEmails = []string{}
	}
	if user.CoursesCustom.IncludeCourses == nil {
		user.CoursesCustom.IncludeCourses = []string{}
	}
	if user.CoursesCustom.ExcludeCourses == nil {
		user.CoursesCustom.ExcludeCourses = []string{}
	}
	if user.CoursesCustom.CoursesSettings == nil {
		user.CoursesCustom.CoursesSettings = []config.CoursesSettings{}
	}

	replaced := false
	for i := range cfg.Users {
		if cfg.Users[i].Uid == user.Uid {
			cfg.Users[i] = dtoToLocalConfigUser(user)
			replaced = true
			break
		}
	}
	if !replaced {
		cfg.Users = append(cfg.Users, dtoToLocalConfigUser(user))
	}

	if err := saveLocalConfig(cfg); err != nil {
		return nil, err
	}
	return &user, nil
}

func deleteLocalConfigUser(uid string, adminPass string) error {
	localConfigMu.Lock()
	defer localConfigMu.Unlock()

	cfg, err := loadLocalConfig()
	if err != nil {
		return err
	}
	ensureLocalUserUIDs(cfg)

	for i := range cfg.Users {
		if cfg.Users[i].Uid != uid {
			continue
		}
		if !cfg.Users[i].Deletable && !verifyStoredPassword(adminPass, cfg.Setting.BasicSetting.AdminPassword) {
			return errors.New("该账号未标记为可删除，需要管理员权限")
		}
		cfg.Users = append(cfg.Users[:i], cfg.Users[i+1:]...)
		return saveLocalConfig(cfg)
	}
	return errors.New("用户不存在")
}

func updateLocalConfigUser(uid string, updateFn func(user *dto.ConfigManagerUser) error) (*dto.ConfigManagerUser, error) {
	localConfigMu.Lock()
	defer localConfigMu.Unlock()

	cfg, err := loadLocalConfig()
	if err != nil {
		return nil, err
	}
	ensureLocalUserUIDs(cfg)

	for i := range cfg.Users {
		if cfg.Users[i].Uid != uid {
			continue
		}

		user := localConfigUserToDTO(cfg.Users[i])
		if err := updateFn(&user); err != nil {
			return nil, err
		}
		user.Uid = uid
		cfg.Users[i] = dtoToLocalConfigUser(user)
		if err := saveLocalConfig(cfg); err != nil {
			return nil, err
		}
		return &user, nil
	}
	return nil, errors.New("用户不存在")
}

func verifyStoredPassword(input string, stored string) bool {
	if stored == "" {
		return true
	}
	if input == "" {
		return false
	}
	if !strings.HasPrefix(stored, "scrypt:") {
		return subtle.ConstantTimeCompare([]byte(input), []byte(stored)) == 1
	}

	parts := strings.SplitN(stored, "$", 3)
	if len(parts) != 3 {
		return false
	}
	params := strings.Split(strings.TrimPrefix(parts[0], "scrypt:"), ":")
	if len(params) != 3 {
		return false
	}

	n, err1 := strconv.Atoi(params[0])
	r, err2 := strconv.Atoi(params[1])
	p, err3 := strconv.Atoi(params[2])
	expected, err4 := hex.DecodeString(parts[2])
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return false
	}

	derived, err := scrypt.Key([]byte(input), []byte(parts[1]), n, r, p, len(expected))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(derived, expected) == 1
}

func verifyEditPassword(user dto.ConfigManagerUser, inputPassword string) bool {
	if verifyStoredPassword(inputPassword, user.Password) {
		return true
	}

	cfg, err := loadLocalConfig()
	if err != nil {
		return false
	}
	return verifyStoredPassword(inputPassword, cfg.Setting.BasicSetting.AdminPassword)
}

func findActiveLogFile() string {
	candidates := []string{
		filepath.Join(getLogDirPath(), "console_server.log"),
		filepath.Join(getLogDirPath(), "yatori_core.log"),
	}

	// 自动扫描 assets/log 目录下的带时间戳日志
	assetsLogDir := "./assets/log"
	if files, err := filepath.Glob(filepath.Join(assetsLogDir, "log*.txt")); err == nil {
		candidates = append(candidates, files...)
	}

	var selected string
	var selectedModTime int64
	for _, path := range candidates {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		modTime := info.ModTime().UnixNano()
		if selected == "" || modTime > selectedModTime {
			selected = path
			selectedModTime = modTime
		}
	}
	return selected
}

func getLocalConfigUserLogs(uid string, limit int) (map[string]any, error) {
	user, err := getLocalConfigUserByUID(uid)
	if err != nil {
		return nil, err
	}

	activeLogFile := findActiveLogFile()
	if activeLogFile == "" {
		return map[string]any{
			"success": true,
			"uid":     uid,
			"logs":    "",
		}, nil
	}

	file, err := os.Open(activeLogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if limit < 1 {
		limit = 1
	}
	if limit > 2000 {
		limit = 2000
	}

	maxScanLines := limit * 20
	if maxScanLines < 1000 {
		maxScanLines = 1000
	}

	account := strings.TrimSpace(user.Account)
	maskedAccount := maskAccountString(account)
	buffer := make([]string, 0, maxScanLines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if len(buffer) >= maxScanLines {
			copy(buffer, buffer[1:])
			buffer[len(buffer)-1] = line
		} else {
			buffer = append(buffer, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	ansiPattern := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	phonePattern := regexp.MustCompile(`(1[3-9]\d)\d{4}(\d{4})`)
	ignoredKeywords := []string{"Request Body:", "SELECT * FROM `user_pos`", "rows:", "[GIN]"}

	matched := make([]string, 0, limit)
	for _, line := range buffer {
		if account != "" && strings.Contains(line, account) || maskedAccount != "" && strings.Contains(line, maskedAccount) {
			cleaned := ansiPattern.ReplaceAllString(line, "")
			ignored := false
			for _, keyword := range ignoredKeywords {
				if strings.Contains(cleaned, keyword) {
					ignored = true
					break
				}
			}
			if ignored {
				continue
			}
			matched = append(matched, cleaned)
		}
	}
	if len(matched) > limit {
		matched = matched[len(matched)-limit:]
	}

	content := phonePattern.ReplaceAllString(strings.Join(matched, ""), "${1}****${2}")
	return map[string]any{
		"success": true,
		"uid":     uid,
		"logs":    content,
	}, nil
}

func maskAccountString(account string) string {
	if matched, _ := regexp.MatchString(`^1[3-9]\d{9}$`, account); matched {
		return account[:3] + "****" + account[len(account)-4:]
	}
	parts := strings.Split(account, "@")
	if len(parts) == 2 {
		prefix := parts[0]
		if len(prefix) > 3 {
			return prefix[:3] + "***@" + parts[1]
		}
		return "***@" + parts[1]
	}
	return account
}

func syncUsersFromConfigManager() ([]pojo.UserPO, error) {
	users, err := fetchConfigManagerUsers()
	if err != nil {
		return nil, err
	}

	uidList := make([]string, 0, len(users))
	result := make([]pojo.UserPO, 0, len(users))
	for _, user := range users {
		userPO, err := configManagerUserToPO(user)
		if err != nil {
			return nil, err
		}
		if err := dao.UpsertUser(global.GlobalDB, userPO); err != nil {
			return nil, err
		}
		uidList = append(uidList, user.Uid)
		result = append(result, *userPO)
	}

	if err := dao.DeleteUsersNotInUIDs(global.GlobalDB, uidList); err != nil {
		return nil, err
	}
	return result, nil
}

func configManagerUserToPO(user dto.ConfigManagerUser) (*pojo.UserPO, error) {
	userConfig := config.User{
		AccountType:   user.AccountType,
		URL:           user.URL,
		RemarkName:    user.RemarkName,
		Account:       user.Account,
		Password:      user.Password,
		IsProxy:       user.IsProxy,
		InformEmails:  user.InformEmails,
		CoursesCustom: user.CoursesCustom,
	}

	userConfigJSON, err := json.Marshal(userConfig)
	if err != nil {
		return nil, err
	}

	return &pojo.UserPO{
		Uid:            user.Uid,
		AccountType:    user.AccountType,
		Url:            user.URL,
		Account:        user.Account,
		Password:       user.Password,
		UserConfigJson: string(userConfigJSON),
	}, nil
}
