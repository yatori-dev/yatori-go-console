package vo

type AccountLoginCheckRequest struct {
	Uid         string `json:"uid"`
	AccountType string `json:"accountType"`
	Url         string `json:"url"`
	Account     string `json:"account"`
	Password    string `json:"password"`
}
