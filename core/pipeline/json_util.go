package pipeline

import (
	"encoding/json"
)

// jsonMarshal 序列化（失败返回 nil）。
func jsonMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

// jsonUnmarshal 反序列化到 map。
func jsonUnmarshal(data string, dst *map[string]string) error {
	if data == "" {
		return nil
	}
	return json.Unmarshal([]byte(data), dst)
}
