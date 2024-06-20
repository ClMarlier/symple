package symple

import (
	"fmt"
	"net/http"
	"runtime"
)

// WithRecoverer handle gracefully any panic that could occur
func WithRecoverer(rb *routerBuilder) error {
	rb.middlewareStack = append(rb.middlewareStack, recoverer)
	return nil
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
				Error(w, error, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
