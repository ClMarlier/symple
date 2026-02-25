package symple

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIdHeader = "X-Request-Id"

func (rs *routerState) WithRequestId() routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, requestIdMiddleware)
		return nil
	}
}

func requestIdMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		requestId, err := uuid.NewV6()
		if err != nil {
			return err
		}

		w.Header().Set(RequestIdHeader, requestId.String())
		r = r.WithContext(context.WithValue(r.Context(), "request_id", requestId.String()))
		return next(w, r)
	}
}
