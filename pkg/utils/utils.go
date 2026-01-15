package utils

import (
	"math/rand"
)

// GenerateRandomString 生成指定长度的随机字符串
func GenerateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

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

// SliceEqual compares two slices and checks whether they have equal content
func SliceEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}

	return true
}

// SliceSubtract subtracts the content of slice a2 from slice a1
func SliceSubtract(a1, a2 []string) []string {
	a := []string{}

	for _, e1 := range a1 {
		found := false

		for _, e2 := range a2 {
			if e1 == e2 {
				found = true
				break
			}
		}

		if !found {
			a = append(a, e1)
		}
	}

	return a
}

// StringMapSubtract subtracts the content of structmap m2 from structmap m1
func StringMapSubtract(m1, m2 map[string]string) map[string]string {
	m := map[string]string{}

	for k1, v1 := range m1 {
		if v2, ok := m2[k1]; ok {
			if v2 != v1 {
				m[k1] = v1
			}
		} else {
			m[k1] = v1
		}
	}

	return m
}

// StructMapSubtract subtracts the content of structmap m2 from structmap m1
func StructMapSubtract(m1, m2 map[string]struct{}) map[string]struct{} {
	m := map[string]struct{}{}

	for k1, v1 := range m1 {
		if _, ok := m2[k1]; !ok {
			m[k1] = v1
		}
	}

	return m
}
