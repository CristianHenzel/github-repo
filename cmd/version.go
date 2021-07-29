package cmd

import (
	"fmt"

	semver "github.com/blang/semver"
	selfupdate "github.com/rhysd/go-github-selfupdate/selfupdate"
	cobra "github.com/spf13/cobra"
)

// Version holds the application version.
// It gets filled automatically at build time.
var Version string

// BuildDate holds the date and time at which the application was build.
// It gets filled automatically at build time.
var BuildDate string

const updateRepo = "CristianHenzel/github-repo"

func init() {
	var update bool

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			if update {
				selfUpdate()
			} else {
				printVersion()
			}
		},
	}

	versionCmd.Flags().BoolVarP(&update, "update", "u", false, "Update application")
	rootCmd.AddCommand(versionCmd)
}

func printVersion() {
	var vSuffix string

	current := semver.MustParse(Version)
	latest, found, err := selfupdate.DetectLatest(updateRepo)
	fatalIfError(err)

	if !found || latest.Version.LTE(current) {
		vSuffix = "(latest)"
	} else {
		vSuffix = "(newer version available: " + latest.Version.String() + ")"
	}

	fmt.Println("gr version:", Version, vSuffix)
	fmt.Println("Built at:", BuildDate)
}

func selfUpdate() {
	current := semver.MustParse(Version)
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Validator: &selfupdate.SHA2Validator{},
	})
	fatalIfError(err)

	latest, err := updater.UpdateSelf(current, updateRepo)
	fatalIfError(err)

	if latest.Version.LTE(current) {
		fmt.Println("You are already using the latest version:", Version)
	} else {
		fmt.Println("Successfully updated to version", latest.Version)
	}
}
