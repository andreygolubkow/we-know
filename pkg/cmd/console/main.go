package main

import (
	"log"
	"os"
	"we-know/pkg/infrastructure/historical_code_storage"
)

const (
	NothingToClone = "Nothing to clone"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal(NothingToClone)
		return
	}
	repositoryLink := os.Args[1]
	branch := os.Args[2]
	workingDir, _ := os.Getwd()
	tmpDir := workingDir + "/tmp"

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return
	}

	codeStorage, err := historical_code_storage.NewGitStorage(repositoryLink, branch, tmpDir)

	err = codeStorage.SetUp()

	if err != nil {
		log.Fatal(err)
		return
	}

	err = codeStorage.Update()

	if err != nil {
		log.Fatal(err)
		return
	}

	repo, err := historical_code_storage.SetUpRepository(tmpDir, repositoryLink, branch)
	if err != nil {
		log.Fatal(err)
	}

	refName, err := historical_code_storage.CheckoutBranch(repo, branch)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(refName)
}
