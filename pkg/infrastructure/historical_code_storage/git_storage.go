package historical_code_storage

import (
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"os"
)

const RefFormat = "refs/heads/%s"

type GitSettings struct {
	url             string
	branch          string
	pathToDirectory string
}

type GitStorage struct {
	settings GitSettings
}

func (g GitStorage) SetUp() error {
	//TODO implement me
	panic("implement me")
}

func (g GitStorage) Update() error {
	//TODO implement me
	panic("implement me")
}

func (g GitStorage) Cleanup() error {
	//TODO implement me
	panic("implement me")
}

func NewGitStorage(url string, branch string, dir string) (HistoricalCodeStorage, error) {
	return &GitStorage{
		settings: GitSettings{url, branch, dir},
	}, nil
}

func SetUpRepository(tmpDir, repositoryLink, branch string) (rep *git.Repository, err error) {
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

func CheckoutBranch(repo *git.Repository, branch string) (branchName string, err error) {
	ref, err := repo.Head()
	if err != nil {
		return
	}

	branchName = ref.Name().Short()
	if branchName != branch {
		err = FetchAndCheckout(repo, branch)
		if err != nil {
			return
		}
	}

	return
}

func FetchAndCheckout(repo *git.Repository, branch string) error {
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
