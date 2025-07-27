package symple

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithPprof(t *testing.T) {
	testTable := []struct {
		name string
		path string
	}{
		{
			name: "index",
			path: "/debug/pprof",
		},
		{
			name: "cmdline",
			path: "/debug/pprof/cmdline",
		},
		{
			name: "profile",
			path: "/debug/pprof/profile?seconds=1",
		},
		{
			name: "symbol",
			path: "/debug/pprof/symbol",
		},
		{
			name: "trace",
			path: "/debug/pprof/trace",
		},
		{
			name: "heap",
			path: "/debug/pprof/heap",
		},
		{
			name: "goroutine",
			path: "/debug/pprof/goroutine",
		},
		{
			name: "block",
			path: "/debug/pprof/block",
		},
		{
			name: "threadcreate",
			path: "/debug/pprof/threadcreate",
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", val.path, nil)

			rs := NewRouter(ErrFuncDefault)
			mux, err := rs.Router(
				rs.WithPprof(),
				rs.WithRoute("GET /test", func(w http.ResponseWriter, r *http.Request) error {
					w.Write([]byte("dummy"))
					return nil
				}),
			)
			if err != nil {
				t.Fatal(err.Error())
			}

			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("%s not found", val.path)
			}
		})
	}
}
