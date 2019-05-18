package cmd

import (
	"fmt"

	cobra "github.com/spf13/cobra"
)

// Version holds the application version
// It gets filled automatically at build time
var Version string

// BuildDate holds the date and time at which the application was build
// It gets filled automatically at build time
var BuildDate string

func init() {
	var updateCmd = &cobra.Command{
		Use:   "version",
		Short: "Display version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("gr version:", Version)
			fmt.Println("Built at:", BuildDate)
		},
	}

	rootCmd.AddCommand(updateCmd)
}
