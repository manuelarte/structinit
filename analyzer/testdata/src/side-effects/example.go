package main

import "fmt"

//go:structinit
type A struct { // want A:"HasFieldOrder\\[b a\\]"
	b, a int
}

func NewA() *A {
	// info: if we reorder the values of a, and b are different.
	incF := inc()
	return &A{
		a: incF(),
		b: incF(),
	}
}

func inc() func() int {
	var i int
	return func() int {
		i++
		return i
	}
}

func main() {
	fmt.Printf("%+v", NewA())
}
