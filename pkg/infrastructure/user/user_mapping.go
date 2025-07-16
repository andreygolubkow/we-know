package user

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// UserInfo represents information about a user
type UserInfo struct {
	DisplayName string
	Team        string
	Aliases     []string
}

// UserMapping provides functionality to map user identifiers to user information
// The mapping is loaded from a CSV file with the following format:
// - First column: Display name of the user
// - Second column: Team name of the user
// - Remaining columns: Aliases for the user (email, username, etc.)
// Each user can have a different number of aliases.
// The display name is also automatically added as an alias.
type UserMapping struct {
	// Map from user identifier (email, username, display name, etc.) to UserInfo
	userMap map[string]*UserInfo
	// Path to the CSV file containing user mapping information
	mappingFilePath string
	// Set of unmapped user identifiers
	unmappedUsers map[string]bool
}

// NewUserMapping creates a new UserMapping with the given mapping file path
func NewUserMapping(mappingFilePath string) *UserMapping {
	return &UserMapping{
		userMap:         make(map[string]*UserInfo),
		mappingFilePath: mappingFilePath,
		unmappedUsers:   make(map[string]bool),
	}
}

// LoadMappingFile loads the user mapping from the CSV file
func (m *UserMapping) LoadMappingFile() error {
	// Check if file exists
	if _, err := os.Stat(m.mappingFilePath); os.IsNotExist(err) {
		return fmt.Errorf("user mapping file does not exist: %s", m.mappingFilePath)
	}

	// Open the CSV file
	file, err := os.Open(m.mappingFilePath)
	if err != nil {
		return fmt.Errorf("failed to open user mapping file: %w", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)
	// Allow variable number of fields per record
	reader.FieldsPerRecord = -1

	// Read all records
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading CSV record: %w", err)
		}

		// Skip if not enough columns (need at least display name and team)
		if len(record) < 2 {
			continue
		}

		displayName := strings.TrimSpace(record[0])
		team := strings.TrimSpace(record[1])

		// Create user info with pre-allocated aliases slice
		// Pre-allocate with capacity for all potential aliases (record length - 2 for display name and team + 1 for display name itself)
		userInfo := &UserInfo{
			DisplayName: displayName,
			Team:        team,
			Aliases:     make([]string, 0, len(record)-1), // -2 for display name and team + 1 for display name itself
		}

		// Add the display name as an alias too (if not empty)
		if displayName != "" {
			userInfo.Aliases = append(userInfo.Aliases, displayName)
			m.userMap[displayName] = userInfo
		}

		// Add aliases (remaining columns)
		for i := 2; i < len(record); i++ {
			alias := strings.TrimSpace(record[i])
			if alias != "" {
				userInfo.Aliases = append(userInfo.Aliases, alias)
				// Map alias to user info
				m.userMap[alias] = userInfo
			}
		}
	}

	return nil
}

// GetUserInfo returns the UserInfo for the given user identifier
// If the user is not found, it will be added to the unmapped users list
func (m *UserMapping) GetUserInfo(userID string) *UserInfo {
	if info, exists := m.userMap[userID]; exists {
		return info
	}
	// Record unmapped user
	m.unmappedUsers[userID] = true
	return nil
}

// GetDisplayName returns the display name for the given user identifier
// If no mapping exists, returns the original identifier
func (m *UserMapping) GetDisplayName(userID string) string {
	if info := m.GetUserInfo(userID); info != nil {
		return info.DisplayName
	}
	return userID
}

// GetTeam returns the team for the given user identifier
// If no mapping exists, returns an empty string
func (m *UserMapping) GetTeam(userID string) string {
	if info := m.GetUserInfo(userID); info != nil {
		return info.Team
	}
	return ""
}

// GetDefaultMappingFilePath returns the default path for the user mapping file
// It creates the config directory if it doesn't exist
func GetDefaultMappingFilePath() string {
	workingDir, _ := os.Getwd()
	configDir := filepath.Join(workingDir, "config")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		// If we can't create the directory, fall back to the working directory
		return filepath.Join(workingDir, "user_mapping.csv")
	}

	return filepath.Join(configDir, "user_mapping.csv")
}

// SaveUnmappedUsers saves the list of unmapped users to a CSV file
// The file will contain one user ID per line
func (m *UserMapping) SaveUnmappedUsers(filePath string) error {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory for unmapped users file: %w", err)
	}

	_, err := os.Stat(filePath)
	newFile := false
	if os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create unmapped users file: %w", err)
		}
		newFile = true
		defer file.Close()
	}
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create unmapped users file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	if newFile {
		if err := writer.Write([]string{"UserID"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
	}

	// Write unmapped users
	for userID := range m.unmappedUsers {
		if err := writer.Write([]string{userID}); err != nil {
			return fmt.Errorf("failed to write user ID to CSV: %w", err)
		}
	}

	return nil
}

// GetUnmappedUsers returns a slice of all unmapped user IDs
func (m *UserMapping) GetUnmappedUsers() []string {
	users := make([]string, 0, len(m.unmappedUsers))
	for userID := range m.unmappedUsers {
		users = append(users, userID)
	}
	return users
}
