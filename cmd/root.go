package cmd

import (
	"fmt"
	"os"

	"github.com/andreygolubkow/we-know/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "we-know",
		Short: "Just trying to understand what is going on",
		Long:  `we-know - the tool which helps you to understand what is going on in your codebase`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags()
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"path to config file (example, ./we-know.yaml)",
	)
}

func initConfig() {
	if cfgFile != "" {
		// Явно заданный конфиг
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("we-know")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("WE_KNOW")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if ok := os.IsNotExist(err); ok {
			fmt.Println("ℹ️  Config file not found, using default values")
			return
		}
		// другие ошибки — критичные
		fmt.Fprintf(os.Stderr, "❌ Error while reading config file: %v\n", err)
		os.Exit(1)
	}

	if _, err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error whilie parsing config file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Using config: %s\n", viper.ConfigFileUsed())
}
