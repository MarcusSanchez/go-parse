package parse

import (
	"fmt"
	"regexp"
	"strings"
)

// formatErrTypeMismatch makes the error message for ErrTypeMismatch more generic and ready for client-side consumption.
func formatErrTypeMismatch(input string) string {
	re := regexp.MustCompile(`cannot unmarshal (\w+) into Go struct field (?:[^.\s]+)?\.(.+) of type (.*)$`)
	matches := re.FindStringSubmatch(input)

	if len(matches) == 4 {
		received := matches[1]
		tag := strings.ToLower(matches[2][0:1] + matches[2][1:]) // first letter to lowercase
		expected := matches[3]

		if expected == "map" {
			return fmt.Sprintf(
				"cannot parse '%s' into <%s>, expected 'object'.",
				received, tag,
			)
		}

		for i := 0; i < len(expected); i++ {
			if expected[i] == '.' {
				return fmt.Sprintf(
					"cannot parse '%s' into <%s>, expected 'object'.",
					received, tag,
				)
			}
		}

		return fmt.Sprintf(
			"cannot parse '%s' into <%s>, expected '%s'.",
			received, tag, expected,
		)
	}

	return input
}

// create makes a new instance of the provided type.
func create[T any]() T {
	var instance T
	return instance
}
