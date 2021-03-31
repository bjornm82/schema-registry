// Package schemaregistry provides a client for Confluent's Kafka Schema Registry REST API.
package schemaregistry

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	errHostEmpty                  = "host can not be empty"
	errHostNotContainForwardSlash = "host can not contain a /"
	errHostNotContainQuestionmark = "host can not contain a ?"
)

// DefaultURL is the address where a local schema registry listens by default.
const DefaultURL = "http://localhost:8081"

type (
	httpDoer interface {
		Do(req *http.Request) (resp *http.Response, err error)
	}
	// Client is the registry schema REST API client.
	Client struct {
		baseURL string

		// the client is created on the `NewClient` function, it can be customized via options.
		client httpDoer
	}

	// Option describes an optional runtime configurator that can be passed on `NewClient`.
	// Custom `Option` can be used as well, it's just a type of `func(*schemaregistry.Client)`.
	//
	// Look `UsingClient`.
	Option func(*Client)
)
type (
	schemaOnlyJSON struct {
		Schema string `json:"schema"`
	}

	idOnlyJSON struct {
		ID int `json:"id"`
	}

	isCompatibleJSON struct {
		IsCompatible bool `json:"is_compatible"`
	}
)

// UsingClient modifies the underline HTTP Client that schema registry is using for contact with the backend server.
func UsingClient(httpClient *http.Client) Option {
	return func(c *Client) {
		if httpClient == nil {
			return
		}

		transport := getTransportLayer(httpClient, 0)
		httpClient.Transport = transport

		c.client = httpClient
	}
}

func getTransportLayer(httpClient *http.Client, timeout time.Duration) (t http.RoundTripper) {
	if t := httpClient.Transport; t != nil {
		return t
	}

	httpTransport := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}

	// if timeout > 0 {
	// 	httpTransport.Dial = func(network string, addr string) (net.Conn, error) {
	// 		return net.DialTimeout(network, addr, timeout)
	// 	}
	// }

	return httpTransport
}

// formatBaseURL will try to make sure that the schema:host:port pattern is followed on the `baseURL` field.
func formatBaseURL(host string, port int, useSSL bool) (string, error) {
	if host == "" {
		return "", errors.New(errHostEmpty)
	}

	if strings.Contains(host, "/") {
		return "", errors.New(errHostNotContainForwardSlash)
	}

	if strings.Contains(host, "?") {
		return "", errors.New(errHostNotContainQuestionmark)
	}

	var scheme = "http"

	if useSSL {
		scheme = "https"
	}

	if port == 443 {
		scheme = "https"
	}

	if port == 0 && useSSL {
		scheme = "https"
		port = 443
	}

	if port == 0 && !useSSL {
		scheme = "http"
		return fmt.Sprintf("%s://%s", scheme, host), nil
	}

	if port == 80 && !useSSL {
		scheme = "http"
		return fmt.Sprintf("%s://%s", scheme, host), nil
	}

	if port == 80 && useSSL {
		scheme = "https"
		return fmt.Sprintf("%s://%s", scheme, host), nil
	}

	return fmt.Sprintf("%s://%s:%d", scheme, host, port), nil
}

// NewClient creates & returns a new Registry Schema Client
// based on the passed url and the options.
func NewClient(host string, port int, useSSL bool, options ...Option) (*Client, error) {
	baseURL, err := formatBaseURL(host, port, useSSL)
	if err != nil {
		return nil, err
	}

	if _, err := url.Parse(baseURL); err != nil {
		return nil, err
	}

	c := &Client{baseURL: baseURL}
	for _, opt := range options {
		opt(c)
	}

	if c.client == nil {
		httpClient := &http.Client{}
		UsingClient(httpClient)(c)
	}

	return c, nil
}

const (
	contentTypeHeaderKey = "Content-Type"
	contentTypeJSON      = "application/json"

	acceptHeaderKey          = "Accept"
	acceptEncodingHeaderKey  = "Accept-Encoding"
	contentEncodingHeaderKey = "Content-Encoding"
	gzipEncodingHeaderValue  = "gzip"
)

// isOK is called inside the `Client#do` and it closes the body reader if no accessible.
func isOK(resp *http.Response) bool {
	return !(resp.StatusCode < 200 || resp.StatusCode >= 300)
}

var noOpBuffer = new(bytes.Buffer)

func acquireBuffer(b []byte) *bytes.Buffer {
	if len(b) > 0 {
		return bytes.NewBuffer(b)
	}

	return noOpBuffer
}

const schemaAPIVersion = "v1"
const contentTypeSchemaJSON = "application/vnd.schemaregistry." + schemaAPIVersion + "+json"

func (c *Client) do(method, path, contentType string, send []byte) (*http.Response, error) {
	if path[0] == '/' {
		path = path[1:]
	}

	uri := c.baseURL + "/" + path

	req, err := http.NewRequest(method, uri, acquireBuffer(send))
	if err != nil {
		return nil, err
	}

	// set the content type if any.
	if contentType != "" {
		req.Header.Set(contentTypeHeaderKey, contentType)
	}

	// response accept gziped content.
	req.Header.Add(acceptEncodingHeaderKey, gzipEncodingHeaderValue)
	req.Header.Add(acceptHeaderKey, contentTypeSchemaJSON+", application/vnd.schemaregistry+json, application/json")

	// send the request and check the response for any connection & authorization errors here.
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if !isOK(resp) {
		defer resp.Body.Close()
		var errBody string
		respContentType := resp.Header.Get(contentTypeHeaderKey)

		if strings.Contains(respContentType, "text/html") {
			// if the body is html, then don't read it, it doesn't contain the raw info we need.
		} else if strings.Contains(respContentType, "json") {
			// if it's json try to read it as confluent's specific error json.
			var resErr ResourceError
			c.readJSON(resp, &resErr)
			return nil, resErr
		} else {
			// else give the whole body to the error context.
			b, err := c.readResponseBody(resp)
			if err != nil {
				errBody = " unable to read body: " + err.Error()
			} else {
				errBody = "\n" + string(b)
			}
		}

		return nil, newResourceError(resp.StatusCode, uri, method, errBody)
	}

	return resp, nil
}

type gzipReadCloser struct {
	respReader io.ReadCloser
	gzipReader io.ReadCloser
}

func (rc *gzipReadCloser) Close() error {
	if rc.gzipReader != nil {
		defer rc.gzipReader.Close()
	}

	return rc.respReader.Close()
}

func (rc *gzipReadCloser) Read(p []byte) (n int, err error) {
	if rc.gzipReader != nil {
		return rc.gzipReader.Read(p)
	}

	return rc.respReader.Read(p)
}

func (c *Client) acquireResponseBodyStream(resp *http.Response) (io.ReadCloser, error) {
	// check for gzip and read it, the right way.
	var (
		reader = resp.Body
		err    error
	)

	if encoding := resp.Header.Get(contentEncodingHeaderKey); encoding == gzipEncodingHeaderValue {
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("client: failed to read gzip compressed content, trace: %v", err)
		}
		// we wrap the gzipReader and the underline response reader
		// so a call of .Close() can close both of them with the correct order when finish reading, the caller decides.
		// Must close manually using a defer on the callers before the `readResponseBody` call,
		// note that the `readJSON` can decide correctly by itself.
		return &gzipReadCloser{
			respReader: resp.Body,
			gzipReader: reader,
		}, nil
	}

	// return the stream reader.
	return reader, err
}

func (c *Client) readResponseBody(resp *http.Response) ([]byte, error) {
	reader, err := c.acquireResponseBodyStream(resp)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if err = reader.Close(); err != nil {
		return nil, err
	}

	// return the body.
	return body, err
}

func (c *Client) readJSON(resp *http.Response, valuePtr interface{}) error {
	b, err := c.readResponseBody(resp)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, valuePtr)
}
