package schemaregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	DefaultHost   = "localhost"
	DefaultPort   = 8081
	DefaultUseSSL = false
)

func TestSetConfig(t *testing.T) {
	cl, err := NewClient(DefaultHost, DefaultPort, DefaultUseSSL)
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
	cl, err := NewClient(DefaultHost, DefaultPort, DefaultUseSSL)
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

func TestSetConfig_PutToBackward(t *testing.T) {
	cl, err := NewClient(DefaultHost, DefaultPort, DefaultUseSSL)
	if err != nil {
		t.Error(t, err)
	}
	c, err := cl.SetConfigLevel(Backward, "")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "BACKWARD", c.Compatibility)
	assert.Equal(t, "", c.CompatibilityLevel)
}

func TestGetConfig_CheckIfBackward(t *testing.T) {
	cl, err := NewClient(DefaultHost, DefaultPort, DefaultUseSSL)
	if err != nil {
		t.Error(t, err)
	}
	c, err := cl.GetConfig("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "", c.Compatibility)
	assert.Equal(t, "BACKWARD", c.CompatibilityLevel)
}
