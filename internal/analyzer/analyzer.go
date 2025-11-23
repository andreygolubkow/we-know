package analyzer

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/andreygolubkow/we-know/internal/config"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type CommitAnalysis struct {
	Hashtag      string
	FeatureId    string
	FilesChanged []string
}

func AnalyzeRepo(ctx context.Context, cfg config.Config, workers int) ([]CommitAnalysis, error) {
	if workers <= 0 {
		workers = 4
	}

	// Базовый репозиторий – только для прохода по истории (однопоточно)
	repo, err := git.PlainOpen(cfg.RepoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	iter, err := repo.Log(&git.LogOptions{
		Order: git.LogOrderCommitterTime,
		All:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}
	defer iter.Close()

	// Ищем featureId в сообщении коммита
	pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(cfg.IssuePrefix) + `\d+\b`)

	// --- Шаг 1: однопоточно собираем задачи (hash + featureID) ---

	type task struct {
		hash      plumbing.Hash
		featureID string
	}

	var tasks []task

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		c, err := iter.Next()
		if err != nil {
			// io.EOF – конец истории
			break
		}

		// Быстрый фильтр по префиксу
		if cfg.IssuePrefix != "" && !containsPrefix(c.Message, cfg.IssuePrefix) {
			continue
		}

		match := pattern.FindString(c.Message)
		if match == "" {
			// если нет featureId – пропускаем коммит
			continue
		}

		tasks = append(tasks, task{
			hash:      c.Hash,
			featureID: match,
		})
	}

	// --- Шаг 2: многопоточно обрабатываем задачи ---

	tasksCh := make(chan task, workers*2)
	resultsCh := make(chan CommitAnalysis, workers*2)

	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Каждый воркер открывает свой экземпляр репозитория
			r, err := git.PlainOpen(cfg.RepoPath)
			if err != nil {
				fmt.Printf("error opening repo in worker: %v\n", err)
				return
			}

			for t := range tasksCh {
				select {
				case <-ctx.Done():
					return
				default:
				}

				commit, err := r.CommitObject(t.hash)
				if err != nil {
					fmt.Printf("error getting commit %s: %v\n", t.hash.String(), err)
					continue
				}

				files, err := changedFiles(commit)
				if err != nil {
					// можно залогировать, можно молча скипать
					fmt.Printf("error getting changed files for %s: %v\n", t.hash.String(), err)
					continue
				}

				resultsCh <- CommitAnalysis{
					Hashtag:      t.hash.String(),
					FeatureId:    t.featureID,
					FilesChanged: files,
				}
			}
		}()
	}

	// Закрываем resultsCh, когда все воркеры закончат
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Кормим воркеров задачами
	go func() {
		defer close(tasksCh)
		for _, t := range tasks {
			select {
			case <-ctx.Done():
				return
			case tasksCh <- t:
			}
		}
	}()

	// Собираем результаты
	var all []CommitAnalysis
	for r := range resultsCh {
		all = append(all, r)
	}

	return all, nil
}

func containsPrefix(msg, prefix string) bool {
	return strings.Contains(msg, prefix)
}

func changedFiles(c *object.Commit) ([]string, error) {
	// if its first commit - return all files
	if c.NumParents() == 0 {
		tree, err := c.Tree()
		if err != nil {
			return nil, fmt.Errorf("get tree: %w", err)
		}

		var files []string
		err = tree.Files().ForEach(func(f *object.File) error {
			files = append(files, f.Name)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("iterate files: %w", err)
		}

		return files, nil
	}

	// take only first parent
	parent, err := c.Parents().Next()
	if err != nil {
		return nil, fmt.Errorf("get parent: %w", err)
	}

	parentTree, err := parent.Tree()
	if err != nil {
		return nil, fmt.Errorf("get parent tree: %w", err)
	}

	tree, err := c.Tree()
	if err != nil {
		return nil, fmt.Errorf("get commit tree: %w", err)
	}

	patch, err := parentTree.Diff(tree)
	if err != nil {
		return nil, fmt.Errorf("diff trees: %w", err)
	}

	seen := make(map[string]struct{})
	var files []string

	for _, fp := range patch {
		from := fp.From
		to := fp.To

		if from.Name != "" {
			if _, ok := seen[from.Name]; !ok {
				seen[from.Name] = struct{}{}
				files = append(files, from.Name)
			}
		}
		if to.Name != "" {
			if _, ok := seen[to.Name]; !ok {
				seen[to.Name] = struct{}{}
				files = append(files, to.Name)
			}
		}
	}

	return files, nil
}
