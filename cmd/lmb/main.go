package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/henry40408/lmb/internal"
	"github.com/spf13/cobra"
)

var version string

func main() {
	var filePath string

	rootCmd := &cobra.Command{Version: version}

	evalCmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate a Lua script file",
		RunE: func(cmd *cobra.Command, args []string) error {
			e := internal.NewExecutor()

			res, err := e.EvalFile(filePath)
			if err != nil {
				return err
			}

			encoded, err := json.Marshal(res)
			if err != nil {
				return err
			}

			fmt.Printf("%s", string(encoded))
			return nil
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
