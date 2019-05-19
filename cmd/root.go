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

var doExit func(code int) = os.Exit
var fatalError = fatalIfError

func repoWorkUnit(fn repoOperation, conf Configuration, repo Repo, status *StatusList) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		fn(conf, repo, status)
		return nil, nil
	}
}

func repoLoop(fn repoOperation, msg string) {
	conf := loadConfig()
	var status StatusList
	var p pool.Pool
	if conf.Concurrency != 0 && !rootCmd.Flags().Lookup("concurrency").Changed {
		p = pool.NewLimited(conf.Concurrency)
		fmt.Println("Worker pool:", conf.Concurrency)
	} else {
		p = pool.NewLimited(rf.Concurrency)
		fmt.Println("Worker pool:", rf.Concurrency)
	}
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
		doExit(255)
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

func init() {
	rootCmd.Version = Version
	var con = uint(runtime.NumCPU() * 2)
	if con == 0 {
		con = 2
	}

	rootCmd.PersistentFlags().UintVarP(
		&rf.Concurrency,
		"concurrency",
		"c",
		uint(con),
		"Concurrency for repository jobs")
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	fatalIfError(err)
}
