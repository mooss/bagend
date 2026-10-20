// This file implements flag registration.

package flag

////////////////////////////
// Generic implementation //

// Register registers a singleton flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func Register[Dec Decoder[T], T any](
	par *Parser, name string, dest *T, docline string,
) BooleanFlag[T] {
	return registerSingleton[Dec](par, name, dest, docline)
}

// RegisterSlice registers a slice flag to a parser.
// Registering different flags to the same destination is undefined behavior.
func RegisterSlice[Dec Decoder[T], T any](
	par *Parser, name string, dest *[]T, docline string,
) SliceFlag[T] {
	flg := sliceFlag[T, Dec]{
		dest:       dest,
		docLine:    docline,
		namesStore: []string{name},
	}

	par.registerflag(&flg)

	return &flg
}

func registerSingleton[Dec Decoder[T], T any](
	par *Parser, name string, dest *T, docline string,
) *singletonFlag[T, Dec] {
	flg := singletonFlag[T, Dec]{
		dest:       dest,
		docLine:    docline,
		namesStore: []string{name},
	}

	par.registerflag(&flg)
	return &flg
}

func registerAllowed[Dec Decoder[T], T any](
	par *Parser, name string, dest *T, docline string,
) SingletonFlag[T] {
	return &allowedSingletonFlag[T, Dec]{
		singletonFlag: registerSingleton[Dec](par, name, dest, docline),
	}
}

////////////////////////////////
// Specific types: singletons //

func (par *Parser) Int(name string, dest *int, docline string) SingletonFlag[int] {
	return registerAllowed[Int](par, name, dest, docline)
}

func (par *Parser) String(name string, dest *string, docline string) SingletonFlag[string] {
	return registerAllowed[String](par, name, dest, docline)
}

func (par *Parser) Bool(name string, dest *bool, docline string) BooleanFlag[bool] {
	return Register[Bool](par, name, dest, docline)
}

func (par *Parser) IntSlice(name string, dest *[]int, docline string) SliceFlag[int] {
	return RegisterSlice[Int](par, name, dest, docline)
}

func (par *Parser) StringSlice(
	name string, dest *[]string, docline string,
) SliceFlag[string] {
	return RegisterSlice[String](par, name, dest, docline)
}
