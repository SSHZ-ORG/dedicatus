package utils

import "google.golang.org/appengine/datastore"

func Contains(s []int, e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

func KeyContains(s []*datastore.Key, e *datastore.Key) int {
	for idx, i := range s {
		if i.Equal(e) {
			return idx
		}
	}
	return -1
}
