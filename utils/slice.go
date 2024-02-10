package utils

func Contains[T comparable](data []T, item T) bool {
	for _, v := range data {
		if v == item {
			return true
		}
	}

	return false
}
