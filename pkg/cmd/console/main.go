package main

import (
	"log"
	"os"
	"path/filepath"
	hs "we-know/pkg/infrastructure/historical_code_storage"
	"we-know/pkg/infrastructure/report"
	"we-know/pkg/infrastructure/user"
	"we-know/pkg/infrastructure/walker"
)

const (
	WrongArgumentsException = "Nothing to clone"
)

func main() {
	repoUrl, repoBranch, repoDir := readArguments()

	codeStorage, err := hs.NewGitStorage(repoUrl, repoBranch, repoDir)

	if err != nil {
		log.Fatal(err)
	}

	err = codeStorage.SetUp()

	if err != nil {
		log.Fatal(err)
		return
	}

	var rootPtr = codeStorage.GetRootNode()
	if rootPtr == nil {
		log.Fatal("Root node is nil")
		return
	}
	var root = *rootPtr
	log.Print(root.GetName())

	// Create a storage for file editors information
	fileEditorsStorage := hs.NewFileEditorsStorage()

	// Initialize user mapping
	userMappingPath := user.GetDefaultMappingFilePath()
	userMapping := user.NewUserMapping(userMappingPath)
	err = userMapping.LoadMappingFile()
	if err != nil {
		log.Printf("Warning: Failed to load user mapping file: %v", err)
		log.Printf("User mapping will not be applied to the report")
		userMapping = nil
	} else {
		log.Printf("User mapping loaded successfully from: %s", userMappingPath)
	}

	// Populate the storage with file editors information
	var ignoreList = []string{".git", ".idea", ".github"}
	var pathBase = ""
	walker.Crawl(rootPtr, codeStorage, fileEditorsStorage, pathBase, &ignoreList, userMapping)

	// Generate CSV report using the storage
	reportsDir := filepath.Join(repoDir, "reports")
	csvReporter := report.NewCSVReportWithType(reportsDir, userMapping, report.ReportByFileUsers)
	reportPath, err := csvReporter.GenerateReportFromStorage(codeStorage, fileEditorsStorage)
	if err != nil {
		log.Printf("Failed to generate report: %v", err)
	} else {
		log.Printf("Report generated successfully: %s", reportPath)
	}
	err = userMapping.SaveUnmappedUsers(filepath.Join(repoDir, "unmapped_users.csv"))
	if err != nil {
		log.Printf("Failed to save unmapped users: %v", err)
	}
}

/*
 */
func readArguments() (repositoryUrl string, repoBranch string, repositoryDir string) {
	if len(os.Args) != 3 {
		log.Fatal(WrongArgumentsException)
	}
	repositoryUrl = os.Args[1]
	branch := os.Args[2]
	workingDir, _ := os.Getwd()
	tmpDir := workingDir + "/tmp"

	info, err := os.Stat(tmpDir)

	if os.IsNotExist(err) || !info.IsDir() {
		err := os.MkdirAll(tmpDir, 0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	return repositoryUrl, branch, tmpDir
}
