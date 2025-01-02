package symple

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

func (rs *routerState) WithZeroLog(w io.Writer) routerOption {
	return func(rb *routerBuilder) error {
		logger := zerolog.New(w) //.Output(zerolog.ConsoleWriter{Out: w})
		rb.middlewareStack = append(rb.middlewareStack, loggerMiddleware(&logger))
		return nil
	}
}

type LoggerConfig struct {
	Log *slog.Logger
}

func loggerMiddleware(logger *zerolog.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
			// user := r.Context().Value(tokenSub{})
			// body := getRequestBody(r)

			err := next(w, r)
			var event *zerolog.Event
			if err != nil {
				event = logger.Error().Str("desc", err.Error())
			} else {
				event = logger.Info()
			}
			event.
				Str("verb", r.Method).
				Str("path", r.URL.Path).
				Str("time", time.Since(start).String()).
				// Any("user", user).
				// Any("body", body). // Only on internal server error to reduce allocs
				Send()
			return err
		}
	}
}

func (lg *LoggerConfig) loggerMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		user := r.Context().Value(tokenSub{})

		next(w, r)

		tracker, ok := w.(*trackWriter)
		if ok {
			if tracker.error == nil && lg.Log.Enabled(r.Context(), slog.LevelInfo) {
				body := getRequestBody(r)
				lg.Log.Info(
					"",
					"status", tracker.statusCode,
					"method", r.Method,
					"path", r.URL.Path,
					"body", body,
					"user", user,
					"duration", time.Since(start).String(),
				)
			} else if tracker.error != nil {
				body := getRequestBody(r)
				lg.Log.Error(
					"",
					"status", tracker.statusCode,
					"method", r.Method,
					"path", r.URL.Path,
					"body", body,
					"user", user,
					"duration", time.Since(start).String(),
					"error", tracker.error,
				)
			}
		}
	}
}

func getRequestBody(r *http.Request) map[string]interface{} {
	body := make(map[string]interface{})
	contentType := getContentType(r)
	switch contentType {
	case "application/json":
		data, err := io.ReadAll(r.Body)
		if err == nil {
			json.Unmarshal(data, &body)
		}
		return body
	case "multipart/form-data":
		if err := r.ParseMultipartForm(32 << 20); err == nil {
			for key, values := range r.MultipartForm.Value {
				if len(values) > 0 {
					body[key] = values
				}
			}
		}
		return body
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err == nil {
			for key, values := range r.Form {
				if len(values) > 0 {
					body[key] = values
				}
			}
		}
		return body
	}
	return body

}
