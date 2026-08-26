package pipeline

import (
	"regexp"
)

// cnMobileRe 中国大陆手机号正则。
var cnMobileRe = regexp.MustCompile(`^1[3-9]\d{9}$`)
