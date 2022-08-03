package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "mds",
	Short: "hypothetical ecommerce platform.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get commands")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
