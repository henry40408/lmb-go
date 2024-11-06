package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/henry40408/lmb/internal/eval_context"
	"github.com/henry40408/lmb/internal/store"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type responseRecorder struct {
	http.ResponseWriter
	Size       int
	StatusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.Size += size
	return size, err
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func init() {
	serveCmd.Flags().StringVar(&bind, "bind", "127.0.0.1:3000", "Bind")
	serveCmd.Flags().StringVar(&scriptPath, "file", "", "Script path (use '-' for stdin)")
	rootCmd.AddCommand(serveCmd)
}

func setHeadersFromState(w http.ResponseWriter, state *sync.Map) {
	rawHeaders, ok := state.Load("headers")
	if !ok {
		return
	}

	headers, ok := rawHeaders.(map[string]interface{})
	if !ok {
		return
	}

	for name, rawValue := range headers {
		setHeader(w, name, rawValue)
	}
}

func setHeader(w http.ResponseWriter, name string, value interface{}) {
	switch v := value.(type) {
	case string:
		w.Header().Set(name, v)
	case float64:
		w.Header().Set(name, strconv.FormatFloat(v, 'f', -1, 64))
	case []interface{}:
		for _, item := range v {
			switch typedItem := item.(type) {
			case string:
				w.Header().Add(name, typedItem)
			case float64:
				w.Header().Add(name, strconv.FormatFloat(typedItem, 'f', -1, 64))
			}
		}
	}
}

func setStatusCode(w http.ResponseWriter, state *sync.Map) {
	const defaultStatusCode = http.StatusOK

	rawStatusCode, ok := state.Load("status_code")
	if !ok {
		w.WriteHeader(defaultStatusCode)
		return
	}

	statusCode := defaultStatusCode
	switch code := rawStatusCode.(type) {
	case int:
		statusCode = code
	case float64:
		statusCode = int(code)
	case string:
		if parsedCode, err := strconv.Atoi(code); err == nil {
			statusCode = parsedCode
		}
	}

	if statusCode < 100 || statusCode > 599 {
		statusCode = defaultStatusCode
	}

	w.WriteHeader(statusCode)
}

var (
	bind     string
	serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Process HTTP requests with Lua script",
		Long:  "Process HTTP requests with Lua script",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := store.NewStore(storePath)
			if err != nil {
				return err
			}
			parsedHttpTimeout, err := time.ParseDuration(httpTimeout)
			if err != nil {
				return err
			}
			httpClient := http.Client{Timeout: parsedHttpTimeout}
			e := eval_context.NewEvalContext(store, os.Stdin, &httpClient)

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

			compiled, err := e.Compile(reader, scriptPath)
			if err != nil {
				return err
			}

			handler := func(w http.ResponseWriter, r *http.Request) {
				var state sync.Map

				requestState := make(map[string]interface{})
				requestHeaders := make(map[string]interface{})
				for key, values := range r.Header {
					requestHeaders[strings.Map(unicode.ToLower, key)] = values
				}
				requestState["headers"] = requestHeaders
				requestState["path"] = r.URL.Path
				requestState["method"] = r.Method
				state.Store("request", requestState)

				ctx, cancel, err := setupTimeoutContext(timeout)
				if err != nil {
					log.Error().Err(err).Msg("failed to set timeout")
					http.Error(w, "", http.StatusInternalServerError)
					return
				}
				defer cancel()

				var buf bytes.Buffer
				res, err := e.Eval(ctx, compiled, &state, &buf)
				if err != nil {
					log.Error().Err(err).Msg("request errored")
					http.Error(w, "", http.StatusInternalServerError)
					return
				}

				setHeadersFromState(w, &state)
				setStatusCode(w, &state)
				if buf.Len() > 0 {
					if res != nil {
						log.Warn().Msg("result will be ignored because buffer is not empty")
					}
					_, err := io.Copy(w, &buf)
					if err != nil {
						log.Error().Err(err).Msg("request errored")
						http.Error(w, "", http.StatusInternalServerError)
						return
					}
				} else {
					fmt.Fprintf(w, "%v", res)
				}
			}

			server := &http.Server{
				Addr: bind,
				Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					recorder := &responseRecorder{
						ResponseWriter: w,
						StatusCode:     http.StatusOK,
					}

					begin := time.Now()
					handler(recorder, r)
					duration := time.Since(begin)

					if e := log.Debug(); e.Enabled() {
						logged := log.Debug().
							Int("size", recorder.Size).
							Int("status", recorder.StatusCode).
							Str("method", r.Method).
							Str("query", r.URL.RawQuery).
							Stringer("duration", duration).
							Stringer("url", r.URL)
						loggedHeaders := zerolog.Dict()
						for key, values := range r.Header {
							loggedHeaders = loggedHeaders.Strs(key, values)
						}
						logged.Dict("headers", loggedHeaders).Msg("request completed")
					}
				}),
			}
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		}}
)
