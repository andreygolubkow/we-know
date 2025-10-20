package walker

import (
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

// TreeCallback is a function type that is called for each node in the file tree
// Use an alias to ensure compatibility with interfaces expecting the raw function type.
type TreeCallback = func(node *hs.FileTreeNode, fullPath string) error

// FileTreeWalker defines the interface for walking a file tree
type FileTreeWalker interface {
	// Walk traverses the file tree and calls the callback function for each node
	Walk(root *hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) error
}

// DefaultFileTreeWalker is the default implementation of FileTreeWalker
type DefaultFileTreeWalker struct {
	pathBuilder PathBuilder
}

// NewFileTreeWalker creates a new DefaultFileTreeWalker with a default PathBuilder
func NewFileTreeWalker() *DefaultFileTreeWalker {
	return &DefaultFileTreeWalker{
		pathBuilder: NewPathBuilder(),
	}
}

// Walk traverses the file tree and calls the callback function for each node
func (w *DefaultFileTreeWalker) Walk(root *hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) error {
	if root == nil {
		return nil
	}

	return w.walkNode(*root, callback, pathBase, ignoredFiles)
}

// walkNode is a helper function for Walk that processes a single node and its children
func (w *DefaultFileTreeWalker) walkNode(node hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) error {
	nextNodes := node.GetNext(ignoredFiles)
	nodePath := w.pathBuilder.BuildPath(pathBase, node.GetName())

	for _, childNode := range nextNodes {
		childPath := w.pathBuilder.BuildPath(nodePath, (*childNode).GetName())

		if err := callback(childNode, childPath); err != nil {
			return err
		}

		if err := w.walkNode(*childNode, callback, nodePath, ignoredFiles); err != nil {
			return err
		}
	}

	return nil
}

// For backward compatibility
// Walk traverses the file tree and calls the callback function for each node
func Walk(root *hs.FileTreeNode, callback func(node *hs.FileTreeNode, fullPath string), pathBase string, ignoredFiles *[]string) {
	walker := NewFileTreeWalker()
	_ = walker.Walk(root, func(node *hs.FileTreeNode, fullPath string) error {
		callback(node, fullPath)
		return nil
	}, pathBase, ignoredFiles)
}
