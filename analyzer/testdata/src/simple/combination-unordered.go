package simple

//go:structinit
type Combination struct { // want Combination:"HasFieldOrder\\[name surname age address city postalCode\\]"
	name, surname, age, address, city, postalCode string
}

func comb1() Combination {
	return Combination{ // want "fields are not initialized in declared order"
		name:       "John",
		surname:    "Doe",
		address:    "123 Main St",
		age:        "42",
		city:       "New York",
		postalCode: "10001",
	}
}

func comb2() Combination {
	return Combination{ // want "fields are not initialized in declared order"
		surname:    "Doe",
		name:       "John",
		address:    "123 Main St",
		age:        "42",
		city:       "New York",
		postalCode: "10001",
	}
}
