package utils

// UniqueDifference 返回(a-b)差集列表
func UniqueDifference(a, b []string) []string {
	bSet := make(map[string]bool)
	for _, item := range b {
		bSet[item] = true
	}

	seen := make(map[string]bool)
	var diff []string

	for _, item := range a {
		if !bSet[item] && !seen[item] {
			diff = append(diff, item)
			seen[item] = true
		}
	}

	return diff
}

// SliceContains 检查字符串切片是否包含指定字符串
func SliceContains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
