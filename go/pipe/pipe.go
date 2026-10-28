package pipe

func z[R any]() R {
	var zero R
	return zero
}

/////////////
// Attempt //

// Attempt is a fallible computation producing a Result or an error.
type Attempt[R any] func() (R, error)

// Bend bends a fallible function into an Attempt.
func Bend[R any](fun func() (R, error)) Attempt[R] { return fun }

// Pure wraps a value into an Attempt that always succeeds.
func Pure[R any](value R) Attempt[R] {
	return func() (R, error) { return value, nil }
}

// Pack applies a fallible function to an argument, packaging it as an Attempt.
func Pack[A any, R any](
	fun func(A) (R, error),
	arg A,
) Attempt[R] {
	return func() (R, error) { return fun(arg) }
}

// Then forwards the result of the attempt to a pure function.
// If the attempt fails, fun is not called and the error propagates.
func (try Attempt[A]) Then[R any](fun func(A) R) Attempt[R] {
	return func() (R, error) {
		tmp, err := try()
		if err != nil {
			return z[R](), err
		}

		return fun(tmp), nil
	}
}

// Chain forwards the result of the attempt to a fallible function.
// If the attempt fails, fun is not called and the error propagates.
func (try Attempt[A]) Chain[R any](fun func(A) (R, error)) Attempt[R] {
	return func() (R, error) {
		tmp, err := try()
		if err != nil {
			return z[R](), err
		}

		return fun(tmp)
	}
}

//////////////
// Attempt1 //

// Attempt1 is a fallible computation over one Attempt argument.
type Attempt1[A any, R any] func(Attempt[A]) (R, error)

// Bend1 bends a fallible function of one argument into an Attempt1.
// The bended function fails if its Attempt argument fails.
func Bend1[A any, R any](
	fun func(A) (R, error),
) Attempt1[A, R] {
	return func(try Attempt[A]) (R, error) {
		arg, err := try()
		if err != nil {
			return z[R](), err
		}

		return fun(arg)
	}
}

//////////////
// Attempt2 //

// Attempt2 is a fallible computation over two Attempt arguments.
type Attempt2[A any, B any, R any] func(Attempt[A], Attempt[B]) (R, error)

// Bend2 bends a fallible function of two arguments into an Attempt2.
// The bended function evaluates its arguments in order and fails
// with the first error encountered.
func Bend2[A any, B any, R any](
	fun func(A, B) (R, error),
) Attempt2[A, B, R] {
	return func(try1 Attempt[A], try2 Attempt[B]) (R, error) {
		arg1, err := try1()
		if err != nil {
			return z[R](), err
		}

		arg2, err := try2()
		if err != nil {
			return z[R](), err
		}

		return fun(arg1, arg2)
	}
}
