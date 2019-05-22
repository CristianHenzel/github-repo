package cmd

import (
	"flag"
	"fmt"
	"math"
	"os"
	"runtime"

	cobra "github.com/spf13/cobra"
	term "golang.org/x/crypto/ssh/terminal"
	pool "gopkg.in/go-playground/pool.v3"
)

var (
	errConfExists = fmt.Errorf("Configuration file already exists in current directory. "+
		"Please run 'update' if you want to update your settings. "+
		"Alternatively, remove %s if you want to initialize the repository again.", configFile)
	errConfNotExists = fmt.Errorf("Couldn't find configuration file. Make sure that you are in the base " +
		"directory and that init has been run successfully.")
)

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

	con, _ := rootCmd.PersistentFlags().GetUint("concurrency")
	if conf.Concurrency > 0 && con == 0 {
		p = pool.NewLimited(conf.Concurrency)
	} else if con > 0 {
		p = pool.NewLimited(con)
	} else {
		var con = float64(runtime.NumCPU() * 2)
		con = math.Max(con, 4)
		p = pool.NewLimited(uint(con))
	}
	defer p.Close()
	batch := p.Batch()

	go func() {
		for _, repo := range conf.Repos {
			batch.Queue(repoWorkUnit(fn, conf, repo, &status))
		}
		batch.QueueComplete()
	}()

	if term.IsTerminal(int(os.Stdout.Fd())) || flag.Lookup("test.v") != nil {
		fmt.Printf("\r%s (0/%d)...", msg, len(conf.Repos))

		i := 1
		for range batch.Results() {
			fmt.Printf("\r%s (%d/%d)...", msg, i, len(conf.Repos))
			i++
		}
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
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

var rootCmd = &cobra.Command{
	Use:   "gr",
	Short: "gr is a github repository management tool",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		fatalIfError(err)
	},
}

func init() {
	rootCmd.Version = Version
	var tmp uint

	rootCmd.PersistentFlags().UintVarP(
		&tmp,
		"concurrency",
		"c",
		0,
		"Concurrency for repository jobs")
}

// Execute executes the root command.
func Execute() {
	err := rootCmd.Execute()
	fatalIfError(err)
}
