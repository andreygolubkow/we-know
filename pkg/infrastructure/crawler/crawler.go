package crawler

import (
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

// UserMapper defines the interface for mapping user IDs to display names
// This is used to convert raw user identifiers to human-friendly names.
type UserMapper interface {
	// GetDisplayName returns the display name for the given user identifier
	GetDisplayName(userID string) string
}

// EditorStorage defines the interface for storing file editor information
// Implementations persist aggregated editors per file.
type EditorStorage interface {
	// SetFileEditors sets the editors for a file
	SetFileEditors(filePath string, editors *map[string]int, errorMsg string)
}

// CodeStorage defines the interface for retrieving historical code information
// Implementations typically pull data from VCS like Git.
type CodeStorage interface {
	// GetEditorsByFile returns the editors for a file
	GetEditorsByFile(filename string) (*map[string]int, string)
}

// FileTreeWalker defines the interface for walking a file tree.
// This local interface is intentionally decoupled from the walker package to avoid import cycles.
type FileTreeWalker interface {
	// Walk traverses the file tree and calls the callback function for each node
	Walk(root *hs.FileTreeNode, callback func(node *hs.FileTreeNode, fullPath string) error, pathBase string, ignoredFiles *[]string) error
}

// FileCrawler defines the interface for crawling a file tree and collecting editor information
// It relies on a FileTreeWalker to enumerate files and then analyses them
// using CodeStorage and stores results in EditorStorage.
type FileCrawler interface {
	// Crawl traverses the file tree and populates the storage with editor information
	Crawl(root *hs.FileTreeNode, pathBase string, ignoredFiles *[]string) error
}

// DefaultFileCrawler is the default implementation of FileCrawler
type DefaultFileCrawler struct {
	walker      FileTreeWalker
	codeStorage CodeStorage
	storage     EditorStorage
	userMapping UserMapper
}

// NewFileCrawler creates a new DefaultFileCrawler
func NewFileCrawler(w FileTreeWalker, codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper) *DefaultFileCrawler {
	return &DefaultFileCrawler{
		walker:      w,
		codeStorage: codeStorage,
		storage:     storage,
		userMapping: userMapping,
	}
}

// Crawl traverses the file tree and populates the storage with editor information
func (c *DefaultFileCrawler) Crawl(root *hs.FileTreeNode, pathBase string, ignoredFiles *[]string) error {
	return c.walker.Walk(root, func(node *hs.FileTreeNode, path string) error {
		editors, errorMsg := c.codeStorage.GetEditorsByFile(path)

		// If we have editors and user mapping, apply the mapping
		if editors != nil && c.userMapping != nil {
			mappedEditors := c.mapEditors(*editors)
			c.storage.SetFileEditors(path, &mappedEditors, errorMsg)
		} else {
			c.storage.SetFileEditors(path, editors, errorMsg)
		}

		return nil
	}, pathBase, ignoredFiles)
}

// mapEditors maps user IDs to display names in the editors map
func (c *DefaultFileCrawler) mapEditors(editors map[string]int) map[string]int {
	mappedEditors := make(map[string]int)
	for userID, lines := range editors {
		displayName := c.userMapping.GetDisplayName(userID)
		mappedEditors[displayName] += lines
	}
	return mappedEditors
}
