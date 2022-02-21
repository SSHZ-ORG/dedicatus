package utils

import "google.golang.org/appengine/v2/datastore"

func Contains(s []int, e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func FindKeyIndex(s []*datastore.Key, e *datastore.Key) int {
	for idx, i := range s {
		if i.Equal(e) {
			return idx
		}
	}
	return -1
}

func Remove(s []int, e int) []int {
	for i, k := range s {
		if k == e {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
