package symple

import (
	"fmt"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func createStack(middlewareStack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewareStack) - 1; i >= 0; i-- {
			fmt.Println(i)
			fn := middlewareStack[i]
			next = fn(next)
		}
		next = baseMiddleware(next)

		return next
	}
}

// WithMiddleware is used to add a custom middleware to the current Router
func WithMiddleware(middleware Middleware) muxOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, middleware)
		return nil
	}
}

func baseMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tw := &trackWriter{w, http.StatusOK, nil}
		next.ServeHTTP(tw, r)
	})
}
