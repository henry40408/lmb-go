package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/henry40408/lmb"
	"github.com/henry40408/lmb/internal/executor"
	"github.com/henry40408/lmb/internal/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var version string

func init() {
	zerolog.MessageFieldName = "msg"

	level := lmb.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	zerolog.SetGlobalLevel(level)
}

func main() {
	var debug bool
	var filePath string
	var timeout string

	rootCmd := &cobra.Command{Version: version}

	evalCmd := &cobra.Command{
		Use:   "eval",
		Short: "Evaluate a Lua script file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var state sync.Map

			// TODO open database from file system
			store, err := store.NewStore(":memory:")
			if err != nil {
				return err
			}
			e := executor.NewExecutor(&store)

			parsedTimeout, err := time.ParseDuration(timeout)
			if err != nil {
				return err
			}

			timeoutDur := time.Duration(parsedTimeout.Seconds()) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
			defer cancel()

			evalLogger := log.With().Str("file_path", filePath).Logger()
			start := time.Now()

			res, err := e.EvalFile(ctx, filePath, &state, &store)

			duration := time.Since(start)
			evalLogger.Debug().Dur("duration", duration).Msg("file evaluated")

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

	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug mode")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "30s", "timeout in duration format e.g. 30s, 1m30s, 90s")
	rootCmd.AddCommand(evalCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
