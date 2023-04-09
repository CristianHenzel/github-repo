package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	color "github.com/fatih/color"
	git "github.com/go-git/go-git/v5"
	cobra "github.com/spf13/cobra"
)

const space = byte(' ')

// Status holds a repository's status.
type Status struct {
	Repo  string
	State string
}

// StatusList is a convenience wrapper around []Status.
type StatusList []Status

func (status *Status) toString() string {
	return status.Repo + "\t" + status.State
}

func (statuslist *StatusList) appendError(repo string, err error) {
	*statuslist = append(*statuslist, Status{
		Repo:  repo,
		State: color.RedString(err.Error()),
	})
}

func (statuslist *StatusList) append(repo, state string) {
	*statuslist = append(*statuslist, Status{
		Repo:  repo,
		State: state,
	})
}

func (statuslist *StatusList) print() {
	// Sort list
	sl := *statuslist

	if len(sl) == 0 {
		return
	}

	sort.Slice(sl, func(i, j int) bool {
		return sl[i].Repo < sl[j].Repo
	})

	// Reset
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 5, 0, 5, space, 0)
	for _, v := range sl {
		_, err := fmt.Fprintln(w, v.toString())
		fatalIfError(err)
	}

	err := w.Flush()
	fatalIfError(err)
}

func init() {
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runStatus, "Checking")
			runLocalStatus()
		},
	}

	rootCmd.AddCommand(statusCmd)
}

func isRepoDir(path string, repos []Repo) bool {
	path = path + "/"
	for _, r := range repos {
		repoDir := r.Dir + "/"
		if strings.HasPrefix(repoDir, path) {
			return true
		}
	}

	return false
}

func runLocalStatus() {
	conf := loadConfig()
	var status StatusList

	files, err := filepath.Glob(conf.BaseDir + "/*")
	fatalIfError(err)

	if conf.SubDirs {
		parents, err := filepath.Glob(conf.BaseDir + "/*/*")
		fatalIfError(err)
		files = append(files, parents...)
	}

	for _, f := range files {
		if !isRepoDir(f, conf.Repos) {
			status.append(f, color.RedString("untracked"))
		}
	}

	status.print()
}

func runStatus(conf *Configuration, repo Repo, status *StatusList) {
	var ret string

	if !pathExists(repo.Dir) {
		status.append(repo.Dir, color.RedString("absent"))

		return
	}

	repository, err := git.PlainOpen(repo.Dir)
	// If we get ErrRepositoryNotExists here, it means the repo is broken
	if errors.Is(err, git.ErrRepositoryNotExists) {
		status.append(repo.Dir, color.RedString("broken"))

		return
	}

	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	head, err := repository.Head()
	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	if branch := head.Name().Short(); branch == repo.Branch {
		ret += color.GreenString(branch)
	} else {
		ret += color.RedString(branch)
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	if repoStatus.IsClean() {
		ret += "\t" + color.GreenString("clean")
	} else {
		ret += "\t" + color.RedString("dirty")
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		status.appendError(repo.Dir, err)

		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				ret += "\t" + color.GreenString("latest")
			} else {
				ret += "\t" + color.RedString("stale")
			}

			break
		}
	}

	status.append(repo.Dir, ret)
}
