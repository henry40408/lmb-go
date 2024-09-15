package cmd

import (
	"io"
	"os"
	"strings"

	"github.com/henry40408/lmb/internal/eval_context"
	"github.com/henry40408/lmb/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	checkSyntaxCmd.Flags().StringVar(&checkFilePath, "file", "", "Script path (use '-' for stdin)")
}

var (
	checkFilePath  string
	checkSyntaxCmd = &cobra.Command{
		Use:   "check-syntax",
		Short: "Check syntax of Lua script",
		Long:  "Check syntax of Lua script",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _ := store.NewStore(":memory:")
			e := eval_context.NewEvalContext(store, strings.NewReader(""))

			var reader io.Reader
			if checkFilePath == "-" {
				reader = os.Stdin
			} else {
				file, err := os.Open(checkFilePath)
				if err != nil {
					return err
				}
				defer file.Close()
				reader = file
			}
			_, err := e.Parse(reader, checkFilePath)
			if err != nil {
				return err
			}
			return nil
		},
	}
)
