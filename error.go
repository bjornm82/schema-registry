package schemaregistry

import (
	"fmt"
	"net/url"
)

// These numbers are used by the schema registry to communicate errors.
const (
	subjectNotFoundCode = 40401
	schemaNotFoundCode  = 40403

	errorMessage    = "client: (%s: %s) failed with error code %d%s"
	requiredMessage = "client: %s is required"
)

var errRequired = func(field string) error {
	return fmt.Errorf(requiredMessage, field)
}

// ResourceError is being fired from all API calls when an error code is received.
type ResourceError struct {
	ErrorCode int    `json:"error_code"`
	Method    string `json:"method,omitempty"`
	URI       string `json:"uri,omitempty"`
	Message   string `json:"message,omitempty"`
}

func (err ResourceError) Error() string {
	return fmt.Sprintf(errorMessage,
		err.Method, err.URI, err.ErrorCode, err.Message)
}

func newResourceError(errCode int, uri, method, body string) ResourceError {
	unescapedURI, _ := url.QueryUnescape(uri)

	return ResourceError{
		ErrorCode: errCode,
		URI:       unescapedURI,
		Method:    method,
		Message:   body,
	}
}

// IsSubjectNotFound checks the returned error to see if it is kind of a subject not found  error code.
func IsSubjectNotFound(err error) bool {
	return checkNotFound(err, subjectNotFoundCode)
}

// IsSchemaNotFound checks the returned error to see if it is kind of a schema not found error code.
func IsSchemaNotFound(err error) bool {
	return checkNotFound(err, schemaNotFoundCode)
}

func checkNotFound(err error, code int) bool {
	if err == nil {
		return false
	}

	if resErr, ok := err.(ResourceError); ok {
		return resErr.ErrorCode == code
	}

	return false
}
