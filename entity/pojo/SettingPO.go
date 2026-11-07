package pojo

// SettingPO 全局设置键值表。当前只使用 key=global 一行，值为 config.Setting 的 JSON。
type SettingPO struct {
	Key   string `gorm:"primaryKey;column:key" json:"key"`
	Value string `gorm:"column:value;type:text;not null" json:"value"`
}
