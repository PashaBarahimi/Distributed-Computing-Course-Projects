package utils

func RemoveDuplicates[T comparable](list []T) []T {
	allKeys := make(map[T]bool)
	res := []T{}
	for _, item := range list {
		if _, ok := allKeys[item]; !ok {
			allKeys[item] = true
			res = append(res, item)
		}
	}
	return res
}
