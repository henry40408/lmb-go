package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/henry40408/lmb/internal/eval_context"
	"github.com/henry40408/lmb/internal/store"
	"github.com/spf13/cobra"
)

func init() {
	checkSyntaxCmd.Flags().StringVar(&scriptPath, "file", "", "Script path (use '-' for stdin)")
	checkSyntaxCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(checkSyntaxCmd)
}

var (
	checkSyntaxCmd = &cobra.Command{
		Use:   "check-syntax",
		Short: "Check syntax of Lua script",
		Long:  "Check syntax of Lua script",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, _ := store.NewStore(":memory:")
			e := eval_context.NewEvalContext(store, strings.NewReader(""), http.DefaultClient)

			var reader io.Reader
			if scriptPath == "-" {
				reader = os.Stdin
			} else {
				file, err := os.Open(scriptPath)
				if err != nil {
					return err
				}
				defer file.Close()
				reader = file
			}
			_, err := e.Parse(reader, scriptPath)
			if err != nil {
				return err
			}
			return nil
		},
	}
)
