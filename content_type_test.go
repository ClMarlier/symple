package symple

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithContentType(t *testing.T) {
	testTable := []struct {
		name          string
		contentType   string
		expectedValue int
	}{
		{
			name:          "matching content type no charset",
			contentType:   "application/json",
			expectedValue: http.StatusOK,
		},
		{
			name:          "matching content with charset",
			contentType:   "application/json; charset=UTF-8",
			expectedValue: http.StatusOK,
		},
		{
			name:          "wrong content type no charset",
			contentType:   "application/xml",
			expectedValue: http.StatusUnsupportedMediaType,
		},
		{
			name:          "wrong content with charset",
			contentType:   "application/xml; charset=UTF-8",
			expectedValue: http.StatusUnsupportedMediaType,
		},
	}
	mux1, err := Router(
		WithContentType(WithApplicationJSON),
		WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	mux2, err := Router(
		WithContentType(WithApplicationJSON, WithFormData),
		WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
	)
	if err != nil {
		t.Fatalf(err.Error())
	}

	muxTable := []struct {
		name string
		mux  *http.ServeMux
	}{
		{
			name: "single Content-Type allowed",
			mux:  mux1,
		},
		{
			name: "multiple Content-Type allowed",
			mux:  mux2,
		},
	}

	for _, router := range muxTable {
		for _, val := range testTable {
			t.Run(fmt.Sprintf("%s %s", router.name, val.name), func(t *testing.T) {
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))
				req.Header.Set("Content-Type", val.contentType)
				router.mux.ServeHTTP(recorder, req)

				res := recorder.Result()

				if res.StatusCode != val.expectedValue {
					t.Fatalf("Invalid status values: %d, expected %d", res.StatusCode, val.expectedValue)
				}
			})
		}
	}
}
