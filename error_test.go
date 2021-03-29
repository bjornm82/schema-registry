package schemaregistry

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrRequiredPassed(t *testing.T) {
	value := "subject"

	err := errRequired(value)
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf(requiredMessage, value), err)
}
func TestResourceError(t *testing.T) {
	code := subjectNotFoundCode
	uri := "http://example.com"
	method := http.MethodGet
	body := "body"

	err := newResourceError(code, uri, method, body)

	assert.Error(t, err)
	assert.NotNil(t, err)
	assert.Equal(t, fmt.Sprintf(errorMessage, method, uri, code, body), err.Error())
}

func TestCheckNotFound_FailedNotResourceError(t *testing.T) {
	err := errors.New("this is incorrect")
	ok := checkNotFound(err, schemaNotFoundCode)
	assert.False(t, ok)
}

func TestCheckNotFound_Passed(t *testing.T) {
	err := newResourceError(subjectNotFoundCode, "http://example.com", http.MethodGet, "body")
	ok := checkNotFound(err, subjectNotFoundCode)
	assert.True(t, ok)
}

func TestCheckNotFound_ReturnFalseWrongCode(t *testing.T) {
	err := newResourceError(http.StatusBadGateway, "http://example.com", http.MethodGet, "body")
	ok := checkNotFound(err, schemaNotFoundCode)
	assert.False(t, ok)
}

func TestCheckNotFound_ReturnFalseValueNil(t *testing.T) {
	ok := checkNotFound(nil, schemaNotFoundCode)
	assert.False(t, ok)
}

func TestIsSubjectNotFound_Passed(t *testing.T) {
	err := newResourceError(subjectNotFoundCode, "http://example.com", http.MethodGet, "body")
	ok := IsSubjectNotFound(err)
	assert.True(t, ok)
}

func TestIsSubjectNotFound_FailedWrongCode(t *testing.T) {
	err := newResourceError(schemaNotFoundCode, "http://example.com", http.MethodGet, "body")
	ok := IsSubjectNotFound(err)
	assert.False(t, ok)
}

func TestIsSchemaNotFound_Passed(t *testing.T) {
	err := newResourceError(schemaNotFoundCode, "http://example.com", http.MethodGet, "body")
	ok := IsSchemaNotFound(err)
	assert.True(t, ok)
}

func TestIsSchemaNotFound_FailedWrongCode(t *testing.T) {
	err := newResourceError(subjectNotFoundCode, "http://example.com", http.MethodGet, "body")
	ok := IsSchemaNotFound(err)
	assert.False(t, ok)
}
