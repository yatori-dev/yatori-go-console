package pojo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"yatori-go-console/config"
)

// 用户实体类
type UserPO struct {
	Uid            string `gorm:"not null;primaryKey" json:"uid"`                         //唯一Uid
	AccountType    string `gorm:"not null;column:account_type" json:"accountType"`        //账号类型
	Url            string `gorm:"not null;column:url" json:"url"`                         //平台url
	Account        string `gorm:"not null;column:account" json:"account"`                 //账号
	Password       string `gorm:"not null;column:password" json:"password"`               //密码
	UserConfigJson string `gorm:"not null;column:user_config_json" json:"userConfigJson"` //配置文件json
}

type StringArray []string

// 字符串转StringArray
func (s StringArray) Value() (driver.Value, error) {
	//if s == nil {
	//	return "[]", nil
	//}
	return json.Marshal(s)
}

// StringArray转字符串
func (s *StringArray) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte: %T", value)
	}
	return json.Unmarshal(bytes, s)
}

// 用户配置信息json转实体
func (po *UserPO) UserConfigTurnEntity() config.User {
	user := config.User{}
	if err := json.Unmarshal([]byte(po.UserConfigJson), &user); err != nil {
		// 不再 panic 拖垮整个进程：记录错误并返回空配置，由调用方按零值处理
		log.Printf("解析用户配置JSON失败 uid=%s: %v", po.Uid, err)
		return config.User{}
	}
	return user
}
