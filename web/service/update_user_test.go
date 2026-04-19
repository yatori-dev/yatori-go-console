package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/dto"
	"yatori-go-console/global"

	"github.com/gin-gonic/gin"
)

func writeTempConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`setting:
  basicSetting:
    webModel: 1
    adminPassword: ""
users:
- uid: test-uid
  accountType: XUEXITONG
  url: ""
  account: "test-account"
  password: test-password
  isProxy: 0
  informEmails: []
  coursesCustom:
    studyTime: 10-30
    cxNode: 3
    cxChapterTestSw: 0
    cxWorkSw: 0
    cxExamSw: 1
    shuffleSw: 0
    videoModel: 0
    autoExam: 3
    examAutoSubmit: 0
    excludeCourses: []
    includeCourses: []
    coursesSettings: []
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func readTempConfig(t *testing.T, path string) config.JSONDataForConfig {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp config: %v", err)
	}
	var cfg config.JSONDataForConfig
	if err := json.Unmarshal(content, &cfg); err == nil {
		return cfg
	}
	// fallback to viper-based loader is not suitable for temp paths, so just verify through local loader
	localCfg, err := loadLocalConfig()
	if err != nil {
		t.Fatalf("load local config: %v", err)
	}
	return config.JSONDataForConfig{
		Setting: localCfg.Setting,
		Users: []config.User{
			{
				AccountType:   localCfg.Users[0].AccountType,
				URL:           localCfg.Users[0].URL,
				RemarkName:    localCfg.Users[0].RemarkName,
				Account:       localCfg.Users[0].Account,
				Password:      localCfg.Users[0].Password,
				IsProxy:       localCfg.Users[0].IsProxy,
				InformEmails:  localCfg.Users[0].InformEmails,
				CoursesCustom: localCfg.Users[0].CoursesCustom,
			},
		},
	}
}

func TestUpdateLocalConfigUserPersistsCoursesCustom(t *testing.T) {
	path := writeTempConfig(t)
	t.Setenv("CONFIG_FILE_PATH", path)

	value := 1
	if _, err := updateLocalConfigUser("test-uid", func(user *dto.ConfigManagerUser) error {
		user.CoursesCustom.CxWorkSw = &value
		return nil
	}); err != nil {
		t.Fatalf("updateLocalConfigUser failed: %v", err)
	}

	localCfg, err := loadLocalConfig()
	if err != nil {
		t.Fatalf("loadLocalConfig failed: %v", err)
	}
	if localCfg.Users[0].CoursesCustom.CxWorkSw == nil || *localCfg.Users[0].CoursesCustom.CxWorkSw != 1 {
		t.Fatalf("expected cxWorkSw=1, got %+v", localCfg.Users[0].CoursesCustom.CxWorkSw)
	}
}

func TestUpdateUserServicePersistsCoursesCustom(t *testing.T) {
	path := writeTempConfig(t)
	t.Setenv("CONFIG_FILE_PATH", path)
	t.Setenv("YATORI_LOG_DIR", t.TempDir())

	db, err := dao.SqliteInit()
	if err != nil {
		t.Fatalf("SqliteInit failed: %v", err)
	}
	global.GlobalDB = db

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	body := map[string]any{
		"uid":          "test-uid",
		"accountType":  "XUEXITONG",
		"url":          "",
		"account":      "15960600263",
		"password":     "Yorushika451.",
		"remarkName":   "",
		"informEmails": []string{},
		"coursesCustom": map[string]any{
			"studyTime":       "10-30",
			"videoModel":      0,
			"autoExam":        3,
			"examAutoSubmit":  0,
			"cxNode":          3,
			"cxChapterTestSw": 0,
			"cxWorkSw":        1,
			"cxExamSw":        1,
			"shuffleSw":       0,
			"includeCourses":  []string{},
			"excludeCourses":  []string{},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/updateAccount", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	ctx.Request = req

	UpdateUserService(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}

	localCfg, err := loadLocalConfig()
	if err != nil {
		t.Fatalf("loadLocalConfig failed: %v", err)
	}
	if localCfg.Users[0].CoursesCustom.CxWorkSw == nil || *localCfg.Users[0].CoursesCustom.CxWorkSw != 1 {
		t.Fatalf("expected cxWorkSw=1 after UpdateUserService, got %+v", localCfg.Users[0].CoursesCustom.CxWorkSw)
	}
}
