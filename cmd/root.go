package cmd

import (
	"fmt"
	"os"
	"runtime"

	cobra "github.com/spf13/cobra"
	term "golang.org/x/crypto/ssh/terminal"
	pool "gopkg.in/go-playground/pool.v3"
)

type rootFlags struct {
	Concurrency uint
}

type repoOperation func(Configuration, Repo, *StatusList)

func repoWorkUnit(fn repoOperation, conf Configuration, repo Repo, status *StatusList) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		fn(conf, repo, status)
		return nil, nil
	}
}

func repoLoop(fn repoOperation, msg string) {
	conf := loadConfig()
	var status StatusList
	p := pool.NewLimited(rf.Concurrency)
	defer p.Close()
	batch := p.Batch()

	go func() {
		for _, repo := range conf.Repos {
			batch.Queue(repoWorkUnit(fn, conf, repo, &status))
		}
		batch.QueueComplete()
	}()

	fmt.Printf("\r%s (0/%d)...", msg, len(conf.Repos))

	i := 1
	for range batch.Results() {
		if term.IsTerminal(int(os.Stdout.Fd())) {
			fmt.Printf("\r%s (%d/%d)...", msg, i, len(conf.Repos))
		}
		i++
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

var rf rootFlags

// Execute executes the root command.
func Execute() {
	rootCmd.Version = Version

	rootCmd.PersistentFlags().UintVarP(
		&rf.Concurrency,
		"concurrency",
		"c",
		uint(runtime.NumCPU()*2),
		"Concurrency for repository jobs")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
