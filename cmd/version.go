/*
Copyright © 2025 Behn Hayhoe hello@behn.dev
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version     = "0.0.0-SNAPSHOT"
	gitCommit   = "ffffff"
	projectName = "actions-toolkit"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  "Print the version information of the actions-toolkit CLI",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s+%s\n", projectName, version, gitCommit)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
