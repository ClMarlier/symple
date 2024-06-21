package symple

import (
	"net/http"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

func chainMiddleware(h http.HandlerFunc, middlewareStack ...Middleware) http.HandlerFunc {
	if len(middlewareStack) == 0 {
		return h
	}
	for i := len(middlewareStack) - 1; i >= 0; i-- {
		h = middlewareStack[i](h)
	}
	h = baseMiddleware(h)
	return h
}

// WithMiddleware is used to add a custom middleware to the current Router
func WithMiddleware(middleware Middleware) muxOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, middleware)
		return nil
	}
}

func baseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tw := &trackWriter{w, http.StatusOK, nil}
		next(tw, r)
	}
}
