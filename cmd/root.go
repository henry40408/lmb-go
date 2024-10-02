package cmd

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	debug      bool
	storePath  string
	scriptPath string
	timeout    string
	rootCmd    = &cobra.Command{
		Use:   "lmb",
		Short: "A Lua function runner",
		Long:  `Lmb is a Lua function runner`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			zerolog.MessageFieldName = "msg"

			level := parseLogLevel(os.Getenv("LOG_LEVEL"))
			zerolog.SetGlobalLevel(level)
			if debug {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Debug")
	rootCmd.PersistentFlags().StringVar(&storePath, "db-path", "db.sqlite3", "Path to store file")
	rootCmd.PersistentFlags().StringVar(&timeout, "timeout", "30s", "Timeout in human-readable format e.g. 30s, 1m30s")
}

func Execute() error {
	return rootCmd.Execute()
}
