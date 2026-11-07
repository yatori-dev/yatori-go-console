package dao

import (
	"yatori-go-console/entity/pojo"

	"gorm.io/gorm"
)

// QueryStudyProgress 读取某个账号某门课的本地学习进度。
func QueryStudyProgress(db *gorm.DB, account, courseID string) (*pojo.StudyProgressPO, error) {
	var po pojo.StudyProgressPO
	// 用 Find 而不是 First：没有记录属于正常情况，不需要打 “record not found” 日志
	if err := db.Where("account = ? AND course_id = ?", account, courseID).Limit(1).Find(&po).Error; err != nil {
		return nil, err
	}
	if po.Account == "" {
		return nil, nil
	}
	return &po, nil
}

// SaveStudyProgress 保存（覆盖）某个账号某门课的本地学习进度。
func SaveStudyProgress(db *gorm.DB, po *pojo.StudyProgressPO) error {
	return db.Save(po).Error
}

// DeleteStudyProgress 删除某个账号的全部本地学习进度。
func DeleteStudyProgress(db *gorm.DB, account string) error {
	return db.Where("account = ?", account).Delete(&pojo.StudyProgressPO{}).Error
}
