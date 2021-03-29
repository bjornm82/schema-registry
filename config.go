package schemaregistry

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type CompatibilityLevel int

const (
	Backward CompatibilityLevel = iota
	BackwardTransitive
	Forward
	ForwardTransitive
	Full
	FullTransitive
	None
)

const (
	configPath          = "/config/%s"
	jsonUnmarhalMessage = "unable to unmarshal json"
)

// Not needed so far
func (d CompatibilityLevel) String() string {
	switch d {
	case Backward:
		return "BACKWARD"
	case BackwardTransitive:
		return "BACKWARD_TRANSITIVE"
	case Forward:
		return "FORWARD"
	case ForwardTransitive:
		return "FORWARD_TRANSITIVE"
	case Full:
		return "FULL"
	case FullTransitive:
		return "FULL_TRANSITIVE"
	case None:
		return "NONE"
	default:
		return ""
	}
}

// Config describes a subject or globa schema-registry configuration
type Config struct {
	// CompatibilityLevel mode of subject or global
	Compatibility      string `json:"compatibility,omitempty"`
	CompatibilityLevel string `json:"compatibilityLevel,omitempty"`
}

// GetConfig returns the configuration (Config type) for global Schema-Registry or a specific
// subject. When Config returned has "compatibilityLevel" empty, it's using global settings.
func (c *Client) GetConfig(subject string) (Config, error) {
	return c.getConfigSubject(subject)
}

// SetConfigLevel according to the predefined compatibility levels
func (c *Client) SetConfigLevel(cl CompatibilityLevel, subject string) (Config, error) {
	var config = Config{}

	path := fmt.Sprintf(configPath, subject)
	b, err := json.Marshal(Config{
		Compatibility: cl.String(),
	})
	if err != nil {
		return config, errors.Wrap(err, jsonUnmarhalMessage)
	}

	resp, respErr := c.do(http.MethodPut, path, contentTypeSchemaJSON, b)

	return c.handle(resp, respErr)
}

func (c *Client) SetConfigLevelFull(subject string) (Config, error) {
	return c.SetConfigLevel(Full, subject)
}

// getConfigSubject returns the Config of global or for a given subject. It handles 404 error in a
// different way, since not-found for a subject configuration means it's using global.
func (c *Client) getConfigSubject(subject string) (Config, error) {
	path := fmt.Sprintf(configPath, subject)
	resp, respErr := c.do(http.MethodGet, path, "", nil)

	return c.handle(resp, respErr)
}

func (c *Client) handle(resp *http.Response, respErr error) (Config, error) {
	var err error
	var config = Config{}

	if respErr != nil && respErr.(ResourceError).ErrorCode != http.StatusNotFound {
		return config, respErr
	}
	if resp != nil {
		err = c.readJSON(resp, &config)
	}

	return config, err
}
