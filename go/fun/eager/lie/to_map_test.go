//revive:disable:cognitive-complexity,cyclomatic
package lie_test

import (
	"errors"
	"testing"

	"github.com/mooss/bagend/go/fun/eager/lie"
)

func TestToMap(t *testing.T) {
	t.Run("maps keys and values and keeps the last value for duplicate keys", func(t *testing.T) {
		result := lie.ToMap(
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
		result := lie.ToMap(
			func(x int) int { return x },
			func(x int) int { return x },
			[]int{},
		)
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})
}

func TestToMapKE(t *testing.T) {
	t.Run("maps values by computed keys", func(t *testing.T) {
		result, err := lie.ToMapKE(func(x int) (int, error) {
			return x % 2, nil
		}, []int{1, 2, 3, 4})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := map[int]int{1: 3, 0: 4}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for key, want := range expected {
			if got := result[key]; got != want {
				t.Errorf("key %d: expected %d, got %d", key, want, got)
			}
		}
	})

	t.Run("empty input returns an empty map", func(t *testing.T) {
		result, err := lie.ToMapKE(func(x int) (int, error) {
			return x, nil
		}, []int{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("returns the first error and stops processing", func(t *testing.T) {
		wantErr := errors.New("key mapping failed")
		var calls []int

		result, err := lie.ToMapKE(func(x int) (int, error) {
			calls = append(calls, x)
			if x == 3 {
				return 0, wantErr
			}
			return x, nil
		}, []int{1, 2, 3, 4})

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Errorf("expected nil map on error, got %v", result)
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

func TestToMapVE(t *testing.T) {
	t.Run("uses source values as keys and maps values", func(t *testing.T) {
		result, err := lie.ToMapVE(func(value string) (int, error) {
			return len(value), nil
		}, []string{"a", "bb"})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		expected := map[string]int{"a": 1, "bb": 2}
		if len(result) != len(expected) {
			t.Fatalf("expected length %d, got %d", len(expected), len(result))
		}
		for key, want := range expected {
			if got := result[key]; got != want {
				t.Errorf("key %q: expected %d, got %d", key, want, got)
			}
		}
	})

	t.Run("empty input returns an empty map", func(t *testing.T) {
		result, err := lie.ToMapVE(func(value int) (int, error) {
			return value * 2, nil
		}, []int{})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty map, got %v", result)
		}
	})

	t.Run("returns the first error and stops processing", func(t *testing.T) {
		wantErr := errors.New("value mapping failed")
		var calls []string

		result, err := lie.ToMapVE(func(value string) (int, error) {
			calls = append(calls, value)
			if value == "b" {
				return 0, wantErr
			}
			return len(value), nil
		}, []string{"a", "b", "c"})

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Errorf("expected nil map on error, got %v", result)
		}

		expectedCalls := []string{"a", "b"}
		if len(calls) != len(expectedCalls) {
			t.Fatalf("expected %d calls, got %d", len(expectedCalls), len(calls))
		}
		for i := range expectedCalls {
			if calls[i] != expectedCalls[i] {
				t.Errorf("call %d: expected input %q, got %q", i, expectedCalls[i], calls[i])
			}
		}
	})
}
