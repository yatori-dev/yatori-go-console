package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"yatori-go-console/entity/pojo"
	"yatori-go-console/global"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&pojo.UserPO{}, &pojo.SettingPO{}); err != nil {
		t.Fatal(err)
	}
	global.GlobalDB = db
}

func performJSON(t *testing.T, method, target string, handler gin.HandlerFunc, body any) (int, map[string]any) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Handle(method, target, handler)

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(method, target, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	result := map[string]any{}
	_ = json.Unmarshal(recorder.Body.Bytes(), &result)
	return recorder.Code, result
}

func TestAddAndUpdateAccountHTTP(t *testing.T) {
	setupTestDB(t)

	body := map[string]any{
		"accountType": "hqkj",
		"url":         "https://demo.haiqikeji.com/student/home",
		"account":     " tester ",
		"password":    "secret",
		"remarkName":  "海旗测试号",
		"coursesCustom": map[string]any{
			"videoModel":     2,
			"autoExam":       2,
			"examAutoSubmit": 1,
		},
	}
	status, response := performJSON(t, http.MethodPost, "/add", AddUserService, body)
	if status != http.StatusOK || int(response["code"].(float64)) != 200 {
		t.Fatalf("add failed: status=%d response=%v", status, response)
	}

	user, err := queryFirstUser()
	if err != nil {
		t.Fatal(err)
	}
	if user.AccountType != "HQKJ" || user.Url != "https://demo.haiqikeji.com" || user.Account != "tester" {
		t.Fatalf("unexpected user: %+v", user)
	}
	if !strings.Contains(user.UserConfigJson, `"autoExam":2`) {
		t.Fatalf("config json missing autoExam: %s", user.UserConfigJson)
	}

	update := map[string]any{
		"uid":         user.Uid,
		"accountType": "HQKJ",
		"url":         user.Url,
		"account":     user.Account,
		"remarkName":  "更新后的备注",
		"coursesCustom": map[string]any{
			"videoModel":     1,
			"autoExam":       0,
			"examAutoSubmit": 0,
		},
	}
	status, response = performJSON(t, http.MethodPost, "/update", UpdateUserService, update)
	if status != http.StatusOK || int(response["code"].(float64)) != 200 {
		t.Fatalf("update failed: status=%d response=%v", status, response)
	}
	updated, err := queryFirstUser()
	if err != nil {
		t.Fatal(err)
	}
	if updated.Password != "secret" {
		t.Fatalf("password should be preserved, got %q", updated.Password)
	}
	if !strings.Contains(updated.UserConfigJson, `"remarkName":"更新后的备注"`) {
		t.Fatalf("remark not updated: %s", updated.UserConfigJson)
	}
}

func TestSettingHTTP(t *testing.T) {
	setupTestDB(t)

	body := map[string]any{
		"basicSetting": map[string]any{"completionTone": 1, "colorLog": 1, "logOutFileSw": 1, "logLevel": "INFO", "logModel": 0},
		"emailInform":  map[string]any{"sw": 0, "SMTPHost": "", "SMTPPort": 465, "userName": "", "password": ""},
		"aiSetting":    map[string]any{"aiType": "DEEPSEEK", "aiUrl": "", "model": "deepseek-chat", "API_KEY": "sk-test"},
		"apiQueSetting": map[string]any{
			"url": "http://localhost:8083",
		},
	}
	status, response := performJSON(t, http.MethodPost, "/setting", SaveSettingService, body)
	if status != http.StatusOK || int(response["code"].(float64)) != 200 {
		t.Fatalf("save setting failed: status=%d response=%v", status, response)
	}
	if global.Setting.AiSetting.Model != "deepseek-chat" {
		t.Fatalf("global setting not updated: %+v", global.Setting)
	}
}

func queryFirstUser() (pojo.UserPO, error) {
	var user pojo.UserPO
	err := global.GlobalDB.First(&user).Error
	return user, err
}
