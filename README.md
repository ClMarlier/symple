# Symple v0.01
## Motivation
TODO

## Install

`go get -u github.com/ClMarlier/symple`

## Features


## Usage
```go
package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/ClMarlier/symple"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	mux, err := symple.Router(
		symple.WithOption(true),
		symple.WithStructLogger(
			slog.NewTextHandler(
				os.Stdout,
				nil,
			),
		),
		symple.WithRecoverer(true),
		symple.WithRoute("GET /hello", hello),
		symple.WithRouter(
			symple.WithPrefix("/error"),
			symple.WithRoute("GET /simple", simpleError),
			symple.WithRoute("GET /panic/{n}", panicError),
		),
		symple.WithRouter(
			symple.WithPrefix("/protected"),
			symple.WithAuthJWT(
				symple.WithSecret("1234"),
				symple.WithSigningMethod(jwt.SigningMethodHS256),
			),
			symple.WithRoute("GET /hello", protected_hello),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	symple.Startup()
	log.Fatal(http.ListenAndServe(":666", mux))
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func protected_hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Token"))
}

func simpleError(w http.ResponseWriter, _ *http.Request) {
	symple.ErrorResponse(w, fmt.Errorf("Big error of the doom"), http.StatusUnprocessableEntity)
}

// to get a panic simply call with n=0
func panicError(w http.ResponseWriter, r *http.Request) {
	pathNumber := r.PathValue("n")
	n, err := strconv.ParseInt(pathNumber, 10, 32)
	if err != nil {
		symple.ErrorResponse(w, err, http.StatusUnprocessableEntity)
	}
	res := 666 / n
	w.Write([]byte(fmt.Sprintf("666/%d = %d", n, res)))
}
```
