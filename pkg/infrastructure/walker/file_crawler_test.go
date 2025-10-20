package walker

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	an "we-know/pkg/infrastructure/crawler"
	hs "we-know/pkg/infrastructure/historical_code_storage"
)

// MockFileTreeWalker is a mock implementation of FileTreeWalker
type MockFileTreeWalker struct {
	mock.Mock
}

func (m *MockFileTreeWalker) Walk(root *hs.FileTreeNode, callback TreeCallback, pathBase string, ignoredFiles *[]string) error {
	args := m.Called(root, callback, pathBase, ignoredFiles)
	return args.Error(0)
}

// MockCodeStorage is a mock implementation of CodeStorage
type MockCodeStorageTestify struct {
	mock.Mock
}

func (m *MockCodeStorageTestify) GetEditorsByFile(filename string) (*map[string]int, string) {
	args := m.Called(filename)

	if args.Get(0) == nil {
		return nil, args.String(1)
	}

	return args.Get(0).(*map[string]int), args.String(1)
}

// MockEditorStorage is a mock implementation of EditorStorage
type MockEditorStorageTestify struct {
	mock.Mock
}

func (m *MockEditorStorageTestify) SetFileEditors(filePath string, editors *map[string]int, errorMsg string) {
	m.Called(filePath, editors, errorMsg)
}

// MockUserMapper is a mock implementation of UserMapper
type MockUserMapperTestify struct {
	mock.Mock
}

func (m *MockUserMapperTestify) GetDisplayName(userID string) string {
	args := m.Called(userID)
	return args.String(0)
}

func TestDefaultFileCrawler_Crawl(t *testing.T) {
	// Arrange
	mockWalker := new(MockFileTreeWalker)
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)
	mockUserMapper := new(MockUserMapperTestify)

	crawler := an.NewFileCrawler(mockWalker, mockCodeStorage, mockEditorStorage, mockUserMapper)

	// Create a mock root node
	mockRoot := new(MockFileTreeNode)
	var rootNode hs.FileTreeNode = mockRoot

	// Set up walker to call the callback with a test path
	mockWalker.On("Walk", &rootNode, mock.Anything, "base", mock.Anything).
		Run(func(args mock.Arguments) {
			callback := args.Get(1).(TreeCallback)
			callback(nil, "base/file.txt")
		}).
		Return(nil)

	// Set up code storage to return editors for the test path
	editors := map[string]int{
		"user1": 10,
		"user2": 20,
	}
	mockCodeStorage.On("GetEditorsByFile", "base/file.txt").Return(&editors, "")

	// Set up user mapper to map user IDs to display names
	mockUserMapper.On("GetDisplayName", "user1").Return("John Doe")
	mockUserMapper.On("GetDisplayName", "user2").Return("Jane Smith")

	// Set up editor storage to receive the mapped editors
	mockEditorStorage.On("SetFileEditors", "base/file.txt", mock.AnythingOfType("*map[string]int"), "").
		Run(func(args mock.Arguments) {
			mappedEditors := args.Get(1).(*map[string]int)
			assert.Equal(t, 10, (*mappedEditors)["John Doe"], "John Doe should have 10 lines")
			assert.Equal(t, 20, (*mappedEditors)["Jane Smith"], "Jane Smith should have 20 lines")
		})

	// Act
	err := crawler.Crawl(&rootNode, "base", nil)

	// Assert
	assert.NoError(t, err, "Crawl should not return an error")
	mockWalker.AssertExpectations(t)
	mockCodeStorage.AssertExpectations(t)
	mockEditorStorage.AssertExpectations(t)
	mockUserMapper.AssertExpectations(t)
}

func TestDefaultFileCrawler_Crawl_WithError(t *testing.T) {
	// Arrange
	mockWalker := new(MockFileTreeWalker)
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)
	mockUserMapper := new(MockUserMapperTestify)

	crawler := an.NewFileCrawler(mockWalker, mockCodeStorage, mockEditorStorage, mockUserMapper)

	// Create a mock root node
	mockRoot := new(MockFileTreeNode)
	var rootNode hs.FileTreeNode = mockRoot

	// Set up walker to return an error
	expectedErr := errors.New("test error")
	mockWalker.On("Walk", &rootNode, mock.Anything, "base", mock.Anything).
		Return(expectedErr)

	// Act
	err := crawler.Crawl(&rootNode, "base", nil)

	// Assert
	assert.Equal(t, expectedErr, err, "Crawl should return the error from Walk")
	mockWalker.AssertExpectations(t)
}

func TestDefaultFileCrawler_Crawl_WithoutUserMapping(t *testing.T) {
	// Arrange
	mockWalker := new(MockFileTreeWalker)
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)

	// Create crawler without user mapping
	crawler := an.NewFileCrawler(mockWalker, mockCodeStorage, mockEditorStorage, nil)

	// Create a mock root node
	mockRoot := new(MockFileTreeNode)
	var rootNode hs.FileTreeNode = mockRoot

	// Set up walker to call the callback with a test path
	mockWalker.On("Walk", &rootNode, mock.Anything, "base", mock.Anything).
		Run(func(args mock.Arguments) {
			callback := args.Get(1).(TreeCallback)
			callback(nil, "base/file.txt")
		}).
		Return(nil)

	// Set up code storage to return editors for the test path
	editors := map[string]int{
		"user1": 10,
		"user2": 20,
	}
	mockCodeStorage.On("GetEditorsByFile", "base/file.txt").Return(&editors, "")

	// Set up editor storage to receive the unmapped editors
	mockEditorStorage.On("SetFileEditors", "base/file.txt", &editors, "")

	// Act
	err := crawler.Crawl(&rootNode, "base", nil)

	// Assert
	assert.NoError(t, err, "Crawl should not return an error")
	mockWalker.AssertExpectations(t)
	mockCodeStorage.AssertExpectations(t)
	mockEditorStorage.AssertExpectations(t)
}
