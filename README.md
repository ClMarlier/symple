### Symple
**work in progress**

```go
package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"symple"
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
		symple.WithRecoverer,
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
				symple.WithHS256,
			),
			symple.WithRoute("GET /hello", protected_hello),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":666", mux))
}

func hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func protected_hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello Token"))
}

func simpleError(w http.ResponseWriter, _ *http.Request) {
	symple.Error(w, fmt.Errorf("Big error of the doom"), http.StatusUnprocessableEntity)
}

// to get a panic simply call with n=0
func panicError(w http.ResponseWriter, r *http.Request) {
	pathNumber := r.PathValue("n")
	n, err := strconv.ParseInt(pathNumber, 10, 32)
	if err != nil {
		symple.Error(w, err, http.StatusUnprocessableEntity)
	}
	res := 666 / n
	w.Write([]byte(fmt.Sprintf("666/%d = %d", n, res)))
}
```

## TODO
- reflechir a éviter que le recoverer ne renvoi le contenu au client (probablement mettre un recovererConfig avec un boolean on/off error details)
- voir pour le cache
