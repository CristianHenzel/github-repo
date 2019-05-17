package cmd

import (
	"fmt"
	"os"

	cobra "github.com/spf13/cobra"
)

type repoOperation func(Configuration, Repo) string

func repoLoop(fn repoOperation, msg string) {
	conf := loadConfig()
	var status StatusList

	for i, repo := range conf.Repos {
		status.append(repo.Dir)
		status.info(msg, conf.Repos)

		status[i].State = fn(conf, repo)
	}

	status.print()
}

func fatalIfError(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(255)
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	fatalIfError(err)
	return false
}

var rootCmd = &cobra.Command{
	Use:   "gr",
	Short: "gr is a github repository management tool",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute executes the root command.
func Execute() {
	rootCmd.Version = "1.0.0"

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
