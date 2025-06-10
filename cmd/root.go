/*
Copyright © 2025 Behn Hayhoe hello@behn.dev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "actions-toolkit",
	Short: "Tool to help with managing GitHub Actions",
	Long:  "A tool to help with managing pinning, caching, and other GitHub Actions related tasks.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	flags := rootCmd.PersistentFlags()
	flags.Bool("debug", false, "Enable debug logging")
	flags.String("token", "", "GitHub token to use for authentication")
	flags.BoolP("write", "w", false, "Write changes to file(s)")

	rootCmd.SetVersionTemplate("{{.Name}} version {{.Version}}+" + gitCommit + "\n")
}
