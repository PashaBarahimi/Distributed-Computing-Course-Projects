package matcher

import (
	"strings"
)

var serverOrders = []string{
	"banana",
	"apple",
	"orange",
	"grape",
	"red apple",
	"kiwi",
	"mango",
	"pear",
	"cherry",
	"green apple",
}

func MatchOrder(order string) []string {
	var result []string
	for _, serverOrder := range serverOrders {
		if strings.Contains(serverOrder, order) {
			result = append(result, serverOrder)
		}
	}
	return result
}
