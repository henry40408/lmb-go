package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/henry40408/lmb/internal/eval_context"
	"github.com/henry40408/lmb/internal/store"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func init() {
	evalCmd.Flags().StringVar(&filePath, "file", "", "Script path (use '-' for stdin)")
}

var (
	filePath string
	evalCmd  = &cobra.Command{
		Use:   "eval",
		Short: "Evaluate a file",
		Long:  "Evaluate a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			var state sync.Map

			store, err := store.NewStore(storePath)
			if err != nil {
				return err
			}
			defer store.Close()
			e := eval_context.NewEvalContext(&store, os.Stdin)

			parsedTimeout, err := time.ParseDuration(timeout)
			if err != nil {
				return err
			}

			timeoutDur := time.Duration(parsedTimeout.Seconds()) * time.Second
			ctx, cancel := context.WithTimeout(context.Background(), timeoutDur)
			defer cancel()

			var reader io.Reader
			if filePath == "-" {
				reader = os.Stdin
			} else {
				file, err := os.Open(filePath)
				if err != nil {
					return err
				}
				defer file.Close()
				reader = file
			}

			evalLogger := log.With().Str("file_path", filePath).Logger()
			start := time.Now()

			compiled, err := e.Compile(reader, filePath)
			if err != nil {
				return err
			}
			res, err := e.Eval(ctx, compiled, &state)

			duration := time.Since(start)
			evalLogger.Debug().Str("duration", duration.String()).Msg("file evaluated")

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
)
