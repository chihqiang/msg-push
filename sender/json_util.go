package sender

import (
	"encoding/json"
)

// jsonDump 序列化调试数据。
func jsonDump(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

// rawToAny 将 json.RawMessage 转为 any（宽松解析字符串/数字）。
func rawToAny(raw json.RawMessage) any {
	if len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}
