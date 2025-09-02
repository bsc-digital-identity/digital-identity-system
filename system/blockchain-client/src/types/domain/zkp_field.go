package domain

type ZkpField[T any] struct {
	Key   string
	Value T
}
