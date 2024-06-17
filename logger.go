package symple

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func WithStructLogger(sl slog.Handler) muxOption {
	return func(rb *routerBuilder) error {
		config := &LoggerConfig{
			Log: slog.New(sl),
		}
		rb.middlewareStack = append(rb.middlewareStack, config.sLogger)
		return nil
	}
}

type trackWriter struct {
	http.ResponseWriter
	statusCode int
}

func (tr *trackWriter) WriteHeader(statusCode int) {
	tr.ResponseWriter.WriteHeader(statusCode)
	tr.statusCode = statusCode
}

type LoggerConfig struct {
	Log *slog.Logger
}

func (lg *LoggerConfig) sLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		user := r.Context().Value("user")

		tracker := &trackWriter{w, http.StatusOK}
		next.ServeHTTP(tracker, r)

		var body map[string]interface{}
		if r.Header.Get("Content-Type") == "application/json" {
			data, err := io.ReadAll(r.Body)
			if err == nil {
				json.Unmarshal(data, &body)
			}
		}

		error, ok := r.Context().Value("error").(error)
		if !ok {
			lg.Log.Info(
				"",
				"status", tracker.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"body", body,
				"user", user,
				"time", time.Since(start).String(),
			)
		} else {
			lg.Log.Error(
				"",
				"status", tracker.statusCode,
				"method", r.Method,
				"path", r.URL.Path,
				"body", body,
				"user", user,
				"time", time.Since(start).String(),
				"error", error,
			)
		}
	})
}
