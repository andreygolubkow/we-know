package walker

import (
	an "we-know/pkg/infrastructure/analyzer"
	hs "we-know/pkg/infrastructure/historical_code_storage"
	"we-know/pkg/infrastructure/user"
)

// Backward compatibility layer: expose analysis types via walker, but delegate implementation to analyzer.

// UserMapper maps raw user IDs to display names.
// Deprecated: use analyzer.UserMapper directly.
type UserMapper = an.UserMapper

// EditorStorage stores aggregated editors per file.
// Deprecated: use analyzer.EditorStorage directly.
type EditorStorage = an.EditorStorage

// CodeStorage provides historical code (e.g., Git) information.
// Deprecated: use analyzer.CodeStorage directly.
type CodeStorage = an.CodeStorage

// FileCrawler crawls files (via walker) and performs analysis.
// Deprecated: use analyzer.FileCrawler directly.
type FileCrawler = an.FileCrawler

// DefaultFileCrawler is the default implementation of FileCrawler.
// Deprecated: use analyzer.DefaultFileCrawler directly.
type DefaultFileCrawler = an.DefaultFileCrawler

// NewFileCrawler creates a new DefaultFileCrawler using analyzer implementation.
// Deprecated: use analyzer.NewFileCrawler directly.
func NewFileCrawler(walker FileTreeWalker, codeStorage CodeStorage, storage EditorStorage, userMapping UserMapper) *DefaultFileCrawler {
	return an.NewFileCrawler(walker, codeStorage, storage, userMapping)
}

// Crawl traverses the file tree and populates the FileEditorsStorage with editor information for each file.
// This wrapper preserves the previous walker.Crawl entry point while delegating analysis to analyzer.
func Crawl(root *hs.FileTreeNode, codeStorage hs.HistoricalCodeStorage, storage *hs.FileEditorsStorage, pathBase string, ignoredFiles *[]string, userMapping *user.UserMapping) {
	w := NewFileTreeWalker()
	crawler := an.NewFileCrawler(w, codeStorage, storage, userMapping)
	_ = crawler.Crawl(root, pathBase, ignoredFiles)
}
