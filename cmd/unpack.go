/*
Copyright © 2026 Kod project Contributors
*/
package cmd

import (
	"context"
	"fmt"
	"kod/internal"
	"os"

	"github.com/spf13/cobra"
)

var packagePath string

// unpackCmd represents the unpack command
var unpackCmd = &cobra.Command{
	Use:   "unpack",
	Short: "Unpacks a kodpkg to a temp folder and returns its location",
	Long: `Unpacks a kodpkg to a temporary folder. This allows for inspection before installation. 

Unpack always recreates the temporary folder, even if it already exists. 
This is useful if you think the deploy command has not completely unpacked the package before installation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("unpack called")

		// TODO sanity check parameter values

		tmpFolder, err := internal.DoUnpack(context.Background(), packagePath, true)
		if err != nil {
			fmt.Println("Error performing unpack", err)
			os.Exit(1)
		}
		fmt.Println("Completing unpack into:", tmpFolder)
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(unpackCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// unpackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// unpackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	unpackCmd.Flags().StringVarP(&packagePath, "package", "p", "", ".kodpkg file to unpack")
}
