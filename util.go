package main

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func ensure(err error) {
	if err != nil {
		panic(err)
	}
}

func cond[T any](cond bool, trueVal, falseVal T) T {
	if cond {
		return trueVal
	} else {
		return falseVal
	}
}
