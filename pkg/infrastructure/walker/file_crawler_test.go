package walker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	an "we-know/pkg/infrastructure/crawler"
)

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

func TestDefaultFileCrawler_AnalyzeFiles(t *testing.T) {
	// Arrange
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)
	mockUserMapper := new(MockUserMapperTestify)

	crawler := an.NewFileCrawler(mockCodeStorage, mockEditorStorage, mockUserMapper)

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
	err := crawler.AnalyzeFiles([]string{"base/file.txt"})

	// Assert
	assert.NoError(t, err, "AnalyzeFiles should not return an error")
	mockCodeStorage.AssertExpectations(t)
	mockEditorStorage.AssertExpectations(t)
	mockUserMapper.AssertExpectations(t)
}

func TestDefaultFileCrawler_AnalyzeFiles_WithStorageError(t *testing.T) {
	// Arrange
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)
	mockUserMapper := new(MockUserMapperTestify)

	crawler := an.NewFileCrawler(mockCodeStorage, mockEditorStorage, mockUserMapper)

	// Code storage returns error message for the file
	mockCodeStorage.On("GetEditorsByFile", "base/file.txt").Return(nil, "test error")
	mockEditorStorage.On("SetFileEditors", "base/file.txt", (*map[string]int)(nil), "test error")

	// Act
	err := crawler.AnalyzeFiles([]string{"base/file.txt"})

	// Assert
	assert.NoError(t, err)
	mockCodeStorage.AssertExpectations(t)
	mockEditorStorage.AssertExpectations(t)
}

func TestDefaultFileCrawler_AnalyzeFiles_WithoutUserMapping(t *testing.T) {
	// Arrange
	mockCodeStorage := new(MockCodeStorageTestify)
	mockEditorStorage := new(MockEditorStorageTestify)

	// Create crawler without user mapping
	crawler := an.NewFileCrawler(mockCodeStorage, mockEditorStorage, nil)

	// Set up code storage to return editors for the test path
	editors := map[string]int{
		"user1": 10,
		"user2": 20,
	}
	mockCodeStorage.On("GetEditorsByFile", "base/file.txt").Return(&editors, "")

	// Set up editor storage to receive the unmapped editors
	mockEditorStorage.On("SetFileEditors", "base/file.txt", &editors, "")

	// Act
	err := crawler.AnalyzeFiles([]string{"base/file.txt"})

	// Assert
	assert.NoError(t, err, "AnalyzeFiles should not return an error")
	mockCodeStorage.AssertExpectations(t)
	mockEditorStorage.AssertExpectations(t)
}
