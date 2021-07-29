package cmd

import (
	"errors"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	transport "github.com/go-git/go-git/v5/plumbing/transport"
	cobra "github.com/spf13/cobra"
)

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
	if errors.Is(err, git.ErrRepositoryNotExists) {
		status.append(repo.Dir, color.RedString("absent"))

		return
	}

	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	err = repository.Push(&git.PushOptions{})

	if errors.Is(err, git.ErrNonFastForwardUpdate) {
		status.append(repo.Dir, color.RedString("non-fast-forward update"))

		return
	}

	if errors.Is(err, transport.ErrAuthenticationRequired) ||
		errors.Is(err, transport.ErrAuthorizationFailed) {
		status.append(repo.Dir, color.RedString("unauthorized"))

		return
	}

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		// Ignore NoErrAlreadyUpToDate
		err = nil
	}

	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	status.append(repo.Dir, color.GreenString("ok"))
}
