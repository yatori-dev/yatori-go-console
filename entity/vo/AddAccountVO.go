package vo

import "yatori-go-console/config"

// AddAccountRequest 新增账号。直接复用 config.User，保证 Web 与 config.yaml 字段一致。
type AddAccountRequest struct {
	config.User
}

// UpdateAccountRequest 更新账号配置。
// Password 为空表示保持原密码不变；其余字段为完整覆盖。
type UpdateAccountRequest struct {
	Uid string `json:"uid"`
	config.User
}

// StartTaskRequest 启动任务时可指定本次只刷哪些课程（课程名称）。
// Courses 为空表示按账号配置执行全部课程。
type StartTaskRequest struct {
	Courses []string `json:"courses"`
}
