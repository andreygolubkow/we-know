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
	repo     *git.Repository
	root     *GitNode
}

func NewGitStorage(url string, branch string, dir string) (HistoricalCodeStorage, error) {
	return &GitStorage{
		settings: GitSettings{url, branch, dir},
	}, nil
}

func (g *GitStorage) SetUp() error {
	repo, err := git.PlainOpen(g.settings.pathToDirectory)

	if err != nil && !errors.Is(err, git.ErrRepositoryNotExists) {
		return err
	}
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.PlainClone(g.settings.pathToDirectory, false, &git.CloneOptions{
			URL:           g.settings.url,
			Progress:      os.Stdout,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf(RefFormat, g.settings.branch)),
		})

		if err != nil {
			return err
		}
	}

	if len(g.settings.branch) != 0 {
		err = CheckoutBranch(repo, g.settings.branch)
		if err != nil {
			return err
		}
	}

	g.repo = repo
	g.root = NewGitNode("", g.settings.pathToDirectory)
	return nil
}

func (g *GitStorage) GetRootNode() *FileTreeNode {
	if g.root == nil {
		return nil
	}

	var node FileTreeNode = g.root
	return &node
}

func CheckoutBranch(repo *git.Repository, branch string) error {
	ref, err := repo.Head()
	if err != nil {
		return err
	}

	branchName := ref.Name().Short()
	if branchName != branch {
		err = FetchAndCheckout(repo, branch)
		if err != nil {
			return err
		}
	}
	return nil
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

func (g *GitStorage) GetEditorsByFile(filename string) (*map[string]int, string) {
	head, err := g.repo.Head()

	if err != nil {
		return nil, err.Error()
	}

	commit, err := g.repo.CommitObject(head.Hash())
	if err != nil {
		return nil, err.Error()
	}

	blame, err := git.Blame(commit, filename)
	if err != nil {
		return nil, err.Error()
	}

	result := map[string]int{}

	for _, v := range blame.Lines {
		_, exists := result[v.Author]
		if !exists {
			result[v.Author] = 0
		}
		result[v.Author]++
	}

	return &result, ""
}
