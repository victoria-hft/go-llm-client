package schema_compliance

import "fmt"

// ErrorKind identifies why Ensure could not return schema-compliant JSON.
type ErrorKind string

const (
	// ErrorKindInvalidJSON indicates the final output could not be parsed as JSON.
	ErrorKindInvalidJSON ErrorKind = "invalid_json"

	// ErrorKindInvalidSchema indicates the supplied schema could not be parsed or compiled.
	ErrorKindInvalidSchema ErrorKind = "invalid_schema"

	// ErrorKindSchemaViolation indicates the JSON did not match the supplied schema.
	ErrorKindSchemaViolation ErrorKind = "schema_violation"
)

// Error is returned for all schema compliance failures.
type Error struct {
	Kind ErrorKind
	Err  error
}

// Error returns a human-readable schema compliance error.
func (e *Error) Error() string {
	if e.Err == nil {
		return string(e.Kind)
	}
	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

// Unwrap returns the wrapped lower-level error, if any.
func (e *Error) Unwrap() error {
	return e.Err
}
