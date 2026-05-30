package dto

import "yatori-go-console/config"

type ConfigManagerUser struct {
	Uid           string               `json:"uid"`
	AccountType   string               `json:"accountType"`
	URL           string               `json:"url"`
	RemarkName    string               `json:"remarkName,omitempty"`
	Account       string               `json:"account"`
	Password      string               `json:"password"`
	IsProxy       int                  `json:"isProxy"`
	InformEmails  []string             `json:"informEmails"`
	CoursesCustom config.CoursesCustom `json:"coursesCustom"`
	Deletable     bool                 `json:"deletable,omitempty"`
}
