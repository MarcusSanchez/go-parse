package parse

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrTypeConstraint = errors.New("type constraint")
	ErrInvalidJSON    = errors.New("invalid JSON format")
	ErrTypeMismatch   = errors.New("type mismatch")
	ErrMissingField   = errors.New("missing required field")
)

// JSON parses the input JSON data into a provided struct and ensures that all required fields are present.
// The struct should be provided as a non-pointer type.
func JSON[T any](data []byte, externals ...*T) (output *T, err error) {
	// Check if the provided type is a pointer, which is not allowed
	kind := reflect.TypeOf(create[T]()).Kind()
	switch {
	case kind == reflect.Ptr:
		return output, fmt.Errorf("%w: generic must be a non-pointer", ErrTypeConstraint)
	case kind != reflect.Struct:
		return output, fmt.Errorf(
			"%w: generic must be a struct not '%s'", ErrTypeConstraint, kind.String(),
		)
	}

	// Unmarshal the input JSON data into a map
	parsed := make(map[string]any)
	if err = json.Unmarshal(data, &parsed); err != nil {
		return output, fmt.Errorf("%w: %s", ErrInvalidJSON, err.Error())
	}

	// if an external struct was provided, use that instead of creating a new one
	if len(externals) > 0 {
		output = externals[0]
	}
	if err = json.Unmarshal(data, &output); err != nil {
		return output, fmt.Errorf("%w: %s", ErrTypeMismatch, formatErrTypeMismatch(err.Error()))
	}

	// Generate a map with JSON tags for all fields in the provided struct with nil values
	skelly := skeleton[T]()
	// overlay passed values with skeleton map
	merged := merge(skelly, parsed)

	// Check if all required fields were present
	if err = check(merged); err != nil {
		return output, err // ErrMissingField
	}

	return output, nil
}
