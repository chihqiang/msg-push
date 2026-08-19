// Package template 提供模板渲染与参数映射能力（消费端使用）。
package worker

import (
	"regexp"

	"chihqiang/msg-push/model"
)

// 简单占位符正则：{variable}。
var simplePlaceholderRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// Renderer 模板渲染器接口。
type Renderer interface {
	// RenderSimple 渲染简单模板（{variable} 占位符），参数不存在时保留原占位符。
	RenderSimple(content string, params map[string]string) string
	// MapParams 按参数映射配置将系统参数转换为供应商变量名到值的映射。
	MapParams(params map[string]string, mapping []model.ParamMappingItem) map[string]string
}

// helper 渲染器实现。
type helper struct{}

// NewRenderer 创建渲染器。
func NewRenderer() Renderer {
	return &helper{}
}

// RenderSimple 渲染 {variable} 占位符：参数存在则替换，否则保留原占位符。
func (h *helper) RenderSimple(content string, params map[string]string) string {
	if content == "" {
		return ""
	}
	return simplePlaceholderRe.ReplaceAllStringFunc(content, func(match string) string {
		name := match[1 : len(match)-1]
		if v, ok := params[name]; ok {
			return v
		}
		return match
	})
}

// MapParams 参数映射：fixed 用配置值，mapping 取系统参数对应值。
func (h *helper) MapParams(params map[string]string, mapping []model.ParamMappingItem) map[string]string {
	result := make(map[string]string, len(mapping))
	for _, item := range mapping {
		var value string
		switch item.Type {
		case model.ParamMappingTypeFixed:
			value = item.Value
		default: // mapping（或空）
			if v, ok := params[item.SystemVar]; ok {
				value = v
			}
		}
		result[item.ProviderVar] = value
	}
	return result
}

// SameNameTemplateParams 仅同名变量自动映射：只透传系统参数中与供应商变量同名的。
func SameNameTemplateParams(templateParams map[string]string, providerVariables []string) map[string]string {
	mapped := make(map[string]string, len(providerVariables))
	for _, variable := range providerVariables {
		if value, ok := templateParams[variable]; ok {
			mapped[variable] = value
		}
	}
	return mapped
}
