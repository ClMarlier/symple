package symple

import "net/http"

type trackWriter struct {
	http.ResponseWriter
	statusCode int
	error      error
}

func (tr *trackWriter) WriteHeader(statusCode int) {
	tr.ResponseWriter.WriteHeader(statusCode)
	tr.statusCode = statusCode
}

// Error is used to build the http error response with the provided http status
// code and make theese values available to the middlewares.
func Error(w http.ResponseWriter, error error, statusCode int) {
	http.Error(w, error.Error(), statusCode)
	if writter, ok := w.(*trackWriter); ok {
		writter.error = error
	}
}
