package sender

import (
	"regexp"
	"sort"
)

// templateVarRe 模板占位符正则。
var templateVarRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// templateKeys 提取模板占位符顺序。
func templateKeys(content string) []string {
	matches := templateVarRe.FindAllStringSubmatch(content, -1)
	keys := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) > 1 {
			keys = append(keys, m[1])
		}
	}
	return keys
}

// sortedParamsFromMap 无模板占位符时的稳定回退：按 key 字典序取值。
// 避免直接遍历 map 产生随机顺序（Go map 迭代无序，会导致短信模板参数内容错位且不稳定）。
func sortedParamsFromMap(mapped map[string]string) []string {
	if len(mapped) == 0 {
		return nil
	}
	keys := make([]string, 0, len(mapped))
	for k := range mapped {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]string, 0, len(keys))
	for _, k := range keys {
		vals = append(vals, mapped[k])
	}
	return vals
}
