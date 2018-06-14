package utils

func Contains(s []int, e int) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}
