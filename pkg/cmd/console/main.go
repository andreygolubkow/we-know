package main

import (
	"errors"
	"fmt"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
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
		return
	}

	log.Print(ref.Name())

	// Fetch remote branches
	fetchErr := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"+refs/heads/*:refs/heads/*"},
	})
	if fetchErr != nil && fetchErr != git.NoErrAlreadyUpToDate {
		fmt.Println(fetchErr)
		os.Exit(1)
	}

	// Get the working directory for the repository
	w, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName("refs/heads/master"),
	})

	ref, err = repo.Head()
	if err != nil {
		log.Fatal(err)
		return
	}

	log.Print(ref.Name())
}
