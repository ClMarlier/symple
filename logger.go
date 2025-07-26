package symple

import (
	"io"
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

func loggerMiddleware(logger *zerolog.Logger) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			start := time.Now()
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
				Send()
			return err
		}
	}
}
