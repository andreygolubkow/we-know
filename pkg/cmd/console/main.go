package main

import (
	"errors"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"
)

// main.exe <giturl> <branch>
func main() {

	if len(os.Args) != 3 {
		log.Fatal("Nothing to clone")
		return
	}

	var repositoryLink string = os.Args[1]
	var branchName string = os.Args[2]
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
			URL:           repositoryLink,
			Progress:      os.Stdout,
			ReferenceName: plumbing.ReferenceName("refs/heads/" + branchName),
		})
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	ref, err := repo.Head()
	if err != nil {
		log.Fatal(err)
	}

	branch := ref.Name().Short()
	if branch != branchName {

	}

	log.Print("Branch:" + branch)

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
