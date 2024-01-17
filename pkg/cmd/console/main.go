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

const (
	RefFormat      = "refs/heads/%s"
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

	repo, err := setUpRepository(tmpDir, repositoryLink, branch)
	if err != nil {
		log.Fatal(err)
	}

	refName, err := checkoutBranch(repo, branch)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(refName)
}

func setUpRepository(tmpDir, repositoryLink, branch string) (rep *git.Repository, err error) {
	rep, err = git.PlainOpen(tmpDir)

	if errors.Is(err, git.ErrRepositoryNotExists) {
		rep, err = git.PlainClone(tmpDir, false, &git.CloneOptions{
			URL:           repositoryLink,
			Progress:      os.Stdout,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf(RefFormat, branch)),
		})
	}

	return
}

func checkoutBranch(repo *git.Repository, branch string) (branchName string, err error) {
	ref, err := repo.Head()
	if err != nil {
		return
	}

	branchName = ref.Name().Short()
	if branchName != branch {
		err = fetchAndCheckout(repo, branch)
		if err != nil {
			return
		}
	}

	return
}

func fetchAndCheckout(repo *git.Repository, branch string) error {
	err := repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []config.RefSpec{"+refs/heads/*:refs/heads/*"},
	})

	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}

	workTree, err := repo.Worktree()
	if err != nil {
		return err
	}

	err = workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.ReferenceName(fmt.Sprintf(RefFormat, branch)),
	})

	return err
}
