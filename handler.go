package symple

import (
	"net/http"
)

type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

func (rs *routerState) MakeHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			rs.errorHandlerFunc(w, r, err)
		}
	}
}
