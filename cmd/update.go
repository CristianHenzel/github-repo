package cmd

import (
	cobra "github.com/spf13/cobra"
)

func init() {
	var updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update configuration",
		Run: func(cmd *cobra.Command, args []string) {
			conf := loadConfig()
			runInit(conf, true)
		},
	}

	rootCmd.AddCommand(updateCmd)
}
