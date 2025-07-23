package main

func Ternary[T any](cond bool, eval_true, eval_false T) T {
	if cond {
		return eval_true
	} else {
		return eval_false
	}
}
