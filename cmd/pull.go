package cmd

import (
	cobra "github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

func init() {
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull all repositories",
	Run: func(cmd *cobra.Command, args []string) {
		conf := loadConfig()
		var status StatusList

		for i, repo := range conf.Repos {
			status.append(repo.Dir)
			status.info("Pulling", conf.Repos)
			var repository *git.Repository
			var err error

			if pathExists(repo.Dir) {
				repository, err = git.PlainOpen(repo.Dir)
				fatalIfError(err)

				workTree, err := repository.Worktree()
				fatalIfError(err)

				err = workTree.Pull(&git.PullOptions{RemoteName: "origin"})

				if err == git.ErrNonFastForwardUpdate {
					status[i].appendRed("Non-fast-forward update")
					continue
				}

				if err != git.NoErrAlreadyUpToDate {
					fatalIfError(err)
				}
			} else {
				repository, err = git.PlainClone(repo.Dir, false, &git.CloneOptions{URL: repo.Url})
				fatalIfError(err)
			}

			repoConf, err := repository.Config()
			fatalIfError(err)
			section := repoConf.Raw.Section("user")
			section.SetOption("name", conf.Fullname)
			section.SetOption("email", conf.Email)
			err = repoConf.Validate()
			fatalIfError(err)
			repository.Storer.SetConfig(repoConf)

			status[i].appendGreen("OK")
		}

		status.print()
	},
}
