// +build integration

package schemaregistry

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var client *Client

func init() {
	addr := fmt.Sprintf("%s:%d", os.Getenv("DOCKER_IP"), 8086)

	cl, err := NewClient(addr)
	if err != nil {
		log.Fatalf("connection not able to be established: %s", err)
	}
	client = cl
}

func TestCreate_Passed(t *testing.T) {
	su, err := client.Subjects()
	assert.Equal(t, 0, len(su))
	assert.NoError(t, err)
}
