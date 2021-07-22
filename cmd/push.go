package cmd

import (
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	cobra "github.com/spf13/cobra"
)

const nonFastForwardUpdatePush = "non-fast-forward update: refs/heads/master"

func init() {
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Push all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runPush, "Pushing")
		},
	}

	rootCmd.AddCommand(pushCmd)
}

func runPush(conf *Configuration, repo Repo, status *StatusList) {
	repository, err := git.PlainOpen(repo.Dir)
	if err == git.ErrRepositoryNotExists {
		status.append(repo.Dir, color.RedString("Absent"))
		return
	}

	if err != nil {
		status.appendError(repo.Dir, err)
		return
	}

	err = repository.Push(&git.PushOptions{})

	if err == git.ErrNonFastForwardUpdate ||
		err.Error() == nonFastForwardUpdatePush {
		status.append(repo.Dir, color.RedString("Non-fast-forward update"))
		return
	}

	if err.Error() == errAuthRequired || err.Error() == errAuthFailed {
		status.append(repo.Dir, color.RedString("Unauthorized"))
		return
	}

	if err == git.NoErrAlreadyUpToDate {
		// Ignore NoErrAlreadyUpToDate
		err = nil
	}

	if err != nil {
		status.appendError(repo.Dir, err)
		return
	}

	status.append(repo.Dir, color.GreenString("OK"))
}
