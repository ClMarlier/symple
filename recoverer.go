package symple

import (
	"fmt"
	"net/http"
	"runtime"
)

type recovererConfig struct {
	writeError bool
}

type recovererOption func(*recovererConfig) error

// WithRecoverer handle gracefully any panic that could occur
func WithRecoverer(writeError bool) routerOption {
	return func(rb *routerBuilder) error {
		config := &recovererConfig{writeError: writeError}
		rb.middlewareStack = append(rb.middlewareStack, config.recoverer)
		return nil
	}
}

func (rc *recovererConfig) recoverer(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				error := fmt.Errorf("%+v", err)
				stack := make([]uintptr, 20)
				n := runtime.Callers(1, stack)
				stack = stack[:n]
				frames := runtime.CallersFrames(stack)
				frames.Next()

				for frame, more := frames.Next(); more; frame, more = frames.Next() {
					error = fmt.Errorf("%s\n%v", frame.Function, error)
				}
				if !rc.writeError {
					error = fmt.Errorf("internal server error")
				}
				ErrorResponse(w, error, http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}
