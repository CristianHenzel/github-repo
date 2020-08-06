package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	color "github.com/fatih/color"
	cobra "github.com/spf13/cobra"
	git "gopkg.in/src-d/go-git.v4"
)

const space = byte(' ')

// Status holds a repository's status
type Status struct {
	Repo  string
	State string
}

// StatusList is a convenience wrapper around []Status
type StatusList []Status

func (status *Status) toString() string {
	return status.Repo + "\t" + status.State
}

func (statuslist *StatusList) append(repo, state string) {
	*statuslist = append(*statuslist, Status{Repo: repo, State: state})
}

func (statuslist *StatusList) print() {
	// Sort list
	sl := *statuslist
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
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			repoLoop(runStatus, "Checking")
		},
	}

	rootCmd.AddCommand(statusCmd)
}

func runStatus(conf *Configuration, repo Repo, status *StatusList) {
	var ret string

	if !pathExists(repo.Dir) {
		status.append(repo.Dir, color.RedString("Absent"))
		return
	}

	repository, err := git.PlainOpen(repo.Dir)
	// If we get ErrRepositoryNotExists here, it means the repo is broken
	if err == git.ErrRepositoryNotExists {
		status.append(repo.Dir, color.RedString("Broken"))
		return
	}
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}

	head, err := repository.Head()
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}

	workTree, err := repository.Worktree()
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}

	repoStatus, err := workTree.Status()
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}

	if repoStatus.IsClean() {
		ret += color.GreenString("Clean")
	} else {
		ret += color.RedString("Dirty")
	}

	remote, err := repository.Remote(git.DefaultRemoteName)
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}
	remoteRef, err := remote.List(&git.ListOptions{})
	if err != nil {
		status.append(repo.Dir, color.RedString("ERROR: " + err.Error()))
		return
	}

	for _, r := range remoteRef {
		if r.Name().String() == "refs/heads/"+repo.Branch {
			if r.Hash() == head.Hash() {
				ret += "\t" + color.GreenString("Fresh")
			} else {
				ret += "\t" + color.RedString("Stale")
			}
			break
		}
	}

	status.append(repo.Dir, ret)
}
