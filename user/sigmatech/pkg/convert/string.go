package convert

import "fmt"

func Concat(values ...interface{}) string {
	var res string

	for _, v := range values {
		res += ToString(v)
	}

	return res
}

func ToString[T any](t T) string {
	return fmt.Sprintf("%v", t)
}
