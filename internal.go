package parse

import (
	"fmt"
	"reflect"
	"strings"
)

// skeleton generates a map from a struct with keys as JSON tags and values as nil for all fields in a given struct type.
// It supports nested structures and optional fields marked with the "optional" tag.
func skeleton[T any](nested ...reflect.Value) map[string]any {

	// Use the provided nested or create a new instance of the struct
	t := reflect.TypeOf(create[T]())
	if len(nested) > 0 {
		t = nested[0].Type()
	}

	data := make(map[string]any, t.NumField())

outer:
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Skip fields marked as "optional" in the JSON tag
		args := strings.Split(tag, ",")
		for _, arg := range args {
			if arg == "optional" || arg == "-" || arg == "" {
				continue outer
			}
		}
		tag = args[0]

		// If the field is a nested structure, recursively call skeleton
		if field.Type.Kind() == reflect.Struct {
			value := reflect.New(field.Type).Elem()
			data[tag] = skeleton[T](value)
			continue
		}

		// Initialize the value as nil for non-struct fields
		data[tag] = nil
	}

	return data
}

// merge merges two maps, taking non-nil values from the second map and nil values from the first map.
// It recursively merges nested maps.
func merge(skelly, parsed map[string]any) map[string]any {
	result := make(map[string]any, len(skelly))

	for tag, value := range skelly {
		// If the tag is present in parsed and the value is not nil, use the value from parsed
		if parsedV, ok := parsed[tag]; ok && parsedV != nil {
			switch v := parsedV.(type) {
			case map[string]any:
				if value == nil {
					// likely a nested object with no required fields
					// i.e. a map[string]any or map[string]<STRUCT>
					value = make(map[string]any)
				}
				// If the value is a nested map, recursively merge
				result[tag] = merge(value.(map[string]any), v)
			default:
				// Otherwise, use the value from parsed
				result[tag] = parsedV
			}
		} else {
			// If the tag is not present in parsed or the value is nil, use the nil from skelly
			if !ok {
				// If the value was a nested map, set the entire nested map to nil
				if _, isNestedMap := value.(map[string]any); isNestedMap {
					result[tag] = nil
					continue
				}
			}
			result[tag] = value
		}
	}

	return result
}

// check recursively checks for required fields in a merged map.
// It returns an ErrMissingField if any required field is missing.
func check(merged map[string]any, parents ...string) error {

	// Use parent if provided
	parent := ""
	if len(parents) > 0 {
		parent = parents[0]
	}

	for tag, value := range merged {
		if parent != "" {
			tag = strings.Join([]string{parent, tag}, ".")
		}

		switch v := value.(type) {
		case map[string]any:
			// If the value is a nested map, recursively validate
			err := check(v, tag)
			if err != nil {
				return err // ErrMissingField
			}
		case nil:
			// If the value is null return a missing field error
			return fmt.Errorf("%w: <%s>", ErrMissingField, tag)
		}
	}

	return nil
}
