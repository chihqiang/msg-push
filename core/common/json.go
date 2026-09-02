package common

import "encoding/json"

// MarshalJSON 序列化（失败返回 nil）。
func MarshalJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// MarshalJSONString 序列化调试数据（失败返回 "{}"）。
func MarshalJSONString(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// UnmarshalJSONString 反序列化字符串为 JSON；空串直接返回 nil。
func UnmarshalJSONString(data string, dst any) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), dst)
}
