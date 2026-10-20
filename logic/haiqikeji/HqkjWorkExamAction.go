package haiqikeji

// ============================================================================
// 海旗科技（HQKJ）自动答题接入层（console 薄封装）
//
// 作业 + 考试的实际写题逻辑已下沉至核心库：
//   AI 答题：github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji.HqkjAutoAnswerAction
//   内置标答：github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji.HqkjAutoFillBuiltinAnswerAction
//   接口层：github.com/yatori-dev/yatori-go-core/api/haiqikeji.HqkjWorkExamApi.go
//
// 本文件只负责把 console 的配置结构（config.Setting / config.User）转换为核心库入参，
// 保持 HqkjPart.go 中的调用点 runAutoAnswer(...) 不变。
//
// 海旗 autoExam 取值约定：
//   0 = 不答题
//   1 = AI 答题（核心库优先采用平台内置标答，缺失时再调用 AI 补答）
//   2 = 内置标答答题（仅从题目自带标答中提取答案填写，不依赖 AI）
// ============================================================================

import (
	"yatori-go-console/config"
	"yatori-go-console/global"

	"github.com/yatori-dev/yatori-go-core/aggregation/haiqikeji"
	hqkjApi "github.com/yatori-dev/yatori-go-core/api/haiqikeji"
	coreConfig "github.com/yatori-dev/yatori-go-core/config"
)

// runAutoAnswer 海旗自动答题总入口（作业 + 考试）。
// 在 nodeListStudy 视频处理之后调用，按 autoExam 选择答案来源：
//   - 1：AI 答题（平台有标答时优先用标答）
//   - 2：仅用平台内置标答答题（无需 AI 配置）
func runAutoAnswer(setting config.Setting, user *config.User, cache *hqkjApi.HqkjUserCache, course *haiqikeji.HqkjCourse, nodeList []haiqikeji.HqkjNode) {
	answerSetting := haiqikeji.HqkjAnswerSetting{
		AccountType:    global.AccountTypeStr[user.AccountType],
		Account:        user.Account,
		AutoExam:       user.CoursesCustom.AutoExam,       // 0=不答题，1=AI答题，2=内置标答答题
		ExamAutoSubmit: user.CoursesCustom.ExamAutoSubmit, // 0=只答题不交卷，1=自动交卷
		AiSetting: coreConfig.AiSetting{
			AiType: setting.AiSetting.AiType,
			AiUrl:  setting.AiSetting.AiUrl,
			Model:  setting.AiSetting.Model,
			APIKEY: setting.AiSetting.APIKEY,
		},
	}
	switch user.CoursesCustom.AutoExam {
	case 1: // AI 答题（核心库内部会先优先使用平台内置标答，缺失时回退 AI）
		haiqikeji.HqkjAutoAnswerAction(cache, *course, nodeList, answerSetting)
	case 2: // 内置标答答题：只从题目自带标答中提取答案，不调用 AI、不校验 AI 配置
		haiqikeji.HqkjAutoFillBuiltinAnswerAction(cache, *course, nodeList, answerSetting)
	}
}
