package crawler

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
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
	concurrency int
	storageMu   sync.Mutex
}

// NewFileCrawler creates a new analyzer that works on a list of file paths.
func NewFileCrawler(codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper) *DefaultFileCrawler {
	c := &DefaultFileCrawler{
		codeStorage: codeStorage,
		storage:     storage,
		userMapping: userMapping,
		concurrency: runtime.NumCPU(),
	}
	if c.concurrency < 1 {
		c.concurrency = 1
	}
	return c
}

// NewFileCrawlerWithConcurrency creates a new analyzer with an explicit concurrency level.
func NewFileCrawlerWithConcurrency(codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper, concurrency int) *DefaultFileCrawler {
	if concurrency < 1 {
		concurrency = 1
	}
	return &DefaultFileCrawler{
		codeStorage: codeStorage,
		storage:     storage,
		userMapping: userMapping,
		concurrency: concurrency,
	}
}

// AnalyzeFiles iterates over file list and populates the storage with editor information
func (c *DefaultFileCrawler) AnalyzeFiles(files []string, reportProgress bool) error {
	// Fast path: sequential processing when concurrency is 1 or there is <=1 file
	if c.concurrency <= 1 || len(files) <= 1 {
		for i, path := range files {
			if reportProgress {
				fmt.Printf("Processing file %d/%d: %s\n", i+1, len(files), path)
			}
			c.processOne(path)
		}
		return nil
	}

	// Concurrent processing with a bounded worker pool
	jobs := make(chan string, c.concurrency*2)
	var wg sync.WaitGroup
	var processed int64
	total := int64(len(files))

	worker := func() {
		defer wg.Done()
		for path := range jobs {
			if reportProgress {
				cur := atomic.AddInt64(&processed, 1)
				fmt.Printf("Processing file %d/%d: %s\n", cur, total, path)
			}
			c.processOne(path)
		}
	}

	workers := c.concurrency
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}
	for _, path := range files {
		jobs <- path
	}
	close(jobs)
	wg.Wait()
	return nil
}

// processOne performs analysis for a single file path and stores the result safely.
func (c *DefaultFileCrawler) processOne(path string) {
	editors, errorMsg := c.codeStorage.GetEditorsByFile(path)

	if editors != nil && c.userMapping != nil {
		mappedEditors := c.mapEditors(*editors)
		c.storageMu.Lock()
		c.storage.SetFileEditors(path, &mappedEditors, errorMsg)
		c.storageMu.Unlock()
		return
	}
	c.storageMu.Lock()
	c.storage.SetFileEditors(path, editors, errorMsg)
	c.storageMu.Unlock()
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
