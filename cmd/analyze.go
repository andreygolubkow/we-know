package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/andreygolubkow/we-know/internal/analyzer"
	"github.com/spf13/cobra"
)

func analyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze-git",
		Short: "Analyze git history and extract feature commits",
		Long:  "Traverses git history, identifies commits with feature IDs, and extracts changed files.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if appConfig == nil {
				return fmt.Errorf("config is not initialized")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
			defer cancel()

			// 8 воркеров можно вынести в конфиг или флаг
			results, err := analyzer.AnalyzeRepo(ctx, *appConfig, 8)
			if err != nil {
				return err
			}

			for _, r := range results {
				fmt.Printf("Feature=%s Commit=%s\n", r.FeatureId, r.Hashtag)
				for _, f := range r.FilesChanged {
					fmt.Printf("  %s\n", f)
				}
				err := store.AppendAnalysis(r.FeatureId, appConfig.ProjectName, r.FilesChanged)
				if err != nil {
					_ = fmt.Errorf("error saving analysis: %v", err)
				}
			}
			store.Close()

			// тут можешь сохранить results → store.SaveGitAnalysis()
			// store.DB.InsertGitCommit(...)

			return nil
		},
	}

	return cmd
}
