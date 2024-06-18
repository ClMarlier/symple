package symple

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
)

func WithRecoverer(rb *routerBuilder) error {
	rb.middlewareStack = append(rb.middlewareStack, recoverer)
	return nil
}

func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				rw, ok := w.(*trackWriter)
				if ok {
					if rw.statusCode == 200 {
						rw.WriteHeader(http.StatusInternalServerError)
					}
					error := fmt.Errorf("%+v", err)
					stack := make([]uintptr, 20)
					n := runtime.Callers(1, stack)
					stack = stack[:n]
					frames := runtime.CallersFrames(stack)
					frames.Next()

					for frame, more := frames.Next(); more; frame, more = frames.Next() {
						error = fmt.Errorf("%s\n%v", frame.Function, error)
					}

					ctx := context.WithValue(r.Context(), "symple_error", error)
					*r = *(r.WithContext(ctx))
				}
			}
		}()
		next.ServeHTTP(w, r)
	})
}
