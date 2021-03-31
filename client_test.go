package schemaregistry

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const testHost = "testhost"
const testPort = 0
const testUseSSL = false

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
	baseURL, err := formatBaseURL(testHost, testPort, testUseSSL)
	if err != nil {
		t.Error(t, err)
	}
	return &Client{baseURL, dummyHTTPHandler(t, method, path, http.StatusOK, reqBody, respBody)}
}

func httpError(t *testing.T, status, errCode int, errMsg string) *Client {
	baseURL, err := formatBaseURL(testHost, testPort, testUseSSL)
	if err != nil {
		t.Error(t, err)
	}
	return &Client{baseURL, dummyHTTPHandler(t, "", "", status, nil, ResourceError{ErrorCode: errCode, Message: errMsg})}
}

type TestStruct struct {
	in  in
	out out
}
type in struct {
	host   string
	port   int
	useSSL bool
}
type out struct {
	exp string
	err error
}

var formatBaseURLTest = []TestStruct{
	{in: in{host: "localhost", port: 8081, useSSL: false},
		out: out{"http://localhost:8081", nil},
	},
	{in: in{host: "http://localhost", port: 8081, useSSL: false},
		out: out{"", errors.New(errHostNotContainForwardSlash)},
	},
	{in: in{host: "", port: 8081, useSSL: false},
		out: out{"", errors.New(errHostEmpty)},
	},
	{in: in{host: "localhost", port: 80, useSSL: false},
		out: out{"http://localhost", nil},
	},
	{in: in{host: "localhost", port: 80, useSSL: true},
		out: out{"https://localhost", nil},
	},
	{in: in{host: "other.usr", port: 8081, useSSL: false},
		out: out{"http://other.usr:8081", nil},
	},
	{in: in{host: "localhost", port: 443, useSSL: false},
		out: out{"https://localhost:443", nil},
	},
	{in: in{host: "localhost", port: 443, useSSL: true},
		out: out{"https://localhost:443", nil},
	},
	{in: in{host: "www.google.com/hello", port: 443, useSSL: true},
		out: out{"", errors.New(errHostNotContainForwardSlash)},
	},
	{in: in{host: "www.google", port: 0, useSSL: true},
		out: out{"https://www.google:443", nil},
	},
	{in: in{host: "localhost?query=123", port: 0, useSSL: true},
		out: out{"", errors.New(errHostNotContainQuestionmark)},
	},
}

func TestNewClient(t *testing.T) {
	cl, err := NewClient("localhost", 1234, false)
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "http://localhost:1234", cl.baseURL)
}

func TestNewClientWithUsingClient(t *testing.T) {
	emptyCl := http.Client{
		Timeout: time.Hour,
	}
	cl, err := NewClient("localhost", 1234, false, UsingClient(&emptyCl))
	if err != nil {
		t.Error(t, err)
	}
	custCl := cl.client

	c := custCl.(*http.Client)

	assert.Equal(t, float64(1), c.Timeout.Hours())
	assert.Equal(t, "http://localhost:1234", cl.baseURL)
}

func TestUsingClient(t *testing.T) {
	cl := http.Client{}
	UsingClient(&cl)
	assert.Equal(t, http.Client{}, cl)
}

func TestFormatBaseURL(t *testing.T) {
	for k, tt := range formatBaseURLTest {
		t.Run(fmt.Sprintf("host %s, with test ID %d", tt.in.host, k), func(t *testing.T) {
			act, err := formatBaseURL(tt.in.host, tt.in.port, tt.in.useSSL)
			assert.Equal(t, act, tt.out.exp)
			assert.Equal(t, err, tt.out.err)
		})
	}
}

func TestNewClient_FailedIncorrectHostEmpty(t *testing.T) {
	_, err := NewClient("", 1324, false)
	assert.Error(t, err)
}

func TestNewClient_FailedIncorrectHostHasSlash(t *testing.T) {
	_, err := NewClient("host://asdlkfj", 443, false)
	assert.Error(t, err)
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
	// s := `{"x":"y"}`
	// ss := schemaOnlyJSON{s}
	// sIn := Schema{s, "mysubject", 4, 7}
	// c := httpSuccess(t, http.MethodPost, "/subjects/mysubject", ss, sIn)
	// i := 2
	// ok, err := c.IsSchemaCompatible("mysubject", s, i)
	// if err != nil {
	// 	t.Error(err)
	// }
	// if !ok {
	// 	t.Error(err)
	// }
	// assert.True(t, ok)
}
