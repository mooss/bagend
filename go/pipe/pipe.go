package pipe

import "github.com/mooss/bagend/go/errs"

func z[R any]() R {
	var zero R
	return zero
}

/////////////
// Attempt //

// Attempt is a fallible computation producing a Result or an error.
type Attempt[R any] func() (R, error)

// Pure wraps a value into an Attempt that always succeeds.
func Pure[R any](value R) Attempt[R] {
	return func() (R, error) { return value, nil }
}

// Try transforms a fallible function with no argument into an Attempt.
func Try[R any](fun func() (R, error)) Attempt[R] { return fun }

// Try1 transforms a fallible function with an argument into an Attempt.
func Try1[A any, R any](
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

// ThenTry forwards the result of the attempt to a fallible function.
// If the attempt fails, fun is not called and the error propagates.
func (try Attempt[A]) ThenTry[R any](fun func(A) (R, error)) Attempt[R] {
	return func() (R, error) {
		tmp, err := try()
		if err != nil {
			return z[R](), err
		}

		return fun(tmp)
	}
}

// Validate forwards the result of the attempt to a validation function.
// If the attempt fails, the error is propagated, otherwise the result is validated and the error
// produced is propagated if non-nil.
func (try Attempt[A]) Validate(fun func(A) error) Attempt[A] {
	return func() (A, error) {
		tmp, err := try()
		if err != nil {
			return z[A](), err
		}

		if err := fun(tmp); err != nil {
			return z[A](), err
		}

		return tmp, nil
	}
}

// Err executes the attempt and returns only the error.
func (try Attempt[A]) Err() error {
	_, err := try()
	return err
}

// Wrap wraps the error, if any with the given format and arguments.
func (try Attempt[A]) Wrap(format string, args ...any) Attempt[A] {
	return func() (A, error) { return errs.Wrapper(format, args).Result(try()) }
}

//////////////
// Attempt1 //

// Attempt1 is a fallible computation over one Attempt argument.
type Attempt1[A any, R any] func(Attempt[A]) (R, error)

// Lift1 transforms a fallible function of one argument into an Attempt1.
// The attempt fails if its argument fails.
func Lift1[A any, R any](
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

// Lift2 transforms a fallible function of two arguments into an Attempt2.
// The attempt evaluates its arguments in order and fails with the first error encountered.
func Lift2[A any, B any, R any](
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
