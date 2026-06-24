package comments

//go:structinit
type Comments struct { // want Comments:"HasFieldOrder\\[Name Surname\\]"
	Name    string
	Surname string
}

func newComments() Comments {
	return Comments{ // want "fields are not initialized in declared order"
		// setting the default value for Surname
		Surname: "Doe",
		// setting the default value for Name
		Name: "John",
	}
}

func newInlineComments() Comments {
	return Comments{ // want "fields are not initialized in declared order"
		Surname: "Doe",  // setting the default value for Surname
		Name:    "John", // setting the default value for Name
	}
}
