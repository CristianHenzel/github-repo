package cmd

import (
	cobra "github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

const nonFastForwardUpdatePush = "non-fast-forward update: refs/heads/master"

func init() {
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push all repositories",
	Run: func(cmd *cobra.Command, args []string) {
		conf := loadConfig()
		var status StatusList

		for i, repo := range conf.Repos {
			status.append(repo.Dir)
			status.info("Pushing", conf.Repos)

			repository, err := git.PlainOpen(repo.Dir)
			if err == git.ErrRepositoryNotExists {
				status[i].appendRed("Absent")
				continue
			}
			fatalIfError(err)

			err = repository.Push(&git.PushOptions{})

			if err == git.ErrNonFastForwardUpdate ||
				err.Error() == nonFastForwardUpdatePush {
				status[i].appendRed("Non-fast-forward update")
				continue
			}

			if err != git.NoErrAlreadyUpToDate {
				fatalIfError(err)
			}

			status[i].appendGreen("OK")
		}

		status.print()
	},
}
