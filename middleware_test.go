package symple

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func writeIdMiddleware(id string) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			w.Write([]byte(id))
			return next(w, r)
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

	dummyHandler := func(w http.ResponseWriter, r *http.Request) error { return nil }

	rs := NewRouter(ErrFuncDefault)
	mux, err := rs.Router(
		rs.WithMiddleware(writeIdMiddleware("1")),
		rs.WithMiddleware(writeIdMiddleware("2")),
		rs.WithRoute("POST /test", dummyHandler),
		rs.WithRouter(
			rs.WithMiddleware(writeIdMiddleware("3")),
			rs.WithMiddleware(writeIdMiddleware("4")),
			rs.WithRoute("POST /nested_test1", dummyHandler),
		),
		rs.WithRouter(
			rs.WithMiddleware(writeIdMiddleware("5")),
			rs.WithMiddleware(writeIdMiddleware("6")),
			rs.WithRoute("POST /nested_test2", dummyHandler),
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
