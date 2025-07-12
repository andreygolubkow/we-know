package walker

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDefaultPathBuilder_BuildPath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		nodeName string
		expected string
	}{
		{
			name:     "Empty base path",
			basePath: "",
			nodeName: "file.txt",
			expected: "file.txt",
		},
		{
			name:     "With base path",
			basePath: "base",
			nodeName: "file.txt",
			expected: "base/file.txt",
		},
		{
			name:     "With nested base path",
			basePath: "base/dir",
			nodeName: "file.txt",
			expected: "base/dir/file.txt",
		},
	}

	builder := NewPathBuilder()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildPath(tt.basePath, tt.nodeName)
			assert.Equal(t, tt.expected, result, "BuildPath(%q, %q) should return %q", tt.basePath, tt.nodeName, tt.expected)
		})
	}
}
