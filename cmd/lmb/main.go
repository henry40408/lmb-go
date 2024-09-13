package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/henry40408/lmb/internal/database"
	"github.com/henry40408/lmb/internal/executor"
	"github.com/spf13/cobra"
)

var version string

func main() {
	var filePath string
	var timeout string

	rootCmd := &cobra.Command{Version: version}

	evalCmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate a Lua script file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var state sync.Map

			// TODO open database from file system
			db, err := database.OpenDB(":memory:")
			if err != nil {
				return err
			}
			e := executor.NewExecutor(db)

			parsedTimeout, err := time.ParseDuration(timeout)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(parsedTimeout.Seconds()))
			defer cancel()
			res, err := e.EvalFile(ctx, filePath, &state, db)
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
	evalCmd.Flags().StringVar(&filePath, "file", "", "script path")
	evalCmd.MarkFlagRequired("file")

	rootCmd.Flags().StringVar(&timeout, "timeout", "30s", "timeout in duration format e.g. 30s, 1m30s, 90s")
	rootCmd.AddCommand(evalCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
