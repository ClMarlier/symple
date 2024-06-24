package symple

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithContentType(t *testing.T) {
	testTable := []struct {
		name            string
		contentType     string
		withContentType routerOption
		expectedValue   int
		error           string
	}{
		{
			name:            "matching content type without charset",
			contentType:     "application/json",
			withContentType: WithContentType(WithApplicationJSON),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "matching content with charset",
			contentType:     "application/json; charset=UTF-8",
			withContentType: WithContentType(WithApplicationJSON),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "matching content with multiple content type accepted",
			contentType:     "application/json; charset=UTF-8",
			withContentType: WithContentType(WithApplicationXML, WithApplicationJSON),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "wrong content type",
			contentType:     "xxx",
			withContentType: WithContentType(WithApplicationJSON),
			expectedValue:   http.StatusUnsupportedMediaType,
			error:           "",
		},
		{
			name:            "matching content application/xml",
			contentType:     "application/xml",
			withContentType: WithContentType(WithApplicationXML),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "matching content application/x-www-form-urlencoded",
			contentType:     "application/x-www-form-urlencoded",
			withContentType: WithContentType(WithFormEncoded),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "matching content multipart/form-data",
			contentType:     "multipart/form-data",
			withContentType: WithContentType(WithFormData),
			expectedValue:   http.StatusOK,
			error:           "",
		},
		{
			name:            "double content application/json error",
			contentType:     "application/json",
			withContentType: WithContentType(WithApplicationJSON, WithApplicationJSON),
			expectedValue:   http.StatusOK,
			error:           "duplicate Content-Type",
		},

		{
			name:            "double content application/xml error",
			contentType:     "application/xml",
			withContentType: WithContentType(WithApplicationXML, WithApplicationXML),
			expectedValue:   http.StatusOK,
			error:           "duplicate Content-Type",
		},
		{
			name:            "double content application/x-www-form-urlencoded error",
			contentType:     "application/x-www-form-urlencoded",
			withContentType: WithContentType(WithFormEncoded, WithFormEncoded),
			expectedValue:   http.StatusOK,
			error:           "duplicate Content-Type",
		},
		{
			name:            "double content multipart/form-data error",
			contentType:     "multipart/form-data",
			withContentType: WithContentType(WithFormData, WithFormData),
			expectedValue:   http.StatusOK,
			error:           "duplicate Content-Type",
		},
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			mux, err := Router(
				val.withContentType,
				WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
			)
			if err != nil {
				if err.Error() == val.error {
					return
				}
				t.Fatalf(err.Error())
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))
			req.Header.Set("Content-Type", val.contentType)
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()

			if res.StatusCode != val.expectedValue {
				t.Fatalf("Invalid status values: %d, expected %d", res.StatusCode, val.expectedValue)
			}
		})
	}

}
