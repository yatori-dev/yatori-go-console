package dao

import (
	"errors"
	"yatori-go-console/entity/pojo"

	"gorm.io/gorm"
)

// 插入用户
func InsertUser(db *gorm.DB, user *pojo.UserPO) error {
	if err := db.Create(&user).Error; err != nil {
		return errors.New("插入数据失败: " + err.Error())
	}
	//log2.Print(log2.DEBUG, "插入数据成功")
	return nil
}

// 删除指定uid用户
func DeleteUser(db *gorm.DB, uid string) error {
	if err := db.Where("uid = ?", uid).Delete(&pojo.UserPO{}).Error; err != nil {
		return errors.New("删除用户失败: " + err.Error())
	}
	return nil
}

// 查询用户
func QueryUsers(db *gorm.DB, page, pageSize int) ([]pojo.UserPO, int64, error) {
	var users []pojo.UserPO
	var total int64

	if err := db.Model(&pojo.UserPO{}).Count(&total).Error; err != nil {
		return nil, 0, errors.New("统计用户总数失败: " + err.Error())
	}

	offset := (page - 1) * pageSize

	if err := db.
		Limit(pageSize).
		Offset(offset).
		Order("uid ASC").
		Find(&users).Error; err != nil {
		return nil, 0, errors.New("查询用户失败: " + err.Error())
	}

	return users, total, nil
}

// QueryUser 查询单个用户（根据传入的 User 字段自动匹配）
func QueryUser(db *gorm.DB, cond pojo.UserPO) (*pojo.UserPO, error) {
	var user pojo.UserPO

	query := db.Model(&pojo.UserPO{})

	// 动态拼接查询条件（如果字段非空才加入查询）
	if cond.Uid != "" {
		query = query.Where("uid = ?", cond.Uid)
	}
	if cond.AccountType != "" {
		query = query.Where("account_type = ?", cond.AccountType)
	}
	if cond.Url != "" {
		query = query.Where("url = ?", cond.Url)
	}
	if cond.Account != "" {
		query = query.Where("account = ?", cond.Account)
	}
	if cond.Password != "" {
		query = query.Where("password = ?", cond.Password)
	}

	// 执行查询
	if err := query.First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户不存在")
		}
		return nil, errors.New("查询用户失败: " + err.Error())
	}

	return &user, nil
}

// UpdateUser 修改指定 UID 用户信息
func UpdateUser(db *gorm.DB, uid string, updateData map[string]interface{}) error {
	if uid == "" {
		return errors.New("更新失败: UID 不能为空")
	}

	// 执行更新
	if err := db.Model(&pojo.UserPO{}).
		Where("uid = ?", uid).
		Updates(updateData).Error; err != nil {

		return errors.New("更新用户失败: " + err.Error())
	}

	return nil
}
