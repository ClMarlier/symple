package symple

import (
	"fmt"
	"net/http"
	"runtime"
)

// WithRecoverer handle gracefully any panic that could occur, if used after
// WithStructLogger the error will be logged
func (rs *routerState) WithRecoverer(writeError bool) routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, recoverer(writeError))
		return nil
	}
}

func recoverer(writeError bool) func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if recErr := recover(); recErr != nil {
					if !writeError {
						err = ErrInternalServer
						return
					} else {
						err = fmt.Errorf("%+v", recErr)
						stack := make([]uintptr, 20)
						n := runtime.Callers(1, stack)
						stack = stack[:n]
						frames := runtime.CallersFrames(stack)
						frames.Next()

						for frame, more := frames.Next(); more; frame, more = frames.Next() {
							err = fmt.Errorf("%s\n%v", frame.Function, err)
						}
						err = fmt.Errorf("%w\n%w", ErrInternalServer, err)
					}
				}
			}()
			return next(w, r)
		}
	}
}
