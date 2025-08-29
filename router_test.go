package symple

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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
			rs := NewRouter(ErrFuncDefault)
			mux, err := rs.Router(
				rs.WithRouter(
					rs.WithPrefix(val.prefix),
					rs.WithRoute(val.path, func(w http.ResponseWriter, r *http.Request) error { return nil }),
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

func TestWithOptions(t *testing.T) {
	rs := NewRouter(ErrFuncDefault)
	mux1, err := rs.Router(
		rs.WithOptions(true),
		rs.WithRoute("GET /test", func(w http.ResponseWriter, r *http.Request) error { return nil }),
		rs.WithRoute("PATCH /test", func(w http.ResponseWriter, r *http.Request) error { return nil }),
	)
	if err != nil {
		t.Fatalf("error building the mux %s", err.Error())
	}

	mux2, err := rs.Router(
		rs.WithOptions(true),
		rs.WithRoute("/test", func(w http.ResponseWriter, r *http.Request) error { return nil }),
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

func TestWithOptionsPatternError(t *testing.T) {
	rs := NewRouter(ErrFuncDefault)
	_, err := rs.Router(
		rs.WithOptions(true),
		rs.WithRoute("GET POST /test", func(w http.ResponseWriter, r *http.Request) error { return nil }),
	)
	if err == nil {
		t.Fatal("should return an error")

	}
	if !strings.Contains(err.Error(), "malformated handler pattern") {
		t.Fatalf("should return a malformated error patern, found: %s", err.Error())
	}
}

func TestWithSitemap(t *testing.T) {
	host := "http://localhost:7331"
	rs := NewRouter(ErrFuncDefault)
	rs.SetHost(host)

	mux, err := rs.Router(
		rs.WithSitemap(true),
		rs.WithRoute("GET /test-1", func(w http.ResponseWriter, r *http.Request) error { return nil }),
		rs.WithRoute("GET /test-2", func(w http.ResponseWriter, r *http.Request) error { return nil }),
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/sitemap.xml", nil)
	mux.ServeHTTP(recorder, req)

	res := recorder.Result()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !(strings.Contains(string(body), host+"/test-1") && strings.Contains(string(body), host+"/test-2")) {
		t.Fatal("sitemap does'nt seems to contain the proper informations")
	}
}

func BenchmarkRouter(b *testing.B) {
	rs := NewRouter(ErrFuncDefault)
	mux, err := rs.Router(
		rs.WithRouter(
			rs.WithRecoverer(true),
			rs.WithRequestId(),
			rs.WithRoute("GET /benchmark", func(w http.ResponseWriter, r *http.Request) error {
				w.Write([]byte("ok"))
				return nil
			}),
		),
	)
	if err != nil {
		b.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/benchmark", nil)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			recorder := httptest.NewRecorder()
			mux.ServeHTTP(recorder, req)
			response := recorder.Result()
			if response.StatusCode != 200 {
				b.Fatal(response.StatusCode)
			}
		}
	})
}
