package schemaregistry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompatibilityLevelString_Passed(t *testing.T) {

	tests := []struct {
		input  CompatibilityLevel
		output string
	}{
		{input: Backward, output: "BACKWARD"},
		{input: BackwardTransitive, output: "BACKWARD_TRANSITIVE"},
		{input: Forward, output: "FORWARD"},
		{input: ForwardTransitive, output: "FORWARD_TRANSITIVE"},
		{input: Full, output: "FULL"},
		{input: FullTransitive, output: "FULL_TRANSITIVE"},
		{input: None, output: "NONE"},
		{input: 7, output: ""},
		{input: -1, output: ""},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.output, tc.input.String())
	}

	d := FullTransitive
	assert.Equal(t, "FULL_TRANSITIVE", d.String())
}
