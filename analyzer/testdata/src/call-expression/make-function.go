package main

//go:structinit
type A struct { // want A:"HasFieldOrder\\[Name Interests Tags\\]"
	Name      string
	Interests []string
	Tags      []string
}

func DefaultA() A {
	return A{ // want "fields are not initialized in declared order"
		Interests: []string{"Go", "Gopher"},
		Tags:      []string{"foo", "bar"},
		Name:      "John",
	}
}

func NewA(name string) A {
	return A{ // want "fields are not initialized in declared order"
		Interests: make([]string, 0),
		Tags:      make([]string, 0),
		Name:      name,
	}
}
