package lazy

import "iter"

// SplitErr transforms an iter.Seq2[T, error] into an iter.Seq[T] plus an error reporter.
// Once the returned sequence has been iterated, the error reporter must be checked.
func SplitErr[T any](seq iter.Seq2[T, error]) (iter.Seq[T], func() error) {
	var err error
	res := func(yield func(T) bool) {
		err = nil
		for el, newerr := range seq {
			if newerr != nil {
				err = newerr
				return
			}

			if !yield(el) {
				return
			}
		}
	}

	return res, func() error { return err }
}
