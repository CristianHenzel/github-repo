package cmd

import (
	color "github.com/fatih/color"
	cobra "github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

const nonFastForwardUpdatePush = "non-fast-forward update: refs/heads/master"

func init() {
	var pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runPush, "Pushing")
		},
	}

	rootCmd.AddCommand(pushCmd)
}

func runPush(conf Configuration, repo Repo) string {
	repository, err := git.PlainOpen(repo.Dir)
	if err == git.ErrRepositoryNotExists {
		return color.RedString("Absent")
	}
	fatalIfError(err)

	err = repository.Push(&git.PushOptions{})

	if err == git.ErrNonFastForwardUpdate ||
		err.Error() == nonFastForwardUpdatePush {
		return color.RedString("Non-fast-forward update")
	}

	if err != git.NoErrAlreadyUpToDate {
		fatalIfError(err)
	}

	return color.GreenString("OK")
}
