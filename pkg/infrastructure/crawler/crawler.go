package crawler

import "fmt"

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

// Analyzer defines behavior for analyzing a provided list of file paths using
// the CodeStorage and persisting results to EditorStorage.
type Analyzer interface {
	// AnalyzeFiles processes the given list of file paths and stores editor info
	AnalyzeFiles(files []string) error
}

// DefaultFileCrawler (now analyzer) processes provided file paths.
type DefaultFileCrawler struct {
	codeStorage CodeStorage
	storage     EditorStorage
	userMapping UserMapper
}

// NewFileCrawler creates a new analyzer that works on a list of file paths.
func NewFileCrawler(codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper) *DefaultFileCrawler {
	return &DefaultFileCrawler{
		codeStorage: codeStorage,
		storage:     storage,
		userMapping: userMapping,
	}
}

// AnalyzeFiles iterates over file list and populates the storage with editor information
func (c *DefaultFileCrawler) AnalyzeFiles(files []string, reportProgress bool) error {
	for i, path := range files {
		if reportProgress {
			fmt.Printf("Processing file %d/%d: %s\n", i+1, len(files), path)
		}
		editors, errorMsg := c.codeStorage.GetEditorsByFile(path)

		// If we have editors and user mapping, apply the mapping
		if editors != nil && c.userMapping != nil {
			mappedEditors := c.mapEditors(*editors)
			c.storage.SetFileEditors(path, &mappedEditors, errorMsg)
		} else {
			c.storage.SetFileEditors(path, editors, errorMsg)
		}
	}
	return nil
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
