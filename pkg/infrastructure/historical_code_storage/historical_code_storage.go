package historical_code_storage

type HistoricalCodeStorage interface {
	SetUp() error
	GetRootNode() *FileTreeNode
}

type FileTreeNode interface {
	GetName() string
	GetNext(ignoredFiles *[]string) []*FileTreeNode
}
