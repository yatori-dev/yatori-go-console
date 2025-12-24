package vo

type AddAccountRequest struct {
	AccountType string `json:"accountType"` //平台类型
	Url         string `json:"url"`
	Account     string `json:"account"`  //账号
	Password    string `json:"password"` //密码
}
