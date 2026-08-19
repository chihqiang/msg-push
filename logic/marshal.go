package logic

import (
	"encoding/json"

	"chihqiang/msg-push/model"
)

// marshalStringSlice 序列化字符串切片为 JSON。
func marshalStringSlice(v []string) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// marshalParamMapping 序列化参数映射为 JSON。
func marshalParamMapping(mapping []model.ParamMappingItem) (string, error) {
	data, err := json.Marshal(mapping)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
