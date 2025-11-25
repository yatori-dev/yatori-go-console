package pojo

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// 用户实体类
type UserPO struct {
	Uid            string      `gorm:"not null;primaryKey" json:"uid"`                       //唯一Uid
	AccountType    string      `gorm:"not null;column:account_type" json:"accountType"`      //账号类型
	Url            string      `gorm:"not null;column:url" json:"url"`                       //平台url
	Account        string      `gorm:"not null;column:account" json:"account"`               //账号
	Password       string      `gorm:"not null;column:password" json:"password"`             //密码
	IsProxy        int         `gorm:"not null;column:is_proxy" json:"isProxy"`              //是否代理IP
	EmailInformSw  int         `gorm:"not null;column:email_inform_sw" json:"emailInformSw"` //是否开启邮箱通知
	InformEmails   StringArray `gorm:"not null;column:inform_emails" json:"informEmails"`
	WeLearnTime    string      `gorm:"not null;column:weLearn_time" json:"weLearnTime" `        //WeLearn设置刷学时的时候范围
	CxNode         int         `gorm:"not null;column:cxNode" json:"cxNode" `                   //学习通多任务点模式下设置同时任务点数量
	ShuffleSw      int         `gorm:"not null;column:shuffle_sw" json:"shuffleSw"`             //是否打乱顺序学习，1为打乱顺序，0为不打乱
	VideoModel     int         `gorm:"not null;column:video_model" json:"videoModel"`           //观看视频模式
	AutoExam       int         `gorm:"not null;column:auto_exam" json:"autoExam" `              //是否自动考试
	ExamAutoSubmit int         `gorm:"not null;column:exam_auto_submit" json:"examAutoSubmit" ` //是否自动提交试卷
	ExcludeCourses StringArray `gorm:"not null;column:exclude_courses" json:"excludeCourses"`   //排除的课程，包含的课程，这个里面存ID，一般是课程数据里面的唯一标识
	IncludeCourses StringArray `gorm:"not null;column:include_courses" json:"includeCourses"`   //包含的课程，这个里面存ID，一般是课程数据里面的唯一标识
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
