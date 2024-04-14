package utils

import (
	"fmt"
	"strings"
)

func RemoveDuplicates[T comparable](list []T) []T {
	allKeys := make(map[T]bool)
	var res []T
	for _, item := range list {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = true
			res = append(res, item)
		}
	}
	return res
}

func ToString[T any](list []T) string {
	var strList []string
	for _, item := range list {
		strList = append(strList, fmt.Sprint(item))
	}
	return "[" + strings.Join(strList, ", ") + "]"
}
