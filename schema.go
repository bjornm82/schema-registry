package schemaregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Schema describes a schema, look `GetSchema` for more.
type Schema struct {
	// Schema is the Avro schema string.
	Schema string `json:"schema"`
	// Subject where the schema is registered for.
	Subject string `json:"subject"`
	// Version of the returned schema.
	Version int `json:"version"`
	ID      int `json:"id,omitempty"`
}

// RegisterNewSchema registers a schema.
// The returned identifier should be used to retrieve
// this schema from the schemas resource and is different from
// the schema’s version which is associated with that name.
func (c *Client) RegisterNewSchema(subject string, avroSchema string) (int, error) {
	if subject == "" {
		return 0, errRequired("subject")
	}
	if avroSchema == "" {
		return 0, errRequired("avroSchema")
	}

	schema := schemaOnlyJSON{
		Schema: avroSchema,
	}

	send, err := json.Marshal(schema)
	if err != nil {
		return 0, err
	}

	// # Register a new schema under a particular subject
	// POST /subjects/(string: subject)/versions

	path := fmt.Sprintf(subjectPath+"/versions", subject)
	resp, err := c.do(http.MethodPost, path, contentTypeSchemaJSON, send)
	if err != nil {
		return 0, err
	}

	var res idOnlyJSON
	err = c.readJSON(resp, &res)
	return res.ID, err
}

// JSONAvroSchema converts and returns the json form of the "avroSchema" as []byte.
func JSONAvroSchema(avroSchema string) (json.RawMessage, error) {
	var raw json.RawMessage
	err := json.Unmarshal(json.RawMessage(avroSchema), &raw)
	if err != nil {
		return nil, err
	}
	return raw, err
}

// GetSchemaByID returns the Auro schema string identified by the id.
// id (int) – the globally unique identifier of the schema.
func (c *Client) GetSchemaByID(subjectID int) (string, error) {
	// # Get the schema for a particular subject id
	// GET /schemas/ids/{int: id}
	path := fmt.Sprintf(schemaPath, subjectID)
	resp, err := c.do(http.MethodGet, path, "", nil)
	if err != nil {
		return "", err
	}

	var res schemaOnlyJSON
	if err = c.readJSON(resp, &res); err != nil {
		return "", err
	}

	return res.Schema, nil
}

// SchemaLatestVersion is the only one valid string for the "versionID", it's the "latest" version string and it's used on `GetLatestSchema`.
const SchemaLatestVersion = "latest"

func checkSchemaVersionID(versionID interface{}) error {
	if versionID == nil {
		return errRequired("versionID (string \"latest\" or int)")
	}

	if verStr, ok := versionID.(string); ok {
		if verStr != SchemaLatestVersion {
			return fmt.Errorf("client: %v string is not a valid value for the versionID input parameter [versionID == \"latest\"]", versionID)
		}
	}

	if verInt, ok := versionID.(int); ok {
		if verInt <= 0 || verInt > 2^31-1 { // it's the max of int32, math.MaxInt32 already but do that check.
			return fmt.Errorf("client: %v integer is not a valid value for the versionID input parameter [ versionID > 0 && versionID <= 2^31-1]", versionID)
		}
	}

	return nil
}

// subject (string) – Name of the subject
// version (versionId [string "latest" or 1,2^31-1]) – Version of the schema to be returned.
// Valid values for versionId are between [1,2^31-1] or the string “latest”.
// The string “latest” refers to the last registered schema under the specified subject.
// Note that there may be a new latest schema that gets registered right after this request is served.
//
// It's not safe to use just an interface to the high-level API, therefore we split this method
// to two, one which will retrieve the latest versioned schema and the other which will accept
// the version as integer and it will retrieve by a specific version.
//
// See `GetLatestSchema` and `GetSchemaAtVersion` instead.
func (c *Client) getSubjectSchemaAtVersion(subject string, versionID interface{}) (s Schema, err error) {
	if subject == "" {
		err = errRequired("subject")
		return
	}

	if err = checkSchemaVersionID(versionID); err != nil {
		return
	}

	// # Get the schema at a particular version
	// GET /subjects/(string: subject)/versions/(versionId: "latest" | int)
	path := fmt.Sprintf(subjectPath+"/versions/%v", subject, versionID)
	resp, respErr := c.do(http.MethodGet, path, "", nil)
	if respErr != nil {
		err = respErr
		return
	}

	err = c.readJSON(resp, &s)
	return
}

// GetSchemaBySubject returns the schema for a particular subject and version.
func (c *Client) GetSchemaBySubject(subject string, versionID int) (Schema, error) {
	return c.getSubjectSchemaAtVersion(subject, versionID)
}

// GetLatestSchema returns the latest version of a schema.
// See `GetSchemaAtVersion` to retrieve a subject schema by a specific version.
func (c *Client) GetLatestSchema(subject string) (Schema, error) {
	return c.getSubjectSchemaAtVersion(subject, SchemaLatestVersion)
}

// subject (string) – Name of the subject
// version (versionId [string "latest" or 1,2^31-1]) – Version of the schema to be returned.
// Valid values for versionId are between [1,2^31-1] or the string “latest”.
// The string “latest” refers to the last registered schema under the specified subject.
// Note that there may be a new latest schema that gets registered right after this request is served.
//
// It's not safe to use just an interface to the high-level API, therefore we split this method
// to two, one which will retrieve the latest versioned schema and the other which will accept
// the version as integer and it will retrieve by a specific version.
//
// See `IsSchemaCompatible` and `IsLatestSchemaCompatible` instead.
func (c *Client) isSchemaCompatibleAtVersion(subject string, avroSchema string, versionID interface{}) (combatible bool, err error) {
	if subject == "" {
		err = errRequired("subject")
		return
	}
	if avroSchema == "" {
		err = errRequired("avroSchema")
		return
	}

	if err = checkSchemaVersionID(versionID); err != nil {
		return
	}

	schema := schemaOnlyJSON{
		Schema: avroSchema,
	}

	send, err := json.Marshal(schema)
	if err != nil {
		return
	}

	// # Test input schema against a particular version of a subject’s schema for compatibility
	// POST /compatibility/subjects/(string: subject)/versions/(versionId: "latest" | int)
	path := fmt.Sprintf("compatibility/"+subjectPath+"/versions/%v", subject, versionID)
	resp, err := c.do(http.MethodPost, path, contentTypeSchemaJSON, send)
	if err != nil {
		return
	}

	var res isCompatibleJSON
	err = c.readJSON(resp, &res)

	return res.IsCompatible, err
}

// IsRegistered tells if the given "schema" is registered for this "subject".
func (c *Client) IsRegistered(subject, schema string) (bool, Schema, error) {
	var fs Schema

	sc := schemaOnlyJSON{schema}
	send, err := json.Marshal(sc)
	if err != nil {
		return false, fs, err
	}

	path := fmt.Sprintf(subjectPath, subject)
	resp, err := c.do(http.MethodPost, path, "", send)
	if err != nil {
		// schema not found?
		if IsSchemaNotFound(err) {
			return false, fs, nil
		}
		// error?
		return false, fs, err
	}

	if err = c.readJSON(resp, &fs); err != nil {
		return true, fs, err // found but error when unmarshal.
	}

	// so we have a schema.
	return true, fs, nil
}

// IsSchemaCompatible tests compatibility with a specific version of a subject's schema.
func (c *Client) IsSchemaCompatible(subject string, avroSchema string, versionID int) (bool, error) {
	return c.isSchemaCompatibleAtVersion(subject, avroSchema, versionID)
}

// IsLatestSchemaCompatible tests compatibility with the latest version of a subject's schema.
func (c *Client) IsLatestSchemaCompatible(subject string, avroSchema string) (bool, error) {
	return c.isSchemaCompatibleAtVersion(subject, avroSchema, SchemaLatestVersion)
}
