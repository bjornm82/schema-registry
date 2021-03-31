// +build integration

package schemaregistry

import (
	"log"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	DefaultHost   = "registry-test"
	DefaultPort   = 8081
	DefaultUseSSL = false
)

// # TODO make integration tests

// # Register a new version of a schema under the subject "Kafka-value"
// $ curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
//     --data '{"schema": "{\"type\": \"string\"}"}' \
//      http://localhost:8081/subjects/Kafka-value/versions
//   {"id":1}

// # List all subjects
// $ curl -X GET http://localhost:8081/subjects
//   ["Kafka-value","Kafka-key"]

// # List all schema versions registered under the subject "Kafka-value"
// $ curl -X GET http://localhost:8081/subjects/Kafka-value/versions
//   [1]

// # Fetch a schema by globally unique id 1
// $ curl -X GET http://localhost:8081/schemas/ids/1
//   {"schema":"\"string\""}

// # Fetch version 1 of the schema registered under subject "Kafka-value"
// $ curl -X GET http://localhost:8081/subjects/Kafka-value/versions/1
//   {"subject":"Kafka-value","version":1,"id":1,"schema":"\"string\""}

// # Fetch the most recently registered schema under subject "Kafka-value"
// $ curl -X GET http://localhost:8081/subjects/Kafka-value/versions/latest
//   {"subject":"Kafka-value","version":1,"id":1,"schema":"\"string\""}

// # Delete version 3 of the schema registered under subject "Kafka-value"
// $ curl -X DELETE http://localhost:8081/subjects/Kafka-value/versions/3
//   3

// # Delete all versions of the schema registered under subject "Kafka-value"
// $ curl -X DELETE http://localhost:8081/subjects/Kafka-value
//   [1, 2, 3, 4, 5]

// # Check whether a schema has been registered under subject "Kafka-key"
// $ curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
//     --data '{"schema": "{\"type\": \"string\"}"}' \
//     http://localhost:8081/subjects/Kafka-key
//   {"subject":"Kafka-key","version":1,"id":1,"schema":"\"string\""}

// # Test compatibility of a schema with the latest schema under subject "Kafka-value"
// $ curl -X POST -H "Content-Type: application/vnd.schemaregistry.v1+json" \
//     --data '{"schema": "{\"type\": \"string\"}"}' \
//     http://localhost:8081/compatibility/subjects/Kafka-value/versions/latest
//   {"is_compatible":true}

// # Get top level config
// $ curl -X GET http://localhost:8081/config
//   {"compatibilityLevel":"BACKWARD"}

// # Update compatibility requirements globally
// $ curl -X PUT -H "Content-Type: application/vnd.schemaregistry.v1+json" \
//     --data '{"compatibility": "NONE"}' \
//     http://localhost:8081/config
//   {"compatibility":"NONE"}

// # Update compatibility requirements under the subject "Kafka-value"
// $ curl -X PUT -H "Content-Type: application/vnd.schemaregistry.v1+json" \
//     --data '{"compatibility": "BACKWARD"}' \
//     http://localhost:8081/config/Kafka-value
//   {"compatibility":"BACKWARD"}

var client *Client

func init() {
	cl, err := NewClient(DefaultHost, DefaultPort, DefaultUseSSL)
	if err != nil {
		log.Fatalf("connection not able to be established: %s", err)
	}
	client = cl
}

func TestNewClient(t *testing.T) {
	cl, err := NewClient(DefaultHost, 1234, false)
	if err != nil {
		t.Error(t, err)
	}
	client := *cl

	assert.Equal(t, "http://"+DefaultHost+":1234", client.baseURL)
}

var subjectName = "schema-" + strconv.Itoa(int(time.Now().Unix()))

func TestCreateNewSchema(t *testing.T) {
	schema := `{
		"type": "record",
		"namespace": "com.example",
		"name": "FullName",
		"fields": [
		  { "name": "first", "type": "string" },
		  { "name": "last", "type": "string" }
		]
   }`

	id, err := client.RegisterNewSchema(subjectName, schema)

	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
	assert.NotEqual(t, -1, id)
}

func DeleteSchema(t *testing.T) {
	schema, err := client.GetLatestSchema(subjectName)
	if err != nil {
		t.Error(t, err)
	}

	log.Println(schema.ID)
	log.Println(schema.Subject)

	assert.NoError(t, err)
}

func TestCreate_Passed(t *testing.T) {
	su, err := client.Subjects()
	assert.Len(t, su, 1)
	assert.NoError(t, err)
}

func TestSetConfig(t *testing.T) {
	c, err := client.SetConfigLevelFull("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "FULL", c.Compatibility)
	assert.Equal(t, "", c.CompatibilityLevel)
}

func TestGetConfig(t *testing.T) {
	c, err := client.GetConfig("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "", c.Compatibility)
	assert.Equal(t, "FULL", c.CompatibilityLevel)
}

func TestSetConfig_PutToBackward(t *testing.T) {
	c, err := client.SetConfigLevel(Backward, "")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "BACKWARD", c.Compatibility)
	assert.Equal(t, "", c.CompatibilityLevel)
}

func TestGetConfig_CheckIfBackward(t *testing.T) {
	c, err := client.GetConfig("")
	if err != nil {
		t.Error(t, err)
	}
	assert.Equal(t, "", c.Compatibility)
	assert.Equal(t, "BACKWARD", c.CompatibilityLevel)
}
