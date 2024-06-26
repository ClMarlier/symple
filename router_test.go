package symple

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	testTable := []struct {
		name        string
		path        string
		requestPath string
		prefix      string
		shouldError bool
	}{
		{
			name:        "empty prefix with method",
			path:        "GET /hello",
			requestPath: "/hello",
			prefix:      "",
			shouldError: false,
		},
		{
			name:        "empty prefix without method",
			path:        "/hello",
			requestPath: "/hello",
			prefix:      "",
			shouldError: false,
		},
		{
			name:        "with prefix with method",
			path:        "GET /hello",
			requestPath: "/prefix/hello",
			prefix:      "/prefix",
			shouldError: false,
		},
		{
			name:        "with prefix without method",
			path:        "/hello",
			requestPath: "/prefix/hello",
			prefix:      "/prefix",
			shouldError: false,
		},
		{
			name:        "wrong prefix error",
			path:        "/hello",
			requestPath: "/prefix/hello",
			prefix:      "/broken prefix",
			shouldError: true,
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", val.requestPath, bytes.NewReader([]byte("body")))
			mux, err := Router(
				WithRouter(
					WithPrefix(val.prefix),
					WithRoute(val.path, func(w http.ResponseWriter, r *http.Request) {}),
				),
			)

			if err != nil {
				if val.shouldError {
					return
				}
				t.Fatalf("unextected initialization error: %s", err.Error())
			}
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("error with the route initialization, returning %d", res.StatusCode)
			}
		})
	}
}

func TestWithOption(t *testing.T) {
	mux1, err := Router(
		WithOption(true),
		WithRoute("GET /test", func(w http.ResponseWriter, r *http.Request) {}),
		WithRoute("PATCH /test", func(w http.ResponseWriter, r *http.Request) {}),
	)
	if err != nil {
		t.Fatalf("error building the mux %s", err.Error())
	}

	mux2, err := Router(
		WithOption(true),
		WithRoute("/test", func(w http.ResponseWriter, r *http.Request) {}),
	)
	if err != nil {
		t.Fatalf("error building the mux %s", err.Error())
	}

	testTable := []struct {
		name   string
		mux    *http.ServeMux
		expect string
	}{
		{
			name:   "option handler with methods",
			mux:    mux1,
			expect: "GET, PATCH",
		},
		{
			name:   "option handler without method",
			mux:    mux2,
			expect: "GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH",
		},
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("OPTIONS", "/test", bytes.NewReader([]byte("body")))
			val.mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			accept := res.Header.Get("Accept")
			if accept != val.expect {
				t.Fatalf("Expected %s found %s", val.expect, accept)
			}
		})
	}
}
