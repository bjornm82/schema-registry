package schemaregistry

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testHost = "testhost:1337"
const testURL = "http://" + testHost

type D func(req *http.Request) (*http.Response, error)

func (d D) Do(req *http.Request) (*http.Response, error) {
	return d(req)
}

// verifies the http.Request, creates an http.Response
func dummyHTTPHandler(t *testing.T, method, path string, status int, reqBody, respBody interface{}) D {
	d := D(func(req *http.Request) (*http.Response, error) {
		if method != "" && req.Method != method {
			t.Errorf("method is wrong, expected `%s`, got `%s`", method, req.Method)
		}
		if req.URL.Host != testHost {
			t.Errorf("expected host `%s`, got `%s`", testHost, req.URL.Host)
		}
		if path != "" && req.URL.Path != path {
			t.Errorf("path is wrong, expected `%s`, got `%s`", path, req.URL.Path)
		}
		if reqBody != nil {
			expbs, err := json.Marshal(reqBody)
			if err != nil {
				t.Error(t, err)
			}
			bs, err := ioutil.ReadAll(req.Body)
			if err != nil {
				t.Error(t, err)
			}
			assert.Equal(t, strings.Trim(string(expbs), "\r\n"), strings.Trim(string(bs), "\r\n"))
		}
		var resp http.Response
		resp.Header = http.Header{contentTypeHeaderKey: []string{contentTypeJSON}}
		resp.StatusCode = status
		if respBody != nil {
			bs, err := json.Marshal(respBody)
			if err != nil {
				t.Error(err)
			}
			resp.Body = ioutil.NopCloser(bytes.NewReader(bs))
		}
		return &resp, nil
	})
	return d
}

func httpSuccess(t *testing.T, method, path string, reqBody, respBody interface{}) *Client {
	return &Client{testURL, dummyHTTPHandler(t, method, path, http.StatusOK, reqBody, respBody)}
}

func httpError(t *testing.T, status, errCode int, errMsg string) *Client {
	return &Client{testURL, dummyHTTPHandler(t, "", "", status, nil, ResourceError{ErrorCode: errCode, Message: errMsg})}
}

func TestSubjects(t *testing.T) {
	subsIn := []string{"rollulus", "hello-subject"}
	c := httpSuccess(t, http.MethodGet, "/subjects", nil, subsIn)
	subs, err := c.Subjects()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, subsIn, subs)
}

func TestVersions(t *testing.T) {
	versIn := []int{1, 2, 3}
	subjectName := "mysubject"
	c := httpSuccess(t, http.MethodGet, "/subjects/"+subjectName+"/versions", nil, versIn)
	vers, err := c.Versions(subjectName)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, versIn, vers)
}

func TestIsRegistered_yes(t *testing.T) {
	s := `{"x":"y"}`
	ss := schemaOnlyJSON{s}
	sIn := Schema{s, "mysubject", 4, 7}
	c := httpSuccess(t, http.MethodPost, "/subjects/mysubject", ss, sIn)
	isreg, sOut, err := c.IsRegistered("mysubject", s)
	if err != nil {
		t.Error(err)
	}
	if !isreg {
		t.Error(err)
	}
	assert.Equal(t, sIn, sOut)
}

func TestIsRegistered_not(t *testing.T) {
	c := httpError(t, http.StatusNotFound, schemaNotFoundCode, "too bad")
	isreg, _, err := c.IsRegistered("mysubject", "{}")
	if err != nil {
		t.Fatal(err)
	}
	assert.False(t, isreg)
}

func TestIsSchemaCompatible(t *testing.T) {
	s := `{"x":"y"}`
	ss := schemaOnlyJSON{s}
	sIn := Schema{s, "mysubject", 4, 7}
	c := httpSuccess(t, http.MethodPost, "/subjects/mysubject", ss, sIn)
	i := 2
	ok, err := c.IsSchemaCompatible("mysubject", s, i)
	if err != nil {
		t.Error(err)
	}
	if !ok {
		t.Error(err)
	}
	assert.True(t, ok)
}
