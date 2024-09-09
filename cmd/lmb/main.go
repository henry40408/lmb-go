package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var filePath string

	rootCmd := &cobra.Command{}

	evalCmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate a Lua script file",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	evalCmd.Flags().StringVar(&filePath, "file", "", "script file")
	evalCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(evalCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
