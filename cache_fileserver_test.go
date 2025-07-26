package symple

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestCacheFileserver(t *testing.T) {
	testTable := []struct {
		name string
		path string
	}{
		{
			name: "fetch a file",
			path: "/cache_fileserver_test.go",
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			file, err := os.Open("." + val.path)
			if err != nil {
				t.Fatal(err.Error())
			}

			fileByte, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err.Error())
			}

			h := sha1.New()
			h.Write(fileByte)
			originalSha := base64.URLEncoding.EncodeToString(h.Sum(nil))

			recorder := httptest.NewRecorder()
			cache, err := CacheFileSystem("./")
			if err != nil {
				t.Fatalf("CacheFileSystem initialization error: %s", err)
			}
			rs := NewRouter(ErrFuncDefault)
			req := httptest.NewRequest("GET", val.path, bytes.NewReader([]byte("body")))
			mux, err := rs.Router(
				rs.WithRouter(
					rs.WithRoute("/",
						func(w http.ResponseWriter, r *http.Request) error {
							http.FileServer(cache).ServeHTTP(w, r)
							return nil
						},
					),
				),
			)

			if err != nil {
				t.Fatalf("unextected initialization error: %s", err)
			}

			mux.ServeHTTP(recorder, req)

			res := recorder.Result()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("error with the route initialization, returning %d", res.StatusCode)
			}

			resByte, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("error reading the response body: %s", err)
			}

			h = sha1.New()
			h.Write(resByte)

			resSha := base64.URLEncoding.EncodeToString(h.Sum(nil))
			if originalSha != resSha {
				t.Fatalf("body received do not match with the original file")
			}
		})
	}
}

func TestWithCacheFileServer(t *testing.T) {
	testTable := []struct {
		name           string
		filename       string
		expectedStatus int
	}{
		{
			name:           "existing file",
			filename:       "cache_file_server_test_file.txt",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "file not found",
			filename:       "god.txt",
			expectedStatus: http.StatusNotFound,
		},
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			rs := NewRouter(ErrFuncDefault)

			mux, err := rs.Router(
				rs.WithCacheFileServer("./", "/", "public, max-age=2592000", time.Hour),
			)
			if err != nil {
				t.Fatalf(err.Error())
			}
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("GET", fmt.Sprintf("/%s", val.filename), nil)
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()

			if res.StatusCode != val.expectedStatus {
				t.Fatalf("Invalid status values: %d, expected %d", res.StatusCode, val.expectedStatus)
			}
		})
	}
}
