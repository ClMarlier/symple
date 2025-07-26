package symple

import (
	"net/http"
)

// HandlerFunc is like http.HandlerFunc but returning an error
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ErrorHandlerFunc is the function being called to handle HandlerFunc errors
type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request, err error)

// MakeHandlerFunc convert a symple.HandlerFunc into a http.HandlerFunc
// any non-nil error will be passed to the symple.ErrorHandlerFunc provided
// during Router initialization
func (rs *routerState) MakeHandlerFunc(h HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			rs.errorHandlerFunc(w, r, err)
		}
	}
}
