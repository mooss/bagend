//revive:disable:cognitive-complexity,cyclomatic,function-length
package lie_test

import (
	"errors"
	"iter"
	"testing"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

func TestCollect2(t *testing.T) {
	t.Run("collects all elements in order", func(t *testing.T) {
		source := iter.Seq2[int, error](func(yield func(int, error) bool) {
			for _, v := range []int{1, 2, 3} {
				if !yield(v, nil) {
					return
				}
			}
		})

		result, err := lie.CollectE(source)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{1, 2, 3}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], result[i])
			}
		}
	})

	t.Run("supports arbitrary element types", func(t *testing.T) {
		source := iter.Seq2[string, error](func(yield func(string, error) bool) {
			for _, v := range []string{"a", "b"} {
				if !yield(v, nil) {
					return
				}
			}
		})

		result, err := lie.CollectE(source)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []string{"a", "b"}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %q, got %q", i, expected[i], result[i])
			}
		}
	})

	t.Run("empty iterator returns empty without error", func(t *testing.T) {
		source := iter.Seq2[int, error](func(_ func(int, error) bool) {})

		result, err := lie.CollectE(source)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})

	t.Run("returns first error", func(t *testing.T) {
		wantErr := errors.New("collect failed")
		source := iter.Seq2[int, error](func(yield func(int, error) bool) {
			if !yield(1, nil) {
				return
			}
			if !yield(0, wantErr) {
				return
			}
			if !yield(3, nil) {
				return
			}
		})

		result, err := lie.CollectE(source)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Fatalf("expected nil result on error, got %v", result)
		}
	})

	t.Run("stops at first error without consuming rest", func(t *testing.T) {
		wantErr := errors.New("boom")
		yieldedAfterErr := false
		source := iter.Seq2[int, error](func(yield func(int, error) bool) {
			if !yield(1, nil) {
				return
			}
			if !yield(2, wantErr) {
				return
			}
			yieldedAfterErr = true
			yield(3, nil)
		})

		result, err := lie.CollectE(source)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Fatalf("expected nil result on error, got %v", result)
		}
		if yieldedAfterErr {
			t.Error("expected iteration to stop at first error")
		}
	})

	t.Run("error on first element returns nil", func(t *testing.T) {
		wantErr := errors.New("immediate failure")
		source := iter.Seq2[int, error](func(yield func(int, error) bool) {
			yield(0, wantErr)
		})

		result, err := lie.CollectE(source)
		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Fatalf("expected nil result on error, got %v", result)
		}
	})
}

func TestEnsureFinal(t *testing.T) {
	tests := []struct {
		name     string
		source   []int
		final    int
		expected []int
	}{
		{
			name:     "empty slice",
			source:   []int{},
			final:    7,
			expected: []int{7},
		},
		{
			name:     "appends missing final element",
			source:   []int{1, 2},
			final:    3,
			expected: []int{1, 2, 3},
		},
		{
			name:     "does not duplicate existing final element",
			source:   []int{1, 2, 3},
			final:    3,
			expected: []int{1, 2, 3},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := lie.EnsureFinal(test.source, test.final)
			if len(result) != len(test.expected) {
				t.Fatalf("expected length %d, got %d", len(test.expected), len(result))
			}
			for i := range test.expected {
				if result[i] != test.expected[i] {
					t.Errorf("at index %d: expected %d, got %d", i, test.expected[i], result[i])
				}
			}
		})
	}
}

func TestPartition(t *testing.T) {
	t.Run("splits elements by predicate preserving order", func(t *testing.T) {
		source := []int{1, 2, 3, 4, 5, 6}
		evens, odds := lie.Partition(func(x int) bool { return x%2 == 0 }, source)

		expectedEvens := []int{2, 4, 6}
		if len(evens) != len(expectedEvens) {
			t.Fatalf("expected length %d, got %d", len(expectedEvens), len(evens))
		}
		for i := range expectedEvens {
			if evens[i] != expectedEvens[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expectedEvens[i], evens[i])
			}
		}

		expectedOdds := []int{1, 3, 5}
		if len(odds) != len(expectedOdds) {
			t.Fatalf("expected length %d, got %d", len(expectedOdds), len(odds))
		}
		for i := range expectedOdds {
			if odds[i] != expectedOdds[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expectedOdds[i], odds[i])
			}
		}
	})

	t.Run("nil input returns empty slices", func(t *testing.T) {
		matched, unmatched := lie.Partition(func(int) bool { return true }, nil)

		if len(matched) != 0 {
			t.Errorf("expected empty matched, got %v", matched)
		}
		if len(unmatched) != 0 {
			t.Errorf("expected empty unmatched, got %v", unmatched)
		}
	})
}

func TestUnzip(t *testing.T) {
	t.Run("splits elements into two slices", func(t *testing.T) {
		source := []string{"a", "bb", "ccc"}
		lens, runes := lie.Unzip(func(s string) (int, rune) {
			return len(s), rune(s[0])
		}, source)

		expectedLens := []int{1, 2, 3}
		if len(lens) != len(expectedLens) {
			t.Fatalf("expected length %d, got %d", len(expectedLens), len(lens))
		}
		for i := range expectedLens {
			if lens[i] != expectedLens[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expectedLens[i], lens[i])
			}
		}

		expectedRunes := []rune{'a', 'b', 'c'}
		if len(runes) != len(expectedRunes) {
			t.Fatalf("expected length %d, got %d", len(expectedRunes), len(runes))
		}
		for i := range expectedRunes {
			if runes[i] != expectedRunes[i] {
				t.Errorf("at index %d: expected %q, got %q", i, expectedRunes[i], runes[i])
			}
		}
	})

	t.Run("empty input returns empty slices", func(t *testing.T) {
		part1, part2 := lie.Unzip(func(x int) (int, int) { return x, x }, []int{})

		if len(part1) != 0 {
			t.Errorf("expected empty part1, got %v", part1)
		}
		if len(part2) != 0 {
			t.Errorf("expected empty part2, got %v", part2)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates all elements in order", func(t *testing.T) {
		var visited []int
		err := lie.Validate(func(x int) error {
			visited = append(visited, x)
			return nil
		}, []int{1, 2, 3})

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		expected := []int{1, 2, 3}
		if len(visited) != len(expected) {
			t.Fatalf("expected %d elements visited, got %d", len(expected), len(visited))
		}
		for i := range expected {
			if visited[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], visited[i])
			}
		}
	})

	t.Run("returns first error and stops processing", func(t *testing.T) {
		wantErr := errors.New("validation failed")
		var visited []int

		err := lie.Validate(func(x int) error {
			visited = append(visited, x)
			if x == 2 {
				return wantErr
			}
			return nil
		}, []int{1, 2, 3})

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		expected := []int{1, 2}
		if len(visited) != len(expected) {
			t.Fatalf("expected %d elements visited, got %d", len(expected), len(visited))
		}
		for i := range expected {
			if visited[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], visited[i])
			}
		}
	})

	t.Run("empty and nil inputs do not invoke function", func(t *testing.T) {
		for _, input := range [][]int{{}, nil} {
			calls := 0
			err := lie.Validate(func(int) error {
				calls++
				return nil
			}, input)

			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
			if calls != 0 {
				t.Errorf("expected function not to be called, got %d calls", calls)
			}
		}
	})
}
