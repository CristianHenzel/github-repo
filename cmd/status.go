package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	color "github.com/fatih/color"
	cobra "github.com/spf13/cobra"
	term "golang.org/x/crypto/ssh/terminal"
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

func (status *Status) appendGreen(s string) {
	status.State = status.State + "\t" + color.GreenString(s)
}

func (status *Status) appendRed(s string) {
	status.State = status.State + "\t" + color.RedString(s)
}

func (status *Status) toString() string {
	return status.Repo + "\t" + status.State
}

func (statuslist *StatusList) append(repo string) {
	*statuslist = append(*statuslist, Status{Repo: repo})
}

func (statuslist *StatusList) info(msg string, repos []Repo) {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		fmt.Printf("\r%s (%d/%d)...", msg, len(*statuslist), len(repos))
	}
}

func (statuslist *StatusList) print() {
	// Reset
	fmt.Println()

	w := tabwriter.NewWriter(os.Stdout, 5, 0, 5, space, 0)
	for _, v := range *statuslist {
		fmt.Fprintln(w, v.toString())
	}
	w.Flush()
}

func init() {
	var statusCmd = &cobra.Command{
		Use:   "status",
		Short: "Show status for all repositories",
		Run: func(cmd *cobra.Command, args []string) {
			runStatus()
		},
	}

	rootCmd.AddCommand(statusCmd)
}

func runStatus() {
	conf := loadConfig()
	var status StatusList

	for i, repo := range conf.Repos {
		status.append(repo.Dir)
		status.info("Checking", conf.Repos)

		if !pathExists(repo.Dir) {
			status[i].appendRed("Absent")
			continue
		}

		repository, err := git.PlainOpen(repo.Dir)
		fatalIfError(err)

		head, err := repository.Head()
		fatalIfError(err)

		workTree, err := repository.Worktree()
		fatalIfError(err)

		repoStatus, err := workTree.Status()
		fatalIfError(err)

		if repoStatus.IsClean() {
			status[i].appendGreen("Clean")
		} else {
			status[i].appendRed("Dirty")
		}

		remote, err := repository.Remote("origin")
		fatalIfError(err)
		remoteRef, err := remote.List(&git.ListOptions{})
		fatalIfError(err)

		for _, r := range remoteRef {
			if r.Name().String() == "refs/heads/"+repo.Branch {
				if r.Hash() == head.Hash() {
					status[i].appendGreen("Fresh")
				} else {
					status[i].appendRed("Stale")
				}
				break
			}
		}
	}

	status.print()
}
