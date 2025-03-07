// This file implements flag registration.

package flag

////////////////////////////
// Generic implementation //

// Register registers a singleton flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func Register[D Decoder[T], T any](par *Parser, name string, dest *T, docline string) FluentFlag[T] {
	flg := singletonflag[T, D]{
		flagBase[T]{
			dest:       dest,
			docLine:    docline,
			namesStore: []string{name},
		},
	}

	par.registerflag(&flg)

	return &flg
}

// RegisterSlice registers a slice flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func RegisterSlice[Dec Decoder[T], T any](par *Parser, name string, dest *[]T, docline string) FluentFlag[[]T] {
	flg := sliceFlag[T, Dec]{
		flagBase[[]T]{
			dest:       dest,
			docLine:    docline,
			namesStore: []string{name},
		},
	}

	par.registerflag(&flg)

	return &flg
}

////////////////////////////////
// Specific types: singletons //

func (par *Parser) Int(name string, dest *int, docline string) FluentFlag[int] {
	return Register[Int](par, name, dest, docline)
}

func (par *Parser) String(name string, dest *string, docline string) FluentFlag[string] {
	return Register[String](par, name, dest, docline)
}

func (par *Parser) Bool(name string, dest *bool, docline string) FluentFlag[bool] {
	return Register[Bool](par, name, dest, docline)
}

func (par *Parser) IntSlice(name string, dest *[]int, docline string) FluentFlag[[]int] {
	return RegisterSlice[Int](par, name, dest, docline)
}

func (par *Parser) StringSlice(name string, dest *[]string, docline string) FluentFlag[[]string] {
	return RegisterSlice[String](par, name, dest, docline)
}
