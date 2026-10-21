//revive:disable:cognitive-complexity,cyclomatic
package lie

import (
	"errors"
	"testing"
)

func testMap[From any, To comparable](
	t *testing.T, name string, input []From, fn func(From) To, expected []To,
) {
	t.Helper()
	t.Run(name, func(t *testing.T) {
		result := Map(fn, input)
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %v, got %v", i, expected[i], result[i])
			}
		}
	})
}

func TestMapFilter(t *testing.T) {
	t.Run("filters nil results and preserves order", func(t *testing.T) {
		result := MapFilter(func(x int) *int {
			if x%2 != 0 {
				return nil
			}
			return &x
		}, []int{1, 2, 3, 4})

		expected := []int{2, 4}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], result[i])
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := MapFilter(func(x int) *int { return &x }, []int{})
		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})

	t.Run("all results nil", func(t *testing.T) {
		result := MapFilter(func(int) *int { return nil }, []int{1, 2})
		if len(result) != 0 {
			t.Fatalf("expected empty result, got %v", result)
		}
	})
}

func TestMap(t *testing.T) {
	testMap(t, "int to square",
		[]int{1, 2, 3, 4},
		func(x int) int { return x * x },
		[]int{1, 4, 9, 16},
	)

	testMap(t, "string to length",
		[]string{"a", "bb", "ccc"},
		func(s string) int { return len(s) },
		[]int{1, 2, 3},
	)

	testMap(t, "empty slice",
		[]int{},
		func(x int) int { return x * 2 },
		[]int{},
	)

	testMap(t, "digit to string",
		[]int{1, 2, 3},
		func(x int) string { return string(rune(x + '0')) }, //nolint:gosec
		[]string{"1", "2", "3"},
	)
}

func TestMapE(t *testing.T) {
	t.Run("maps all elements in order", func(t *testing.T) {
		result, err := MapE(func(x int) (int, error) {
			return x * 2, nil
		}, []int{1, 2, 3})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := []int{2, 4, 6}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for i := range expected {
			if result[i] != expected[i] {
				t.Errorf("at index %d: expected %d, got %d", i, expected[i], result[i])
			}
		}
	})

	t.Run("empty and nil inputs do not invoke function", func(t *testing.T) {
		for _, input := range [][]int{{}, nil} {
			calls := 0
			result, err := MapE(func(x int) (int, error) {
				calls++
				return x, nil
			}, input)

			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if len(result) != 0 {
				t.Errorf("expected empty result, got %v", result)
			}
			if calls != 0 {
				t.Errorf("expected function not to be called, got %d calls", calls)
			}
		}
	})

	t.Run("returns first error and stops processing", func(t *testing.T) {
		wantErr := errors.New("mapping failed")
		var calls []int

		result, err := MapE(func(x int) (int, error) {
			calls = append(calls, x)
			if x == 3 {
				return 0, wantErr
			}
			return x * 2, nil
		}, []int{1, 2, 3, 4})

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Errorf("expected nil result on error, got %v", result)
		}

		expectedCalls := []int{1, 2, 3}
		if len(calls) != len(expectedCalls) {
			t.Fatalf("expected %d calls, got %d", len(expectedCalls), len(calls))
		}
		for i := range expectedCalls {
			if calls[i] != expectedCalls[i] {
				t.Errorf("call %d: expected input %d, got %d", i, expectedCalls[i], calls[i])
			}
		}
	})
}

func TestToMap(t *testing.T) {
	t.Run("maps keys and values and keeps the last value for duplicate keys", func(t *testing.T) {
		result := ToMap(
			func(x int) int { return x % 2 },
			func(x int) int { return x * x },
			[]int{1, 2, 3, 4},
		)

		expected := map[int]int{1: 9, 0: 16}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for key, want := range expected {
			if got := result[key]; got != want {
				t.Errorf("key %d: expected %d, got %d", key, want, got)
			}
		}
	})

	t.Run("empty input", func(t *testing.T) {
		result := ToMap(
			func(x int) int { return x },
			func(x int) int { return x },
			[]int{},
		)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})
}

func TestValidate(t *testing.T) {
	t.Run("validates all elements in order", func(t *testing.T) {
		var visited []int
		err := Validate(func(x int) error {
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

		err := Validate(func(x int) error {
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
			err := Validate(func(int) error {
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
