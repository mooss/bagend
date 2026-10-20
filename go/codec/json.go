package codec

import (
	jsonv2 "encoding/json/v2"
	"errors"
	"log"

	"github.com/mooss/bagend/go/fun"
)

// UnknownFieldBehavior is implemented by the types that select how FromJSONOpt handles unknown JSON
// object members.
type UnknownFieldBehavior interface {
	unknownFieldBehavior()
}

// IgnoreUnknownFields is the default unknown field behavior: unknown JSON object members are
// ignored.
type IgnoreUnknownFields struct{}

func (IgnoreUnknownFields) unknownFieldBehavior() {}

// WarnUnknownFields logs a warning when unknown JSON object members are found, then unmarshals
// while ignoring them.
// Will unmarshall twice when unknown fields are found.
type WarnUnknownFields struct{}

func (WarnUnknownFields) unknownFieldBehavior() {}

// ErrorUnknownFields will returns an error when unknown JSON object members are found.
type ErrorUnknownFields struct{}

func (ErrorUnknownFields) unknownFieldBehavior() {}

// FromJSON unmarshals JSON data to the given type.
// Unknown JSON object members are ignored by default.
func FromJSON[T any](data []byte) (T, error) {
	return FromJSONOpt[T, IgnoreUnknownFields](data)
}

// FromJSONOpt unmarshals JSON data to the given type, applying the behavior specified by U to
// unknown JSON object members.
func FromJSONOpt[T any, U UnknownFieldBehavior](data []byte) (T, error) {
	var res T
	var err error

	switch any(*new(U)).(type) {
	case IgnoreUnknownFields:
		err = unmarshalIgnoreUnknownFields(data, &res)
	case WarnUnknownFields:
		err = unmarshalWarnUnknownFields(data, &res)
	case ErrorUnknownFields:
		err = unmarshalErrorUnknownFields(data, &res)
	default:
		return fun.Zero[T](), errors.New("unsupported unknown-field behavior")
	}
	if err != nil {
		return fun.Zero[T](), err
	}

	return res, nil
}

func unmarshalIgnoreUnknownFields[T any](data []byte, res *T) error {
	return jsonv2.Unmarshal(data, res)
}

func unmarshalWarnUnknownFields[T any](data []byte, res *T) error {
	err := jsonv2.Unmarshal(data, res, jsonv2.RejectUnknownMembers(true))
	if err == nil {
		return nil
	}

	var semanticErr *jsonv2.SemanticError
	if !errors.As(err, &semanticErr) || !errors.Is(semanticErr, jsonv2.ErrUnknownName) {
		return err
	}

	log.Printf("%v", err)
	return jsonv2.Unmarshal(data, res)
}

func unmarshalErrorUnknownFields[T any](data []byte, res *T) error {
	return jsonv2.Unmarshal(data, res, jsonv2.RejectUnknownMembers(true))
}
