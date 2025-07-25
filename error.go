package symple

import (
	"errors"
	"net/http"
)

type HttpError interface {
	Error() string
	StatusCode() int
}

type Error struct {
	message    string
	statusCode int
}

func (e Error) Error() string {
	return e.message
}

func (e Error) StatusCode() int {
	return e.statusCode
}

func GetSympleError(err error) (error, int) {
	for {
		sympleError, ok := err.(HttpError)
		if ok {
			return sympleError, sympleError.StatusCode()
		}
		err = errors.Unwrap(err)
		if err == nil {
			return nil, 0
		}
	}
}

var (
	ErrBadRequest                    = Error{"bad request", http.StatusBadRequest}
	ErrUnauthorized                  = Error{"unauthorized", http.StatusUnauthorized}
	ErrPaymentRequired               = Error{"payment required", http.StatusPaymentRequired}
	ErrForbidden                     = Error{"forbidden", http.StatusForbidden}
	ErrNotFound                      = Error{"ressource not found", http.StatusNotFound}
	ErrMethodNotAllowed              = Error{"method not allowed", http.StatusMethodNotAllowed}
	ErrNotAcceptable                 = Error{"not acceptable", http.StatusNotAcceptable}
	ErrProxyAuthRequired             = Error{"proxy authentication required", http.StatusProxyAuthRequired}
	ErrRequestTimeout                = Error{"request time-out", http.StatusRequestTimeout}
	ErrConflict                      = Error{"conflict", http.StatusConflict}
	ErrGone                          = Error{"gone", http.StatusGone}
	ErrLengthRequired                = Error{"length required", http.StatusLengthRequired}
	ErrPreconditionFailed            = Error{"precondition failed", http.StatusPreconditionFailed}
	ErrRequestEntityTooLarge         = Error{"request entity too large", http.StatusRequestEntityTooLarge}
	ErrRequestURITooLong             = Error{"request URI too long", http.StatusRequestURITooLong}
	ErrUnsuportedMediaType           = Error{"unsupported media type", http.StatusUnsupportedMediaType}
	ErrRequestedRangeNotSatisfiable  = Error{"requested range unsatisfiable", http.StatusRequestedRangeNotSatisfiable}
	ErrExpectationFailed             = Error{"expectation failed", http.StatusExpectationFailed}
	ErrTeapot                        = Error{"i'm a teapot", http.StatusTeapot}
	ErrMisdirectedRequest            = Error{"misdirected request", http.StatusMisdirectedRequest}
	ErrUnprocessableEntity           = Error{"unprocessable entiy", http.StatusUnprocessableEntity}
	ErrLocked                        = Error{"locked", http.StatusLocked}
	ErrFailedDependency              = Error{"failed dependency", http.StatusFailedDependency}
	ErrTooEarly                      = Error{"too early", http.StatusTooEarly}
	ErrUpgradeRequired               = Error{"upgrade required", http.StatusUpgradeRequired}
	ErrPreconditionRequired          = Error{"precondition required", http.StatusPreconditionRequired}
	ErrTooManyRequests               = Error{"too many requests", http.StatusTooManyRequests}
	ErrRequestHeaderFieldsTooLarge   = Error{"request header fields too large", http.StatusRequestHeaderFieldsTooLarge}
	ErrUnavailableForLegalReasons    = Error{"unavailable for legal reasons", http.StatusUnavailableForLegalReasons}
	ErrInternalServer                = Error{"internal server error", http.StatusInternalServerError}
	ErrNotImplemented                = Error{"not implemented", http.StatusNotImplemented}
	ErrBadGateway                    = Error{"bad gateway", http.StatusBadGateway}
	ErrServiceUnavailable            = Error{"service unavailable", http.StatusServiceUnavailable}
	ErrGatewayTimeout                = Error{"gateway time-out", http.StatusGatewayTimeout}
	ErrHTTPVersionNotSupported       = Error{"http version not supported", http.StatusHTTPVersionNotSupported}
	ErrVariantAlsoNegotiates         = Error{"variant also negotiates", http.StatusVariantAlsoNegotiates}
	ErrInsufficientStorage           = Error{"insufficient storage", http.StatusInsufficientStorage}
	ErrLoopDetected                  = Error{"loop detected", http.StatusLoopDetected}
	ErrNotExtended                   = Error{"not extended", http.StatusNotExtended}
	ErrNetworkAuthenticationRequired = Error{"network authentication required", http.StatusNetworkAuthenticationRequired}
)

func ErrFuncDefault(w http.ResponseWriter, r *http.Request, err error) {
	if err == nil {
		return
	}

	sympleErr, statusCode := GetSympleError(err)
	if sympleErr != nil {
		http.Error(w, sympleErr.Error(), statusCode)
		return
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
