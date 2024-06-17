package symple

import (
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func createStack(middlewareStack ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewareStack) - 1; i >= 0; i-- {
			fn := middlewareStack[i]
			next = fn(next)
		}
		return next
	}
}
