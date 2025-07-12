package walker

import (
	hs "we-know/pkg/infrastructure/historical_code_storage"
	"we-know/pkg/infrastructure/user"
)

// TreeCallback is a function type that is called for each node in the file tree
type TreeCallback func(node *hs.FileTreeNode, fullPath string) error

// FileTreeWalker defines the interface for walking a file tree
type FileTreeWalker interface {
	// Walk traverses the file tree and calls the callback function for each node
	Walk(root *hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) error
}

// UserMapper defines the interface for mapping user IDs to display names
type UserMapper interface {
	// GetDisplayName returns the display name for the given user identifier
	GetDisplayName(userID string) string
}

// EditorStorage defines the interface for storing file editor information
type EditorStorage interface {
	// SetFileEditors sets the editors for a file
	SetFileEditors(filePath string, editors *map[string]int, errorMsg string)
}

// CodeStorage defines the interface for retrieving historical code information
type CodeStorage interface {
	// GetEditorsByFile returns the editors for a file
	GetEditorsByFile(filename string) (*map[string]int, string)
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

// NewFileTreeWalkerWithPathBuilder creates a new DefaultFileTreeWalker with a custom PathBuilder
func NewFileTreeWalkerWithPathBuilder(pathBuilder PathBuilder) *DefaultFileTreeWalker {
	return &DefaultFileTreeWalker{
		pathBuilder: pathBuilder,
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

// FileCrawler defines the interface for crawling a file tree and collecting editor information
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
func NewFileCrawler(walker FileTreeWalker, codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper) *DefaultFileCrawler {
	return &DefaultFileCrawler{
		walker:      walker,
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

// For backward compatibility
// Walk traverses the file tree and calls the callback function for each node
func Walk(root *hs.FileTreeNode, callback func(node *hs.FileTreeNode, fullPath string), pathBase string, ignoredFiles *[]string) {
	walker := NewFileTreeWalker()
	_ = walker.Walk(root, func(node *hs.FileTreeNode, fullPath string) error {
		callback(node, fullPath)
		return nil
	}, pathBase, ignoredFiles)
}

// Crawl traverses the file tree and populates the FileEditorsStorage with editor information for each file
// If userMapping is provided, it will be used to map user IDs to display names
func Crawl(root *hs.FileTreeNode, codeStorage hs.HistoricalCodeStorage, storage *hs.FileEditorsStorage, pathBase string, ignoredFiles *[]string, userMapping *user.UserMapping) {
	walker := NewFileTreeWalker()
	crawler := NewFileCrawler(walker, codeStorage, storage, userMapping)
	_ = crawler.Crawl(root, pathBase, ignoredFiles)
}
