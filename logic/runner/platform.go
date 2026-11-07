package runner

import (
	"net/url"
	"strings"

	yatori "github.com/yatori-dev/yatori-go-core/v3"
)

// ParsePlatform 是 yatori.ParsePlatform 的薄封装，额外兼容旧配置里的套壳平台别名。
//
// CANGHUI 是英华套壳平台，v3 没有独立 adapter，旧 console 文档也建议改成 YINGHUA。
// 这里直接做别名映射，避免历史账号打开详情时提示“不支持平台”。
func ParsePlatform(raw string) (yatori.Platform, error) {
	value := strings.ToUpper(strings.TrimSpace(raw))
	if value == "CANGHUI" {
		return yatori.YingHua, nil
	}
	return yatori.ParsePlatform(raw)
}

// NormalizeURL 把平台 URL 统一成 origin，兼容历史数据里保存的完整页面地址。
//
// 例如 https://xxx.haiqikeji.com/student/home -> https://xxx.haiqikeji.com
func NormalizeURL(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if !strings.Contains(value, "://") {
		value = "https://" + value
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return strings.TrimSpace(raw)
	}
	return parsed.Scheme + "://" + parsed.Host
}
