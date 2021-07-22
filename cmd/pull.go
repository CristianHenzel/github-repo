package cmd

import (
	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	cobra "github.com/spf13/cobra"
)

func init() {
	pullCmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runPull, "Pulling")
		},
	}

	rootCmd.AddCommand(pullCmd)
}

func updateRepoConfig(conf *Configuration, repository *git.Repository) {
	repoConf, err := repository.Config()
	fatalIfError(err)

	section := repoConf.Raw.Section("user")
	section.SetOption("name", conf.Fullname)
	section.SetOption("email", conf.Email)

	err = repoConf.Validate()
	fatalIfError(err)

	err = repository.Storer.SetConfig(repoConf)
	fatalIfError(err)
}

func pullSubmodule(submodule *git.Submodule) error {
	status, err := submodule.Status()
	if err != nil {
		return err
	}

	repository, err := submodule.Repository()
	if err != nil {
		return err
	}

	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}

	if status.Branch == "" {
		remote, err := repository.Remote(git.DefaultRemoteName)
		if err != nil {
			return err
		}

		remoteRefs, err := remote.List(&git.ListOptions{})
		if err != nil {
			return err
		}

		for _, v := range remoteRefs {
			if v.Name() == "HEAD" && v.Target() != "" {
				branchRef := v.Target()
				err := repository.Fetch(&git.FetchOptions{
					RefSpecs: []gitconfig.RefSpec{"refs/*:refs/*"},
				})

				if err != nil {
					return err
				}

				err = repository.CreateBranch(&gitconfig.Branch{
					Name:   branchRef.Short(),
					Remote: git.DefaultRemoteName,
					Merge:  branchRef,
				})
				if err != nil {
					return err
				}

				err = worktree.Checkout(&git.CheckoutOptions{
					Branch: branchRef,
				})
				if err != nil {
					return err
				}
			}
		}
	}

	err = worktree.Pull(&git.PullOptions{})

	if err == git.NoErrAlreadyUpToDate {
		// Ignore NoErrAlreadyUpToDate
		err = nil
	}

	return err
}

func runPull(conf *Configuration, repo Repo, status *StatusList) {
	var repository *git.Repository
	var workTree *git.Worktree
	var err error

	if pathExists(repo.Dir) {
		repository, err = git.PlainOpen(repo.Dir)
		// If we get ErrRepositoryNotExists here, it means the repo is broken
		if err == git.ErrRepositoryNotExists {
			status.append(repo.Dir, color.RedString("Broken"))
			return
		}

		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}

		err = workTree.Pull(&git.PullOptions{
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})

		if err == git.ErrNonFastForwardUpdate {
			status.append(repo.Dir, color.RedString("Non-fast-forward update"))
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
	} else {
		repository, err = git.PlainClone(repo.Dir, false, &git.CloneOptions{
			URL:               repo.URL,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		})
		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}

		workTree, err = repository.Worktree()
		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}
	}

	submodules, err := workTree.Submodules()
	if err != nil {
		status.appendError(repo.Dir, err)
		return
	}

	for _, s := range submodules {
		err := pullSubmodule(s)
		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}
	}

	updateRepoConfig(conf, repository)
	_, err = repository.Remote("upstream")

	if repo.Parent != "" && err == git.ErrRemoteNotFound {
		_, err := repository.CreateRemote(&gitconfig.RemoteConfig{
			Name: "upstream",
			URLs: []string{repo.Parent},
		})
		if err != nil {
			status.appendError(repo.Dir, err)
			return
		}
	}

	status.append(repo.Dir, color.GreenString("OK"))
}
