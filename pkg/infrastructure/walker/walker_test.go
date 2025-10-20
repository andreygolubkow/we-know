package walker

import (
	"errors"
	"testing"
	an "we-know/pkg/infrastructure/analyzer"
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

// MockFileTreeNode is a mock implementation of hs.FileTreeNode for testing
type MockFileTreeNode struct {
	name     string
	children []*hs.FileTreeNode
}

func (m *MockFileTreeNode) GetName() string {
	return m.name
}

func (m *MockFileTreeNode) GetNext(ignoredFiles *[]string) []*hs.FileTreeNode {
	if ignoredFiles != nil {
		// Filter out ignored files
		filtered := make([]*hs.FileTreeNode, 0, len(m.children))
		for _, child := range m.children {
			ignored := false
			for _, ignoredFile := range *ignoredFiles {
				if (*child).GetName() == ignoredFile {
					ignored = true
					break
				}
			}
			if !ignored {
				filtered = append(filtered, child)
			}
		}
		return filtered
	}
	return m.children
}

func (m *MockFileTreeNode) SetEditors(editors *[]string) {
	// Not needed for tests
}

// TestDefaultPathBuilder tests the DefaultPathBuilder implementation
func TestDefaultPathBuilder(t *testing.T) {
	builder := NewPathBuilder()

	tests := []struct {
		name     string
		basePath string
		nodeName string
		expected string
	}{
		{
			name:     "Empty base path",
			basePath: "",
			nodeName: "file.txt",
			expected: "file.txt",
		},
		{
			name:     "With base path",
			basePath: "base",
			nodeName: "file.txt",
			expected: "base/file.txt",
		},
		{
			name:     "With nested base path",
			basePath: "base/dir",
			nodeName: "file.txt",
			expected: "base/dir/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := builder.BuildPath(tt.basePath, tt.nodeName)
			if result != tt.expected {
				t.Errorf("BuildPath(%q, %q) = %q, want %q", tt.basePath, tt.nodeName, result, tt.expected)
			}
		})
	}
}

// MockTreeCallback is a mock implementation of TreeCallback for testing
type MockTreeCallback struct {
	called     bool
	calledWith map[string]bool
	returnErr  error
}

func (m *MockTreeCallback) Call(node *hs.FileTreeNode, fullPath string) error {
	m.called = true
	if m.calledWith == nil {
		m.calledWith = make(map[string]bool)
	}
	m.calledWith[fullPath] = true
	return m.returnErr
}

// TestDefaultFileTreeWalker tests the DefaultFileTreeWalker implementation
func TestDefaultFileTreeWalker(t *testing.T) {
	// Create a simple file tree for testing
	file1 := &MockFileTreeNode{name: "file1.txt"}
	file2 := &MockFileTreeNode{name: "file2.txt"}
	dir1 := &MockFileTreeNode{name: "dir1"}
	file3 := &MockFileTreeNode{name: "file3.txt"}

	// Convert to hs.FileTreeNode interface
	file1Node := hs.FileTreeNode(file1)
	file2Node := hs.FileTreeNode(file2)
	dir1Node := hs.FileTreeNode(dir1)
	file3Node := hs.FileTreeNode(file3)

	// Set up the tree structure
	dir1.children = []*hs.FileTreeNode{&file3Node}
	root := &MockFileTreeNode{
		name:     "root",
		children: []*hs.FileTreeNode{&file1Node, &file2Node, &dir1Node},
	}
	rootNode := hs.FileTreeNode(root)

	t.Run("Walk with nil root", func(t *testing.T) {
		walker := NewFileTreeWalker()
		callback := &MockTreeCallback{}

		err := walker.Walk(nil, callback.Call, "", nil)

		if err != nil {
			t.Errorf("Walk(nil, ...) returned error: %v", err)
		}

		if callback.called {
			t.Error("Callback was called with nil root")
		}
	})

	t.Run("Walk with valid tree", func(t *testing.T) {
		walker := NewFileTreeWalker()
		callback := &MockTreeCallback{}

		err := walker.Walk(&rootNode, callback.Call, "", nil)

		if err != nil {
			t.Errorf("Walk returned error: %v", err)
		}

		if !callback.called {
			t.Error("Callback was not called")
		}

		// Check that all nodes were visited
		expectedPaths := []string{
			"root/file1.txt",
			"root/file2.txt",
			"root/dir1",
			"root/dir1/file3.txt",
		}

		for _, path := range expectedPaths {
			if !callback.calledWith[path] {
				t.Errorf("Callback was not called with path: %s", path)
			}
		}
	})

	t.Run("Walk with error in callback", func(t *testing.T) {
		walker := NewFileTreeWalker()
		expectedErr := errors.New("test error")
		callback := &MockTreeCallback{returnErr: expectedErr}

		err := walker.Walk(&rootNode, callback.Call, "", nil)

		if err != expectedErr {
			t.Errorf("Walk did not return expected error, got: %v, want: %v", err, expectedErr)
		}
	})

	t.Run("Walk with ignored files", func(t *testing.T) {
		walker := NewFileTreeWalker()
		callback := &MockTreeCallback{}
		ignoredFiles := []string{"file2.txt"}

		err := walker.Walk(&rootNode, callback.Call, "", &ignoredFiles)

		if err != nil {
			t.Errorf("Walk returned error: %v", err)
		}

		// Check that file2.txt was not visited
		if callback.calledWith["root/file2.txt"] {
			t.Error("Callback was called with ignored path: root/file2.txt")
		}

		// Check that other nodes were visited
		expectedPaths := []string{
			"root/file1.txt",
			"root/dir1",
			"root/dir1/file3.txt",
		}

		for _, path := range expectedPaths {
			if !callback.calledWith[path] {
				t.Errorf("Callback was not called with path: %s", path)
			}
		}
	})
}

// MockUserMapper is a mock implementation of UserMapper for testing
type MockUserMapper struct {
	displayNames map[string]string
}

func (m *MockUserMapper) GetDisplayName(userID string) string {
	if name, ok := m.displayNames[userID]; ok {
		return name
	}
	return userID
}

// MockEditorStorage is a mock implementation of EditorStorage for testing
type MockEditorStorage struct {
	fileEditors map[string]*map[string]int
	fileErrors  map[string]string
}

func NewMockEditorStorage() *MockEditorStorage {
	return &MockEditorStorage{
		fileEditors: make(map[string]*map[string]int),
		fileErrors:  make(map[string]string),
	}
}

func (m *MockEditorStorage) SetFileEditors(filePath string, editors *map[string]int, errorMsg string) {
	m.fileEditors[filePath] = editors
	if errorMsg != "" {
		m.fileErrors[filePath] = errorMsg
	}
}

// MockCodeStorage is a mock implementation of CodeStorage for testing
type MockCodeStorage struct {
	editors  map[string]*map[string]int
	errorMsg map[string]string
}

func (m *MockCodeStorage) GetEditorsByFile(filename string) (*map[string]int, string) {
	editors, ok := m.editors[filename]
	if !ok {
		return nil, "File not found"
	}

	errorMsg, ok := m.errorMsg[filename]
	if !ok {
		errorMsg = ""
	}

	return editors, errorMsg
}

// TestDefaultFileCrawler tests the DefaultFileCrawler implementation
func TestDefaultFileCrawler(t *testing.T) {
	// Create a simple file tree for testing
	file1 := &MockFileTreeNode{name: "file1.txt"}
	file1Node := hs.FileTreeNode(file1)
	root := &MockFileTreeNode{
		name:     "root",
		children: []*hs.FileTreeNode{&file1Node},
	}
	rootNode := hs.FileTreeNode(root)

	// Create mock dependencies
	walker := NewFileTreeWalker()

	t.Run("Crawl with user mapping", func(t *testing.T) {
		// Set up mock code storage
		codeStorage := &MockCodeStorage{
			editors: map[string]*map[string]int{
				"root/file1.txt": {
					"user1": 10,
					"user2": 20,
				},
			},
			errorMsg: map[string]string{
				"root/file1.txt": "",
			},
		}

		// Set up mock user mapper
		userMapper := &MockUserMapper{
			displayNames: map[string]string{
				"user1": "John Doe",
				"user2": "Jane Smith",
			},
		}

		// Set up mock editor storage
		editorStorage := NewMockEditorStorage()

		// Create crawler
		crawler := an.NewFileCrawler(walker, codeStorage, editorStorage, userMapper)

		// Run crawl
		err := crawler.Crawl(&rootNode, "", nil)

		// Check results
		if err != nil {
			t.Errorf("Crawl returned error: %v", err)
		}

		// Check that editors were mapped correctly
		editors := editorStorage.fileEditors["root/file1.txt"]
		if editors == nil {
			t.Fatal("No editors found for file1.txt")
		}

		expected := map[string]int{
			"John Doe":   10,
			"Jane Smith": 20,
		}

		for name, lines := range expected {
			if (*editors)[name] != lines {
				t.Errorf("Expected %s to have %d lines, got %d", name, lines, (*editors)[name])
			}
		}
	})

	t.Run("Crawl without user mapping", func(t *testing.T) {
		// Set up mock code storage
		fileEditors := map[string]int{
			"user1": 10,
			"user2": 20,
		}
		codeStorage := &MockCodeStorage{
			editors: map[string]*map[string]int{
				"root/file1.txt": &fileEditors,
			},
			errorMsg: map[string]string{
				"root/file1.txt": "",
			},
		}

		// Set up mock editor storage
		editorStorage := NewMockEditorStorage()

		// Create crawler without user mapping
		crawler := an.NewFileCrawler(walker, codeStorage, editorStorage, nil)

		// Run crawl
		err := crawler.Crawl(&rootNode, "", nil)

		// Check results
		if err != nil {
			t.Errorf("Crawl returned error: %v", err)
		}

		// Check that editors were not mapped
		editors := editorStorage.fileEditors["root/file1.txt"]
		if editors == nil {
			t.Fatal("No editors found for file1.txt")
		}

		// Should be the same as the original
		if editors != &fileEditors {
			t.Error("Editors were modified when they shouldn't have been")
		}
	})
}

// TestBackwardCompatibility tests the backward compatibility functions
func TestBackwardCompatibility(t *testing.T) {
	// Create a simple file tree for testing
	file1 := &MockFileTreeNode{name: "file1.txt"}
	file1Node := hs.FileTreeNode(file1)
	root := &MockFileTreeNode{
		name:     "root",
		children: []*hs.FileTreeNode{&file1Node},
	}
	rootNode := hs.FileTreeNode(root)

	t.Run("Walk compatibility", func(t *testing.T) {
		called := false
		calledWith := make(map[string]bool)

		Walk(&rootNode, func(node *hs.FileTreeNode, fullPath string) {
			called = true
			calledWith[fullPath] = true
		}, "", nil)

		if !called {
			t.Error("Callback was not called")
		}

		if !calledWith["root/file1.txt"] {
			t.Error("Callback was not called with path: root/file1.txt")
		}
	})
}
