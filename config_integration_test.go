package schemaregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DefaultClientURI = "http://localhost:8081"
)

func TestSetConfig(t *testing.T) {
	cl, err := NewClient(DefaultClientURI)
	if err != nil {
		t.Error(t, err)
	}
	c, err := cl.SetConfigLevelFull("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "FULL", c.Compatibility)
	assert.Equal(t, "", c.CompatibilityLevel)
}

func TestGetConfig(t *testing.T) {
	cl, err := NewClient(DefaultClientURI)
	if err != nil {
		t.Error(t, err)
	}
	c, err := cl.GetConfig("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "", c.Compatibility)
	assert.Equal(t, "FULL", c.CompatibilityLevel)
}
