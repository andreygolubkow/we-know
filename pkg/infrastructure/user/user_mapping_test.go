package user

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestUnmappedUsers(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "user_mapping_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test mapping file
	mappingFilePath := filepath.Join(tempDir, "test_mapping.csv")
	mappingContent := `John Doe,Team A,john.doe,jdoe
Jane Smith,Team B,jane.smith,jsmith`
	err = os.WriteFile(mappingFilePath, []byte(mappingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test mapping file: %v", err)
	}

	// Create a UserMapping and load the mapping file
	userMapping := NewUserMapping(mappingFilePath)
	err = userMapping.LoadMappingFile()
	if err != nil {
		t.Fatalf("Failed to load mapping file: %v", err)
	}

	// Test with mapped users
	mappedUser1 := "john.doe"
	mappedUser2 := "jsmith"

	// Test with unmapped users
	unmappedUser1 := "unknown.user"
	unmappedUser2 := "another.unknown"

	// Get user info for both mapped and unmapped users
	info1 := userMapping.GetUserInfo(mappedUser1)
	info2 := userMapping.GetUserInfo(mappedUser2)
	info3 := userMapping.GetUserInfo(unmappedUser1)
	info4 := userMapping.GetUserInfo(unmappedUser2)

	// Verify mapped users have info
	if info1 == nil {
		t.Errorf("Expected user info for %s, got nil", mappedUser1)
	}
	if info2 == nil {
		t.Errorf("Expected user info for %s, got nil", mappedUser2)
	}

	// Verify unmapped users have no info
	if info3 != nil {
		t.Errorf("Expected nil for unmapped user %s, got %+v", unmappedUser1, info3)
	}
	if info4 != nil {
		t.Errorf("Expected nil for unmapped user %s, got %+v", unmappedUser2, info4)
	}

	// Get the list of unmapped users
	unmappedUsers := userMapping.GetUnmappedUsers()

	// Verify unmapped users are tracked
	if len(unmappedUsers) != 2 {
		t.Errorf("Expected 2 unmapped users, got %d", len(unmappedUsers))
	}

	// Check if unmapped users are in the list
	foundUser1 := false
	foundUser2 := false
	for _, user := range unmappedUsers {
		if user == unmappedUser1 {
			foundUser1 = true
		}
		if user == unmappedUser2 {
			foundUser2 = true
		}
	}
	if !foundUser1 {
		t.Errorf("Unmapped user %s not found in the list", unmappedUser1)
	}
	if !foundUser2 {
		t.Errorf("Unmapped user %s not found in the list", unmappedUser2)
	}

	// Save unmapped users to a file
	unmappedFilePath := filepath.Join(tempDir, "unmapped_users.csv")
	err = userMapping.SaveUnmappedUsers(unmappedFilePath)
	if err != nil {
		t.Fatalf("Failed to save unmapped users: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(unmappedFilePath); os.IsNotExist(err) {
		t.Errorf("Unmapped users file was not created")
	}

	// Read the file content to verify
	content, err := os.ReadFile(unmappedFilePath)
	if err != nil {
		t.Fatalf("Failed to read unmapped users file: %v", err)
	}

	// Check if the file contains the header and unmapped users
	expectedContent := "UserID\n" + unmappedUser1 + "\n" + unmappedUser2 + "\n"
	alternativeContent := "UserID\n" + unmappedUser2 + "\n" + unmappedUser1 + "\n"

	contentStr := string(content)
	if contentStr != expectedContent && contentStr != alternativeContent {
		t.Errorf("Unexpected file content. Got:\n%s\nExpected either:\n%s\nor\n%s", contentStr, expectedContent, alternativeContent)
	}

	fmt.Println("Test completed successfully")
}
