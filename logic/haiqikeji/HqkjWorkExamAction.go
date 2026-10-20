package haiqikeji

// ============================================================================
// 海旗科技（HQKJ）AI 自动答题接入层（console 薄封装）
//
// 作业 + 考试的实际写题逻辑已下沉至核心库：
//   github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji.HqkjAutoAnswerAction
//   接口层：github.com/yatori-dev/yatori-go-core/api/haiqikeji.HqkjWorkExamApi.go
//
// 本文件只负责把 console 的配置结构（config.Setting / config.User）转换为核心库入参，
// 保持 HqkjPart.go 中的调用点 runAutoAnswer(...) 不变。
// ============================================================================

import (
	"yatori-go-console/config"
	"yatori-go-console/global"

	"github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	hqkjApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	coreConfig "github.com/yatori-dev/yatori-go-core/config"
)

// runAutoAnswer 海旗 AI 自动答题总入口（作业 + 考试）。
// 在 nodeListStudy 视频处理之后调用；仅当 AutoExam==1（AI 答题）时由核心库生效。
func runAutoAnswer(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, nodeList []haiqikeji.HqkjNode) {
	haiqikeji.HqkjAutoAnswerAction(cache, *course, nodeList, haiqikeji.HqkjAnswerSetting{
		AccountType:    global.AccountTypeStr[user.AccountType],
		Account:        user.Account,
		AutoExam:       user.CoursesCustom.AutoExam,       // 0=不答题，1=AI答题
		ExamAutoSubmit: user.CoursesCustom.ExamAutoSubmit, // 0=只答题不交卷，1=自动交卷
		AiSetting: coreConfig.AiSetting{
			AiType: setting.AiSetting.AiType,
			AiUrl:  setting.AiSetting.AiUrl,
			Model:  setting.AiSetting.Model,
			APIKEY: setting.AiSetting.APIKEY,
		},
	})
}
