package dao

import (
	"encoding/json"
	"errors"

	"yatori-go-console/config"
	"yatori-go-console/entity/pojo"

	"gorm.io/gorm"
)

const globalSettingKey = "global"

// LoadSetting 从数据库读取全局设置；不存在时返回 nil, nil。
func LoadSetting(db *gorm.DB) (*config.Setting, error) {
	var row pojo.SettingPO
	if err := db.Where("key = ?", globalSettingKey).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	setting := config.Setting{}
	if err := json.Unmarshal([]byte(row.Value), &setting); err != nil {
		return nil, err
	}
	return &setting, nil
}

// SaveSetting 保存全局设置。
func SaveSetting(db *gorm.DB, setting config.Setting) error {
	data, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	return db.Save(&pojo.SettingPO{Key: globalSettingKey, Value: string(data)}).Error
}
