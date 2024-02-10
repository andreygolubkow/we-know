package historical_code_storage

type HistoricalCodeStorage interface {
	SetUp() error
	GetRootNode() *FileTreeNode
	GetEditorsByFile(filename string) (*map[string]int, string)
}

type FileTreeNode interface {
	GetName() string
	GetNext(ignoredFiles *[]string) []*FileTreeNode
	SetEditors(editors *[]string)
}
