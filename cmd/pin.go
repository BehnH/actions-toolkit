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

var pinCmd = &cobra.Command{
	Use:   "pin",
	Short: "Pin GitHub Actions to a specific version using release commit SHAs",
	Long: `Pin GitHub Actions to a specific version using release commit SHAs. This satisfies GitHub's recommended best practices for Actions security, as detailed here:
https://docs.github.com/en/actions/security-for-github-actions/security-guides/security-hardening-for-github-actions#using-third-party-actions`,
	Example: `  # Pin all actions to the latest release in a directory
  actions-toolkit pin --all --dir .github/workflows --write

  # Pin a specific action to a specific version in a file
  actions-toolkit pin --action actions/checkout --version v4.2.2 --file .github/workflows/lint.yaml --write

  # Pin a specific action to a version in a directory
  actions-toolkit pin --action actions/checkout --version v4.2.2 --dir .github/workflows --write
`,
	Run: func(cmd *cobra.Command, args []string) {
		actionName, _ := cmd.Flags().GetString("action")
		version, _ := cmd.Flags().GetString("version")
		all, _ := cmd.Flags().GetBool("all")
		dirPath, _ := cmd.Flags().GetString("dir")
		filePath, _ := cmd.Flags().GetString("file")
		write, _ := cmd.Flags().GetBool("write")
		token, _ := cmd.Flags().GetString("token")

		if all && (actionName != "" || version != "") {
			slog.Error("Cannot specify both --all and --action or --version")
			return
		}

		if !all && (actionName == "" || version == "") {
			slog.Error("Must specify --action and --version when --all is not provided")
			return
		}

		if dirPath != "" && filePath != "" {
			slog.Error("Cannot specify both --dir and --file")
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

		// result := processor.FindActionsInFiles(filesToProcess)
		// slog.Info("Found the following actions to pin", "actions", result, "count", len(result))

		if all {
			processor.PinAllActions(filesToProcess, token, write)
		} else {
			processor.PinAction(filesToProcess, actionName, version, token, write)
		}

	},
}

func init() {
	rootCmd.AddCommand(pinCmd)

	pinCmd.Flags().String("action", "", "Action name to pin (required if --all is not specified)")
	pinCmd.Flags().String("version", "", "Version to pin to (required if --all is not specified)")
	pinCmd.Flags().BoolP("all", "a", false, "Pin all actions to the latest release")
	pinCmd.Flags().String("dir", "", "Directory containing workflow files")
	pinCmd.Flags().String("file", "", "Specific workflow file to pin")
	pinCmd.Flags().BoolP("write", "w", false, "Write changes to files (default is dry run)")
}
