package walker

import (
	"path/filepath"
)

// PathBuilder builds file paths
type PathBuilder interface {
	// BuildPath builds a path from a base path and a name
	BuildPath(basePath, name string) string
}

// DefaultPathBuilder is the default implementation of PathBuilder
type DefaultPathBuilder struct{}

// NewPathBuilder creates a new DefaultPathBuilder
func NewPathBuilder() *DefaultPathBuilder {
	return &DefaultPathBuilder{}
}

// BuildPath builds a path from a base path and a name
func (b *DefaultPathBuilder) BuildPath(basePath, name string) string {
	if basePath == "" {
		return name
	}
	return filepath.Join(basePath, name)
}
