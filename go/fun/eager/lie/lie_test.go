//revive:disable:cognitive-complexity,cyclomatic
package lie_test

import (
	"errors"
	"testing"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

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
