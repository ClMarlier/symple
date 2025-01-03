package symple

import (
	"errors"
	"net/http"
)

//TODO wrap error with a custom struct to add Status() func and avoid the switch statement

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUnsuportedMedia = errors.New("unsupported media type")
	ErrInternalServer  = errors.New("internal server error")
	ErrNotFound        = errors.New("ressource not found")
)

func ErrFuncDefault(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}
	http.Error(w, err.Error(), errorStatusCode(err))
}

func errorStatusCode(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrUnsuportedMedia):
		return http.StatusUnsupportedMediaType
	case errors.Is(err, ErrInternalServer):
		return http.StatusInternalServerError
	default:
		return http.StatusTeapot
	}
}
