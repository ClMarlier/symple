# Symple v0.1.0
## Motivation
TODO

## Install

`go get -u github.com/ClMarlier/symple`

## Features


## Usage
```go
import (
        "fmt"
        "log"
        "net/http"
        "os"
        "strconv"

        "github.com/ClMarlier/symple"
)

func main() {
        // ErrFuncDefault is used to handle client response in case of an error
        // occuring anywhere in the chain of middleware or the handler itself.
        // You can provide your own implementation
        rs := symple.NewRouter(symple.ErrFuncDefault)

        mux, err := rs.Router(
                rs.WithZeroLog(os.Stdout),
                rs.WithRecoverer(true),
                rs.WithRoute("GET /hello", hello),
                rs.WithRouter(
                        rs.WithPrefix("/error"),
                        rs.WithRoute("GET /simple", simpleError),
                        rs.WithRoute("GET /panic/{n}", panicError),
                ),
        )
        if err != nil {
                log.Fatal(err)
        }
        log.Fatal(http.ListenAndServe(":8000", mux))
}

func hello(w http.ResponseWriter, r *http.Request) error {
        _, err := w.Write([]byte("Hello World"))
        return err
}

func simpleError(w http.ResponseWriter, _ *http.Request) error {
        return fmt.Errorf("%w custom description of the error", symple.ErrUnauthorized)
}

// to get a panic simply call with n=0
func panicError(w http.ResponseWriter, r *http.Request) error {
        pathNumber := r.PathValue("n")
        n, err := strconv.ParseInt(pathNumber, 10, 32)
        if err != nil {
                symple.ErrorResponse(w, err, http.StatusUnprocessableEntity)
        }
        res := 666 / n
        _, err = w.Write([]byte(fmt.Sprintf("666/%d = %d", n, res)))
        return err
}
```
