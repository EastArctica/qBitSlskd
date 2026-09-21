package utils

import (
	"bytes"

	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func Some[T any](slice []T, fn func(T) bool) bool {
	for _, elem := range slice {
		if fn(elem) {
			return true
		}
	}

	return false
}

func Find[T any](slice []T, fn func(T) bool) *T {
	for _, elem := range slice {
		if fn(elem) {
			return &elem
		}
	}

	return nil
}

func Average[T Number](slice []T) T {
	if len(slice) == 0 {
		return 0
	}

	var sum T
	for _, elem := range slice {
		sum += elem
	}

	return sum / T(len(slice))
}

func Includes(str string, substr string) bool {
	return bytes.Contains([]byte(str), []byte(substr))
}
