package symple

import (
	"net/http"

	"github.com/google/uuid"
)

func (rs *routerState) WithRequestId() routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, requestIdMiddleware)
		return nil
	}
}

func requestIdMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		rid, err := uuid.NewV6()
		if err != nil {
			return err
		}
		w.Header().Set("X-Request-Id", rid.String())
		return next(w, r)
	}
}
