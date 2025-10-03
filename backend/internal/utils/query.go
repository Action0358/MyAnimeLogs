package utils

import "strconv"

// ParseLimitQuery - クエリパラメータからlimitを解析・検証
// デフォルト値と最大値を指定して、安全な範囲内の値を返す
func ParseLimitQuery(limitStr string, defaultLimit, maxLimit int) int {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
