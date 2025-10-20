package historical_code_storage

import (
	"github.com/go-git/go-git/v5"
)

type GitSettings struct {
	pathToDirectory string
}

type GitStorage struct {
	settings GitSettings
	repo     *git.Repository
	root     *GitNode
}

func NewGitStorage(dir string) (HistoricalCodeStorage, error) {
	return &GitStorage{
		settings: GitSettings{dir},
	}, nil
}

func (g *GitStorage) SetUp() error {
	repo, err := git.PlainOpen(g.settings.pathToDirectory)

	if err != nil {
		return err
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
