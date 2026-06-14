package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"yatori-go-console/config"
	"yatori-go-console/dao"
	"yatori-go-console/entity/pojo"
	"yatori-go-console/entity/vo"
	"yatori-go-console/global"
	"yatori-go-console/utils"
	"yatori-go-console/web/activity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
)

// 拉取账号列表
func UserListService(c *gin.Context) {
	users, total, err := dao.QueryUsers(global.GlobalDB, 1, 10)
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	//转换列表--------------
	resUserList := []map[string]any{}
	for _, user := range users {
		toMap := utils.StructToMap(user)
		userActivity := global.GetUserActivity(user)
		if userActivity != nil {
			if xxt, ok := (*userActivity).(*activity.XXTActivity); ok {
				toMap["isRunning"] = xxt.IsRunning
			}
		} else {
			toMap["isRunning"] = false
		}

		resUserList = append(resUserList, toMap)
	}
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "拉取账号成功",
		Data: gin.H{
			"users": resUserList,
			"total": total,
		},
	})
}

// 添加账号
func AddUserService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req vo.AddAccountRequest
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}
	//检测账号是否已存在
	userPo := pojo.UserPO{
		AccountType: req.AccountType,
		Url:         req.Url,
		Account:     req.Account,
	}
	user, _ := dao.QueryUser(global.GlobalDB, userPo)
	if user != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号已存在",
		})
		return
	}

	uuidV7, _ := uuid.NewV7()
	userPo.Uid = uuidV7.String()   //设置uuid值
	userPo.Password = req.Password //设置密码
	userConfig := config.User{
		AccountType: userPo.AccountType,
		URL:         userPo.Url,
		Account:     userPo.Account,
		Password:    userPo.Password,
	}

	userConfigJson, err2 := json.Marshal(userConfig)
	if err2 != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err2.Error(),
		})
		return
	}
	userPo.UserConfigJson = string(userConfigJson) //赋值Config配置

	err := dao.InsertUser(global.GlobalDB, &userPo)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  err.Error(),
		})
		return
	}
	//登录成功
	c.JSON(200,
		vo.Response{
			Code:    200,
			Message: "添加账号成功",
			Data:    &userPo,
		})
}

// 删除账号
func DeleteUserService(c *gin.Context) {
	var req vo.DeleteAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "数据转换失败",
		})
		return
	}
	//如果uid不为空则采用uid方式删除
	if req.Uid != "" {
		err := dao.DeleteUser(global.GlobalDB, &pojo.UserPO{Uid: req.Uid})
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: "删除失败",
			})
			return
		}
	} else if req.AccountType != "" && req.Account != "" { //如果uid方式没有，则直接使用账号和账号类型方式联合查询删除
		err := dao.DeleteUser(global.GlobalDB, &pojo.UserPO{
			AccountType: req.AccountType,
			Url:         req.Url,
			Account:     req.Account,
		})
		if err != nil {
			c.JSON(http.StatusOK, vo.Response{
				Code:    400,
				Message: "删除失败",
			})
			return
		}
	}

	c.JSON(200,
		vo.Response{
			Code:    200,
			Message: "删除成功",
		})
}

// 检查账号密码是否正确
func AccountLoginCheckService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req vo.AccountLoginCheckRequest
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK,
			vo.Response{
				Code:    400,
				Message: "请求参数错误: " + err.Error(),
			})
		return
	}
	//如果是uid检测登录则先查询
	if req.Uid != "" {
		user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: req.Uid})
		if user == nil {
			c.JSON(http.StatusOK,
				vo.Response{
					Code:    400,
					Message: "该账号不存在",
				})
			return
		}
		//登录逻辑......
	}

	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "账号登录正常",
	})
}

// 获取账号配置信息
func GetAccountInformService(c *gin.Context) {
	uid := c.Param("uid")
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if user == nil {
		c.JSON(http.StatusOK, gin.H{
			"code": 400,
			"msg":  "该账号不存在",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "拉取信息成功",
		"data": gin.H{
			"user": user,
		},
	})
}

// 登录账号
func LoginUserService(c *gin.Context) {
	// 1. 定义结构体用于接收 JSON
	var req pojo.UserPO
	// 2. 解析 JSON 到结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"code": 400,
			"msg":  "请求参数错误: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
	})
}

// 更新账号信息
func UpdateUserService(c *gin.Context) {
	var req pojo.UserPO

	// 绑定 JSON
	if err := c.ShouldBindJSON(&req); err != nil {
		data, _ := c.GetRawData()
		fmt.Println(string(data))
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "JSON 解析失败",
		})
		return
	}

	// Uid 必须存在
	if req.Uid == "" {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "UID 不能为空",
		})
		return
	}

	// 将结构体转为 map 并过滤空字段
	updateMap := make(map[string]interface{})

	// 手动挑选可修改字段（最安全方式）
	if req.AccountType != "" {
		updateMap["account_type"] = req.AccountType
	}
	if req.Url != "" {
		updateMap["url"] = req.Url
	}
	if req.Account != "" {
		updateMap["account"] = req.Account
	}
	if req.Password != "" {
		updateMap["password"] = req.Password
	}

	// 空字段检查
	if len(updateMap) == 0 {
		c.JSON(200, vo.Response{
			Code:    400,
			Message: "没有可更新的字段",
		})
		return
	}

	// 调用 DAO 更新
	if err := dao.UpdateUser(global.GlobalDB, req.Uid, updateMap); err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 拉取课程列表
func AccountCourseListService(c *gin.Context) {
	// 2. 解析 JSON 到结构体
	uid := c.Param("uid")
	//检测账号是否已存在
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if user == nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: "该账号不存在",
		})
		return
	}

	userActivity := global.GetUserActivity(*user)
	//如果没有活动中的账号则添加活动账号
	if userActivity == nil {
		//构建用户活动
		createActivity := activity.BuildUserActivity(*user)
		userActivity = &createActivity
		global.PutUserActivity(*user, &createActivity)
	}
	//如果是学习通
	if xxt, ok1 := (*userActivity).(activity.XXTAbility); ok1 {
		list, err := xxt.PullCourseList()
		if err != nil {
			fmt.Println(err)
		}
		//转换为标准类型的列表数据
		courseList := []vo.CourseInformResponse{}
		for _, course := range list {
			courseList = append(courseList, vo.CourseInformResponse{
				CourseId:   course.CourseID,
				CourseName: course.CourseName,
				Progress:   float32(course.JobRate),
				Instructor: course.CourseTeacher,
			})
		}
		//fmt.Println(list)
		c.JSON(http.StatusOK, vo.Response{
			Code:    200,
			Message: "拉取信息成功",
			Data:    gin.H{"courseList": courseList},
		})
	}

}

// 开始刷课
func StartBrushService(c *gin.Context) {
	uid := c.Param("uid")
	user, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if err != nil {
		c.JSON(http.StatusOK, vo.Response{
			Code:    400,
			Message: err.Error(),
		})
		return
	}
	userActivity := global.GetUserActivity(*user)
	if userActivity != nil {
		//userActivity.IsRunning = false
	}
	go func() {
		// 调用Start方法
		(*userActivity).Start()
	}()

	c.JSON(200, gin.H{
		"code": 200,
		"msg":  "启动成功",
	})
}

// 停止任务
func StopBrushService(c *gin.Context) {
	uid := c.Param("uid")
	user, err := dao.QueryUser(global.GlobalDB, pojo.UserPO{
		Uid: uid,
	})
	if err != nil {
		c.JSON(400, gin.H{})
	}
	if user == nil {
		c.JSON(400, gin.H{})
		return
	}
	userActivity := global.GetUserActivity(*user)
	if userActivity == nil {
		c.JSON(400, gin.H{})
	}
	// 根据账号类型断言为具体活动类型并设置IsRunning
	if xxt, ok := (*userActivity).(*activity.XXTActivity); ok {
		xxt.IsRunning = false
	} else if yinghua, ok := (*userActivity).(*activity.YingHuaActivity); ok {
		yinghua.IsRunning = true
	}
	(*userActivity).Stop()
	//userActivity.Kill()
	c.JSON(http.StatusOK, vo.Response{
		Code:    200,
		Message: "停止成功",
	})
}


// 获取账号日志 (精简版，待 upstream 合入 local_config 后替换)
func AccountLogsService(c *gin.Context) {
	uid := c.Param("uid")
	user, _ := dao.QueryUser(global.GlobalDB, pojo.UserPO{Uid: uid})
	if user == nil {
		c.JSON(http.StatusOK, vo.Response{Code: 400, Message: "该账号不存在"})
		return
	}
	c.JSON(http.StatusOK, vo.Response{Code: 200, Message: "拉取日志成功", Data: []any{}})
}

func decomposeAiUrl(fullUrl string) (string, string) {
	if fullUrl == "" { return "", "chat" }
	if strings.HasSuffix(fullUrl, "/v1/responses") {
		return strings.TrimSuffix(fullUrl, "/v1/responses"), "responses"
	}
	if strings.HasSuffix(fullUrl, "/v1/chat/completions") {
		return strings.TrimSuffix(fullUrl, "/v1/chat/completions"), "chat"
	}
	idx := strings.LastIndex(fullUrl, "/")
	if idx > 8 { return fullUrl[:idx], "custom:" + fullUrl[idx+1:] }
	return fullUrl, "custom"
}

func resolveEndpoint(endpoint, customEp string) string {
	switch endpoint {
	case "responses": return "/v1/responses"
	case "chat": return "/v1/chat/completions"
	case "custom":
		if customEp != "" { return "/" + customEp }
		return "/v1/chat/completions"
	default: return "/v1/chat/completions"
	}
}

func GetAiConfigService(c *gin.Context) {
	configPath := "./config.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "读取配置文件失败: " + err.Error()})
		return
	}
	raw := make(map[string]any)
	if err := yaml.Unmarshal(data, &raw); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "解析配置文件失败: " + err.Error()})
		return
	}
	aiSetting := map[string]any{"provider": "", "model": "", "apiKey": "", "baseUrl": "", "endpoint": "chat", "aiUrl": ""}
	if setting, ok := raw["setting"].(map[string]any); ok {
		if a, ok := setting["aiSetting"].(map[string]any); ok {
			if v, ok := a["aiType"].(string); ok { aiSetting["provider"] = v }
			if v, ok := a["model"].(string); ok { aiSetting["model"] = v }
			if v, ok := a["API_KEY"].(string); ok { aiSetting["apiKey"] = v }
			if v, ok := a["aiUrl"].(string); ok {
				aiSetting["aiUrl"] = v
				aiSetting["baseUrl"], aiSetting["endpoint"] = decomposeAiUrl(v)
			}
		}
	}
	externalBankUrl := ""
	if setting, ok := raw["setting"].(map[string]any); ok {
		if q, ok := setting["apiQueSetting"].(map[string]any); ok {
			if v, ok := q["url"].(string); ok { externalBankUrl = v }
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "aiSetting": aiSetting, "externalBankUrl": externalBankUrl})
}

func SaveAiConfigService(c *gin.Context) {
	var req struct {
		Provider  string `json:"provider"`
		Model     string `json:"model"`
		ApiKey    string `json:"apiKey"`
		BaseUrl   string `json:"baseUrl"`
		Endpoint  string `json:"endpoint"`
		CustomEp  string `json:"customEndpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "请求参数错误: " + err.Error()})
		return
	}
	fullUrl := req.BaseUrl + resolveEndpoint(req.Endpoint, req.CustomEp)
	configPath := "./config.yaml"
	data, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "读取配置文件失败: " + err.Error()})
		return
	}
	raw := make(map[string]any)
	if err := yaml.Unmarshal(data, &raw); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "解析配置文件失败: " + err.Error()})
		return
	}
	setting, _ := raw["setting"].(map[string]any)
	if setting == nil { setting = make(map[string]any); raw["setting"] = setting }
	aiSetting, _ := setting["aiSetting"].(map[string]any)
	if aiSetting == nil { aiSetting = make(map[string]any) }
	aiSetting["aiType"] = req.Provider
	aiSetting["model"] = req.Model
	aiSetting["API_KEY"] = req.ApiKey
	if fullUrl != "" { aiSetting["aiUrl"] = fullUrl }
	setting["aiSetting"] = aiSetting
	out, err := yaml.Marshal(raw)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "序列化配置失败: " + err.Error()})
		return
	}
	if err := os.WriteFile(configPath, out, 0644); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "写入配置文件失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func TestAiConfigService(c *gin.Context) {
	var req struct {
		Provider  string `json:"provider"`
		Model     string `json:"model"`
		ApiKey    string `json:"apiKey"`
		BaseUrl   string `json:"baseUrl"`
		Endpoint  string `json:"endpoint"`
		CustomEp  string `json:"customEndpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "请求参数错误: " + err.Error()})
		return
	}
	fullUrl := req.BaseUrl + resolveEndpoint(req.Endpoint, req.CustomEp)
	client := &http.Client{Timeout: 10 * time.Second}
	testBody := `{"model":"` + req.Model + `","messages":[{"role":"user","content":"hi"}]}`
	resp, err := client.Post(fullUrl, "application/json", strings.NewReader(testBody))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "连接失败: " + err.Error()})
		return
	}
	defer resp.Body.Close()
	io.ReadAll(resp.Body)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		c.JSON(http.StatusOK, gin.H{"success": true, "status": resp.StatusCode})
	} else {
		c.JSON(http.StatusOK, gin.H{"success": false, "error": "响应状态: " + fmt.Sprintf("%d", resp.StatusCode)})
	}
}
