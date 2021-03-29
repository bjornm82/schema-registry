package schemaregistry

import (
	"fmt"
	"net/http"
)

const (
	subjectsPath = "subjects"
	subjectPath  = subjectsPath + "/%s"
	schemaPath   = "schemas/ids/%d"
)

// Subjects returns a list of the available subjects(schemas).
// https://docs.confluent.io/current/schema-registry/docs/api.html#subjects
func (c *Client) Subjects() (subjects []string, err error) {
	// # List all available subjects
	// GET /subjects
	resp, respErr := c.do(http.MethodGet, subjectsPath, "", nil)
	if respErr != nil {
		err = respErr
		return
	}

	err = c.readJSON(resp, &subjects)
	return
}

// Versions returns all schema version numbers registered for this subject.
func (c *Client) Versions(subject string) (versions []int, err error) {
	if subject == "" {
		err = errRequired("subject")
		return
	}

	// # List all versions of a particular subject
	// GET /subjects/(string: subject)/versions
	path := fmt.Sprintf(subjectPath, subject+"/versions")
	resp, respErr := c.do(http.MethodGet, path, "", nil)
	if respErr != nil {
		err = respErr
		return
	}

	err = c.readJSON(resp, &versions)
	return
}

// DeleteSubject deletes the specified subject and its associated compatibility level if registered.
// It is recommended to use this API only when a topic needs to be recycled or in development environment.
// Returns the versions of the schema deleted under this subject.
func (c *Client) DeleteSubject(subject string) (versions []int, err error) {
	if subject == "" {
		err = errRequired("subject")
		return
	}

	// DELETE /subjects/(string: subject)
	path := fmt.Sprintf(subjectPath, subject)
	resp, respErr := c.do(http.MethodDelete, path, "", nil)
	if respErr != nil {
		err = respErr
		return
	}

	err = c.readJSON(resp, &versions)
	return
}
