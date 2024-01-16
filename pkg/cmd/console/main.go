package main

import (
	"errors"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"
)

func main() {

	if len(os.Args) == 1 {
		log.Fatal("Nothing to clone")
		return
	}

	var repositoryLink string = os.Args[1]
	var workingDir, _ = os.Getwd()
	var tmpDir = workingDir + "/tmp"
	var dirErr = os.MkdirAll(tmpDir, 0755)
	if dirErr != nil {
		return
	}

	var repo *git.Repository
	repo, err := git.PlainOpen(tmpDir)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
			URL:      repositoryLink,
			Progress: os.Stdout,
		})
		if err != nil {
			log.Fatal(err)
			return
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	branches, _ := repo.Branches()
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		log.Println(ref.Name())
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return
	}
}
