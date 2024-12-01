package symple

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func writeIdMiddleware(id string) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(id))
			next(w, r)
		}
	}
}

func TestChainMiddleware(t *testing.T) {
	testTable := []struct {
		name         string
		path         string
		expectedBody string
	}{
		{
			name:         "same layer middleware",
			path:         "/test",
			expectedBody: "12",
		},
		{
			name:         "nested layer middleware",
			path:         "/nested_test1",
			expectedBody: "1234",
		},
		{
			name:         "nested layer middleware 2",
			path:         "/nested_test2",
			expectedBody: "1256",
		},
	}

	dummyHandler := func(w http.ResponseWriter, r *http.Request) {}
	mux, err := Router(
		WithMiddleware(writeIdMiddleware("1")),
		WithMiddleware(writeIdMiddleware("2")),
		WithRoute("POST /test", dummyHandler),
		WithRouter(
			WithMiddleware(writeIdMiddleware("3")),
			WithMiddleware(writeIdMiddleware("4")),
			WithRoute("POST /nested_test1", dummyHandler),
		),
		WithRouter(
			WithMiddleware(writeIdMiddleware("5")),
			WithMiddleware(writeIdMiddleware("6")),
			WithRoute("POST /nested_test2", dummyHandler),
		),
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", val.path, bytes.NewReader([]byte("body")))
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err.Error())
			}

			if string(body) != val.expectedBody {
				t.Fatalf("Invalid status values: %s, expected %s", string(body), val.expectedBody)
			}
		})
	}
}
