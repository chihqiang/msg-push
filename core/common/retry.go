package common

import "time"

// LinearBackoff 线性退避延迟：第 n 次重试延迟 n 秒（上限 60s）。
func LinearBackoff(retryCount int) time.Duration {
	seconds := retryCount + 1
	if seconds > 60 {
		seconds = 60
	}
	return time.Duration(seconds) * time.Second
}
