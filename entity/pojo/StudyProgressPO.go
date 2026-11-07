package pojo

import "time"

// StudyProgressPO 记录本地确认完成的学习节点。
//
// 学习通的课程完成率（jobRate）和章节卡片状态都有缓存，刚刷完的节点经常还是旧数据。
// 这里把“平台确认过”和“我们自己提交成功过”的节点一起落库，用于跳过节点和覆盖进度显示。
type StudyProgressPO struct {
	Account   string    `gorm:"primaryKey;column:account"`
	CourseID  string    `gorm:"primaryKey;column:course_id"`
	Total     int       `gorm:"column:total"`
	DoneIDs   string    `gorm:"column:done_ids;type:text"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (StudyProgressPO) TableName() string {
	return "study_progress"
}
