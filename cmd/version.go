package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Print the version number of mcp-executor`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mcp-executor version %s\n", version)
	},
}

func init() {
	// Add version command to root
	rootCmd.AddCommand(versionCmd)
}
