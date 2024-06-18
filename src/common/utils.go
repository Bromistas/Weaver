package common

import "sort"

func insertIntoSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] >= item })
	slice = append(slice, 0)
	copy(slice[i+1:], slice[i:])
	slice[i] = item
	return slice
}

func removeFromSorted(slice []int, item int) []int {
	i := sort.Search(len(slice), func(i int) bool { return slice[i] > item })
	if i < len(slice) && slice[i] == item {
		slice = append(slice[:i], slice[i+1:]...)
	}
	return slice
}
