package codec

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"testing"
	"time"
)

type jsonTestDocument struct {
	Value int `json:"value"`
}

type jsonDecodeCase struct {
	name       string
	data       []byte
	want       jsonTestDocument
	hasUnknown bool
	wantErr    bool
}

func generatedJSONDecodeCases(seed int64, count int) []jsonDecodeCase {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec
	cases := make([]jsonDecodeCase, 0, count*2)

	for i := range count {
		value := rng.Intn(1_000_000)
		cases = append(cases,
			jsonDecodeCase{
				name: fmt.Sprintf("generated-%d-known-only", i),
				data: fmt.Appendf([]byte{}, `{"value":%d}`, value),
				want: jsonTestDocument{Value: value},
			},
			jsonDecodeCase{
				name: fmt.Sprintf("generated-%d-with-unknown", i),
				data: fmt.Appendf(
					[]byte{}, `{"value":%d,"extra_%d":%d}`, value, i, rng.Intn(1_000_000),
				),
				want:       jsonTestDocument{Value: value},
				hasUnknown: true,
			},
		)
	}

	return cases
}

//revive:disable:cognitive-complexity
func runJSONDecodeCases(
	t *testing.T,
	cases []jsonDecodeCase,
	decode func([]byte) (jsonTestDocument, error),
	warnOnUnknown bool,
) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			previousOutput := log.Writer()
			log.SetOutput(&output)

			got, err := decode(tc.data)

			log.SetOutput(previousOutput)

			if (err != nil) != tc.wantErr {
				t.Fatalf("decode error = %v, want error %t", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("decoded value = %#v, want %#v", got, tc.want)
			}

			wantWarning := warnOnUnknown && tc.hasUnknown && !tc.wantErr
			gotWarning := strings.Contains(output.String(), "json: cannot unmarshal")
			if gotWarning != wantWarning {
				t.Errorf(
					"warning logged = %t, want %t; log output: %q",
					gotWarning, wantWarning, output.String(),
				)
			}
		})
	}
}

func malformedJSONDecodeCases() []jsonDecodeCase {
	return []jsonDecodeCase{
		{name: "empty-json", data: []byte{}, wantErr: true},
		{name: "truncated-object", data: []byte(`{"value":`), wantErr: true},
		{name: "invalid-value", data: []byte(`{"value":not-a-number}`), wantErr: true},
		{name: "trailing-data", data: []byte(`{"value":1} {}`), wantErr: true},
	}
}

//revive:disable:cognitive-complexity
func TestFromJSONOpt(t *testing.T) {
	cases := generatedJSONDecodeCases(42, 10)
	cases = append(cases, malformedJSONDecodeCases()...)

	seed := time.Now().UnixNano() //revive:disable:epoch-naming
	t.Logf("Using seed %d", seed)
	nondeterministicCases := generatedJSONDecodeCases(seed, 10)
	cases = append(cases, nondeterministicCases...)

	t.Run("ignore-unknown-fields", func(t *testing.T) {
		runJSONDecodeCases(t, cases, func(data []byte) (jsonTestDocument, error) {
			return FromJSONOpt[jsonTestDocument, IgnoreUnknownFields](data)
		}, false)
	})

	t.Run("warn-unknown-fields", func(t *testing.T) {
		runJSONDecodeCases(t, cases, func(data []byte) (jsonTestDocument, error) {
			return FromJSONOpt[jsonTestDocument, WarnUnknownFields](data)
		}, true)
	})

	t.Run("error-unknown-fields", func(t *testing.T) {
		errorCases := make([]jsonDecodeCase, len(cases))
		copy(errorCases, cases)
		for i := range errorCases {
			if errorCases[i].hasUnknown {
				errorCases[i].wantErr = true
				errorCases[i].want = jsonTestDocument{}
			}
		}

		runJSONDecodeCases(t, errorCases, func(data []byte) (jsonTestDocument, error) {
			return FromJSONOpt[jsonTestDocument, ErrorUnknownFields](data)
		}, false)
	})
}

type unsupportedUnknownFieldBehavior struct{}

func (unsupportedUnknownFieldBehavior) unknownFieldBehavior() {}

func TestFromJSONOptUnsupportedBehavior(t *testing.T) {
	got, err := FromJSONOpt[
		jsonTestDocument, unsupportedUnknownFieldBehavior,
	]([]byte(`{"value":1}`))
	if err == nil {
		t.Fatal("expected an error for unsupported unknown-field behavior")
	}
	if got != (jsonTestDocument{}) {
		t.Errorf("decoded value = %#v, want zero value", got)
	}
}
