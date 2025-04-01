package services

import (
	"fmt"
)

func ListCompare(a, b []string) (added []string, removed []string, intersection []string) {

	// 建立一個 map 來追蹤 a 中的元素
	m := make(map[string]bool)
	for _, item := range a {
		m[item] = true
	}
	for _, item := range b {
		if _, ok := m[item]; !ok {
			// 如果 b 中的元素不在 a 中，加入 added
			added = append(added, item)
		} else {
			// 如果 b 中的元素也在 a 中，加入 intersection
			intersection = append(intersection, item)
			delete(m, item)
		}
	}
	// 遍歷 map 中剩下的元素，這些是被移除的元素
	for item := range m {
		removed = append(removed, item)
	}
	// return added, removed
	fmt.Println("add:", added)
	fmt.Println("removed: ", removed)
	fmt.Println("Intersection:", intersection)

	return added, removed, intersection
}
