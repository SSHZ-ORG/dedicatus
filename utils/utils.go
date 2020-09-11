package utils

import "time"

func JST() *time.Location {
	l, _ := time.LoadLocation("Asia/Tokyo")
	return l
}
