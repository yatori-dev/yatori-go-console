package vo

import "yatori-go-console/config"

type AddAccountRequest struct {
	AccountType   string                `json:"accountType"` //平台类型
	Url           string                `json:"url"`
	Account       string                `json:"account"`                 //账号
	Password      string                `json:"password"`                //密码
	IsProxy       int                   `json:"isProxy"`                 //是否使用代理
	CoursesCustom *config.CoursesCustom `json:"coursesCustom,omitempty"` //可选预设配置
}

type UpdateAccountRequest struct {
	Uid           string               `json:"uid"`
	AccountType   string               `json:"accountType"`
	Url           string               `json:"url"`
	RemarkName    string               `json:"remarkName,omitempty"`
	Account       string               `json:"account"`
	Password      string               `json:"password"`
	IsProxy       int                  `json:"isProxy"`
	InformEmails  []string             `json:"informEmails"`
	CoursesCustom config.CoursesCustom `json:"coursesCustom"`
}
