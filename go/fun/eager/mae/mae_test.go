//revive:disable:cognitive-complexity
//nolint:gosec
package mae

import (
	"errors"
	"testing"
)

func assertMapEqual[K comparable, V comparable](t *testing.T, got, want map[K]V) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected map length %d, got %d", len(want), len(got))
	}
	for key, wantValue := range want {
		gotValue, ok := got[key]
		if !ok {
			t.Errorf("expected key %v in result", key)
			continue
		}
		if gotValue != wantValue {
			t.Errorf("key %v: expected %v, got %v", key, wantValue, gotValue)
		}
	}
}

func TestMap(t *testing.T) {
	t.Run("maps values and preserves keys", func(t *testing.T) {
		result := MapV(func(value int) string {
			return string(rune('a' + value))
		}, map[string]int{"first": 0, "second": 1})

		assertMapEqual(t, result, map[string]string{"first": "a", "second": "b"})
	})

	t.Run("empty and nil inputs", func(t *testing.T) {
		for _, input := range []map[string]int{{}, nil} {
			calls := 0
			result := MapV(func(value int) int {
				calls++
				return value * 2
			}, input)

			if len(result) != 0 {
				t.Errorf("expected empty result, got %v", result)
			}
			if calls != 0 {
				t.Errorf("expected function not to be called, got %d calls", calls)
			}
		}
	})
}

func TestMapE(t *testing.T) {
	t.Run("maps values and preserves keys", func(t *testing.T) {
		result, err := MapVE(func(value int) (string, error) {
			return string(rune('a' + value)), nil
		}, map[string]int{"first": 0, "second": 1})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		assertMapEqual(t, result, map[string]string{"first": "a", "second": "b"})
	})

	t.Run("empty and nil inputs", func(t *testing.T) {
		for _, input := range []map[string]int{{}, nil} {
			calls := 0
			result, err := MapVE(func(value int) (int, error) {
				calls++
				return value, nil
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

	t.Run("returns the first error and stops", func(t *testing.T) {
		wantErr := errors.New("mapping failed")
		calls := 0

		result, err := MapVE(func(value int) (int, error) {
			calls++
			if calls == 1 {
				return 0, wantErr
			}
			return value, nil
		}, map[string]int{"first": 1, "second": 2, "third": 3})

		if !errors.Is(err, wantErr) {
			t.Fatalf("expected error %v, got %v", wantErr, err)
		}
		if result != nil {
			t.Errorf("expected nil result on error, got %v", result)
		}
		if calls != 1 {
			t.Errorf("expected mapping to stop after one call, got %d calls", calls)
		}
	})
}

func TestMapFilter(t *testing.T) {
	t.Run("filters nil results and preserves keys", func(t *testing.T) {
		result := MapVFilter(func(value int) *string {
			if value%2 != 0 {
				return nil
			}
			mapped := string(rune('a' + value))
			return &mapped
		}, map[string]int{"one": 1, "two": 2, "three": 3, "four": 4})

		assertMapEqual(t, result, map[string]string{"two": "c", "four": "e"})
	})

	t.Run("all results nil", func(t *testing.T) {
		result := MapVFilter(func(int) *string { return nil }, map[string]int{"one": 1})
		if len(result) != 0 {
			t.Errorf("expected empty result, got %v", result)
		}
	})

	t.Run("empty and nil inputs", func(t *testing.T) {
		for _, input := range []map[string]int{{}, nil} {
			calls := 0
			result := MapVFilter(func(value int) *int {
				calls++
				return &value
			}, input)

			if len(result) != 0 {
				t.Errorf("expected empty result, got %v", result)
			}
			if calls != 0 {
				t.Errorf("expected function not to be called, got %d calls", calls)
			}
		}
	})
}
