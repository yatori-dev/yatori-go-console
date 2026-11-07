package dao

import (
	"time"
	"yatori-go-console/entity/pojo"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// 初始化Sqlite
func SqliteInit() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("yatori.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	//自动创建User表
	err = db.AutoMigrate(&pojo.UserPO{})
	if err != nil {
		return nil, err
	}
	//自动创建设置表
	if err = db.AutoMigrate(&pojo.SettingPO{}); err != nil {
		return nil, err
	}
	// 创建学习进度表（本地节点完成记录）
	if err = db.AutoMigrate(&pojo.StudyProgressPO{}); err != nil {
		return nil, err
	}
	////自动创建Course表
	//err = db.AutoMigrate(&entity.Course{})
	//if err != nil {
	//	return nil, err
	//}

	sqlDB, err := db.DB() //数据库连接池
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}
