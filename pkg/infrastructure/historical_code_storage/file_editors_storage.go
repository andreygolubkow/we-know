package historical_code_storage

// FileEditorsStorage is a storage for file editors information
type FileEditorsStorage struct {
	// Map of file paths to editor information
	fileEditors map[string]*map[string]int
	// Map of file paths to error messages
	fileErrors map[string]string
}

// NewFileEditorsStorage creates a new FileEditorsStorage
func NewFileEditorsStorage() *FileEditorsStorage {
	return &FileEditorsStorage{
		fileEditors: make(map[string]*map[string]int),
		fileErrors:  make(map[string]string),
	}
}

// SetFileEditors sets the editors for a file
func (s *FileEditorsStorage) SetFileEditors(filePath string, editors *map[string]int, errorMsg string) {
	s.fileEditors[filePath] = editors
	if errorMsg != "" {
		s.fileErrors[filePath] = errorMsg
	}
}

// GetFileEditors gets the editors for a file
func (s *FileEditorsStorage) GetFileEditors(filePath string) (*map[string]int, string) {
	editors, exists := s.fileEditors[filePath]
	if !exists {
		return nil, "File not found in storage"
	}

	errorMsg, exists := s.fileErrors[filePath]
	if !exists {
		errorMsg = ""
	}

	return editors, errorMsg
}

// GetAllFiles returns all file paths in the storage
func (s *FileEditorsStorage) GetAllFiles() []string {
	files := make([]string, 0, len(s.fileEditors))
	for file := range s.fileEditors {
		files = append(files, file)
	}
	return files
}
