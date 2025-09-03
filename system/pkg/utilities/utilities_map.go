package utilities

func Map[T any, U any](arr []T, fn func(T) U) []U {
	mapped := make([]U, len(arr))
	for i, x := range arr {
		mapped[i] = fn(x)
	}

	return mapped
}
