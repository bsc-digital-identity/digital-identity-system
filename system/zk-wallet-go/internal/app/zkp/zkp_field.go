package zkp

type ZkpField[T any] struct {
	Key   string
	Value T
}
