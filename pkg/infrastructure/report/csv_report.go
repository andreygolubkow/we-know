package report

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"time"
	hs "we-know/pkg/infrastructure/historical_code_storage"
	"we-know/pkg/infrastructure/user"
)

// ReportType defines the type of report to generate
type ReportType int

const (
	// ReportByFileUsers generates a report with files as rows and users as columns
	ReportByFileUsers ReportType = iota
	// ReportByFileTeams generates a report with files as rows and teams as columns
	ReportByFileTeams
)

// CSVReport implements the Reporter interface for generating CSV reports
type CSVReport struct {
	// OutputDir is the directory where the report will be saved
	OutputDir string
	// UserMapping provides mapping between user identifiers and display names/teams
	UserMapping *user.UserMapping
	// ReportType defines the type of report to generate
	ReportType ReportType
}

// NewCSVReport creates a new CSVReport with the given output directory
// If userMapping is nil, no user mapping will be applied
// Default report type is ReportByFileUsers
func NewCSVReport(outputDir string, userMapping *user.UserMapping) *CSVReport {
	return &CSVReport{
		OutputDir:   outputDir,
		UserMapping: userMapping,
		ReportType:  ReportByFileUsers,
	}
}

// NewCSVReportWithType creates a new CSVReport with the given output directory and report type
// If userMapping is nil, no user mapping will be applied
func NewCSVReportWithType(outputDir string, userMapping *user.UserMapping, reportType ReportType) *CSVReport {
	return &CSVReport{
		OutputDir:   outputDir,
		UserMapping: userMapping,
		ReportType:  reportType,
	}
}

// GenerateReport generates a CSV report for the given code storage
// The report contains each scanned file and statistics about how many lines were changed by each user
func (r *CSVReport) GenerateReport(codeStorage hs.HistoricalCodeStorage) (string, error) {
	return r.GenerateReportFromStorage(codeStorage, nil)
}

// GenerateReportFromStorage generates a CSV report using the provided FileEditorsStorage
func (r *CSVReport) GenerateReportFromStorage(codeStorage hs.HistoricalCodeStorage, storage *hs.FileEditorsStorage) (string, error) {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.OutputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate a unique filename based on the current timestamp and report type
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	var reportTypeName string
	switch r.ReportType {
	case ReportByFileUsers:
		reportTypeName = "file_users"
	case ReportByFileTeams:
		reportTypeName = "file_teams"
	}
	filename := filepath.Join(r.OutputDir, fmt.Sprintf("file_changes_report_by_%s_%s.csv", reportTypeName, timestamp))

	// Create the CSV file
	file, err := os.Create(filename)
	if err != nil {
		return "", fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()

	// Create CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// If storage is nil, we can't generate a report
	if storage == nil {
		return "", fmt.Errorf("file editors storage is nil")
	}

	// Generate the appropriate report based on the report type
	switch r.ReportType {
	case ReportByFileUsers:
		err = r.generateReportByFiles(writer, storage)
	case ReportByFileTeams:
		err = r.generateReportByFileTeams(writer, storage)
	default:
		err = fmt.Errorf("unsupported report type: %d", r.ReportType)
	}

	if err != nil {
		return "", err
	}

	return filename, nil
}

// generateReportByFiles generates a CSV report with files as rows and users as columns
func (r *CSVReport) generateReportByFiles(writer *csv.Writer, storage *hs.FileEditorsStorage) error {
	// Write CSV header
	// First column is the file path, subsequent columns will be dynamically added for each user
	header := []string{"File Path"}
	userMap := make(map[string]int)          // Map to track column indices for users
	unmappedUserIDs := make(map[string]bool) // Track unmapped user IDs

	// Collect all files and user data
	fileData := make(map[string]map[string]int)

	// Get all files from storage
	files := storage.GetAllFiles()
	for _, path := range files {
		blame, _ := storage.GetFileEditors(path)
		if blame != nil {
			fileData[path] = *blame
			// Track all unique users
			for userID := range *blame {
				// Check if user is mapped
				isMapped := r.UserMapping != nil && r.UserMapping.GetUserInfo(userID) != nil

				if isMapped {
					if _, exists := userMap[userID]; !exists {
						// If user mapping is available, use display name
						displayName := r.UserMapping.GetDisplayName(userID)
						userMap[userID] = len(userMap) + 1
						header = append(header, displayName)
					}
				} else {
					// Mark as unmapped
					unmappedUserIDs[userID] = true
				}
			}
		}
	}

	// Add a single "Unmapped" column if there are any unmapped users
	unmappedColumnIndex := -1
	if len(unmappedUserIDs) > 0 {
		unmappedColumnIndex = len(header)
		header = append(header, "Unmapped")
	}

	// Write header
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for filePath, userData := range fileData {
		row := make([]string, len(header))
		row[0] = filePath // File path

		totalLines := 0
		unmappedLines := 0

		for _, lines := range userData {
			totalLines += lines
		}

		// Fill in user data and calculate unmapped lines
		for userID, lines := range userData {
			// Check if user is mapped
			isMapped := r.UserMapping != nil && r.UserMapping.GetUserInfo(userID) != nil

			if isMapped {
				colIndex := userMap[userID]
				row[colIndex] = fmt.Sprintf("%.2f", float64(lines)/float64(totalLines)*100)
			} else {
				// Accumulate unmapped lines
				unmappedLines += lines
			}
		}

		// Add unmapped percentage if there are any unmapped users
		if unmappedColumnIndex != -1 && totalLines > 0 {
			row[unmappedColumnIndex] = fmt.Sprintf("%.2f", float64(unmappedLines)/float64(totalLines)*100)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// generateReportByFileTeams generates a CSV report with files as rows and teams as columns
func (r *CSVReport) generateReportByFileTeams(writer *csv.Writer, storage *hs.FileEditorsStorage) error {
	// If no user mapping is available, we can't generate a team report
	if r.UserMapping == nil {
		return fmt.Errorf("user mapping is required for team report")
	}

	// Write CSV header
	// First column is the file path, subsequent columns will be dynamically added for each team
	header := []string{"File Path"}
	teamMap := make(map[string]int) // Map to track column indices for teams
	hasUnmappedUsers := false       // Flag to track if there are any unmapped users

	// Collect all files and user data
	fileData := make(map[string]map[string]int)

	// Get all files from storage
	files := storage.GetAllFiles()
	for _, path := range files {
		blame, _ := storage.GetFileEditors(path)
		if blame != nil {
			fileData[path] = *blame
			// Track all unique teams and check for unmapped users
			for userID := range *blame {
				userInfo := r.UserMapping.GetUserInfo(userID)
				if userInfo != nil && userInfo.Team != "" && userInfo.Team != "Unknown" {
					team := userInfo.Team
					if _, exists := teamMap[team]; !exists {
						teamMap[team] = len(teamMap) + 1
						header = append(header, team)
					}
				} else {
					// Mark that we have unmapped users
					hasUnmappedUsers = true
				}
			}
		}
	}

	// Add "Unmapped" column if there are any unmapped users
	unmappedColumnIndex := -1
	if hasUnmappedUsers {
		unmappedColumnIndex = len(header)
		header = append(header, "Unmapped")
	}

	// Write header
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for filePath, userData := range fileData {
		row := make([]string, len(header))
		row[0] = filePath // File path

		// Create a map to aggregate lines by team
		teamLines := make(map[string]int)
		unmappedLines := 0
		totalLines := 0

		// Calculate total lines
		for _, lines := range userData {
			totalLines += lines
		}

		// Aggregate lines by team and track unmapped lines
		for userID, lines := range userData {
			userInfo := r.UserMapping.GetUserInfo(userID)
			if userInfo != nil && userInfo.Team != "" && userInfo.Team != "Unknown" {
				teamLines[userInfo.Team] += lines
			} else {
				// Accumulate unmapped lines
				unmappedLines += lines
			}
		}

		// Fill in team data
		for team, lines := range teamLines {
			colIndex := teamMap[team]
			row[colIndex] = fmt.Sprintf("%.2f", float64(lines)/float64(totalLines)*100)
		}

		// Add unmapped percentage if there are any unmapped users
		if unmappedColumnIndex != -1 && totalLines > 0 {
			row[unmappedColumnIndex] = fmt.Sprintf("%.2f", float64(unmappedLines)/float64(totalLines)*100)
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}
