/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version string = "unknown"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Outputs the version of kod",
	Long:  `Outputs the version of kod.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kod %s\n", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
