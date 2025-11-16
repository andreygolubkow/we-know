package config

import "github.com/spf13/viper"

type Config struct {
	RepoPath    string            `mapstructure:"repo_path"`
	IssuePrefix string            `mapstructure:"issue_prefix"`
	Tracker     string            `mapstructure:"tracker"`
	AzureDevOps AzureDevOpsConfig `mapstructure:"azure_devops"`
}

type AzureDevOpsConfig struct {
	Organization string `mapstructure:"organization"`
	Project      string `mapstructure:"project"`
	Token        string `mapstructure:"token"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
