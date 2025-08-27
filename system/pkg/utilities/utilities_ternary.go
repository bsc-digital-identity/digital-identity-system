package utilities

func Ternary[T any](cond bool, evalTrue, evalFalse T) T {
	if cond {
		return evalTrue
	} else {
		return evalFalse
	}
}
