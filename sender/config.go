package sender

import (
	"strconv"
	"strings"

	"chihqiang/msg-push/model"
)

// configMap 反序列化服务商配置。
func configMap(pa *model.ProviderAccount) (map[string]any, error) {
	if pa == nil {
		return map[string]any{}, nil
	}
	return pa.GetConfig()
}

// strVal 读取字符串配置。
func strVal(cfg map[string]any, key string) string {
	if cfg == nil {
		return ""
	}
	switch v := cfg[key].(type) {
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case int:
		return strconv.Itoa(v)
	}
	return ""
}

// intVal 读取整型配置（兼容字符串/数字）。
func intVal(cfg map[string]any, key string) int {
	if cfg == nil {
		return 0
	}
	switch v := cfg[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}

// boolVal 读取布尔配置。
func boolVal(cfg map[string]any, key string) bool {
	if cfg == nil {
		return false
	}
	switch v := cfg[key].(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	case float64:
		return v != 0
	}
	return false
}

// providerAccountLike 最小配置读取接口（避免重复实现 GetConfig）。
type providerAccountLike struct {
	Config string
}
