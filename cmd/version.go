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
