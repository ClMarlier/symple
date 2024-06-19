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

type LoggerConfig struct {
	Log *slog.Logger
}

func (lg *LoggerConfig) sLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		user := r.Context().Value("user")

		next.ServeHTTP(w, r)

		var body map[string]interface{}
		if r.Header.Get("Content-Type") == "application/json" {
			data, err := io.ReadAll(r.Body)
			if err == nil {
				json.Unmarshal(data, &body)
			}
		}

		tracker, ok := w.(*trackWriter)
		if ok {
			if tracker.error == nil {
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
					"error", tracker.error,
				)
			}
		}
	})
}
