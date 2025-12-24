package vo

type Response struct {
	Code    int    `json:"code"`    //状态码
	Message string `json:"message"` //返回信息
	Data    any    `json:"data"`    //主要的数据
}
