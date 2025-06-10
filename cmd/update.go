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
	"github.com/behnh/actions-toolkit/internal/file"
	"github.com/behnh/actions-toolkit/internal/processor"
	"github.com/spf13/cobra"
	"log/slog"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update GitHub Actions to their latest versions",
	Long: `Update GitHub Actions to their latest versions in workflow files.
You can specify a specific action to update, or update all actions in a file or directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		actionName, _ := cmd.Flags().GetString("action")
		dirPath, _ := cmd.Flags().GetString("dir")
		filePath, _ := cmd.Flags().GetString("file")
		write, _ := cmd.Flags().GetBool("write")
		token, _ := cmd.Flags().GetString("token")

		if actionName == "" {
			slog.Error("Action name is required")
			return
		}

		var filesToProcess []string
		var err error

		if filePath != "" {
			filesToProcess = []string{filePath}
		} else if dirPath != "" {
			filesToProcess, err = file.GetYAMLFiles(dirPath)
			if err != nil {
				slog.Error("Failed to get YAML files", "error", err)
				return
			}
		} else {
			slog.Error("Either --dir or --file must be specified")
			return
		}

		for _, f := range filesToProcess {
			processor.UpdateAction(f, actionName, token, write)
		}
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Add flags specific to the update command
	updateCmd.Flags().String("action", "", "Action name to update (required)")
	updateCmd.Flags().String("dir", "", "Directory containing workflow files")
	updateCmd.Flags().String("file", "", "Specific workflow file to update")
	updateCmd.Flags().BoolP("write", "w", false, "Write changes to files (default is dry run)")

	// Mark action as required
	updateCmd.MarkFlagRequired("action")
}
