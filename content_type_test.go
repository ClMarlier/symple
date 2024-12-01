package symple

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestContentType(t *testing.T) {
	testTable := []struct {
		name          string
		contentType   string
		contentTypes  []ContentType
		expectedValue int
		error         string
	}{
		{
			name:          "matching content type without charset",
			contentType:   "application/json",
			contentTypes:  []ContentType{ContentTypeJson},
			expectedValue: http.StatusOK,
			error:         "",
		},
		{
			name:          "matching content with charset",
			contentType:   "application/json; charset=UTF-8",
			contentTypes:  []ContentType{ContentTypeJson},
			expectedValue: http.StatusOK,
			error:         "",
		},
		{
			name:          "matching content with multiple content type accepted",
			contentType:   "application/json; charset=UTF-8",
			contentTypes:  []ContentType{ContentTypeXml, ContentTypeJson},
			expectedValue: http.StatusOK,
			error:         "",
		},
		{
			name:          "wrong content type",
			contentType:   "xxx",
			contentTypes:  []ContentType{ContentTypeJson},
			expectedValue: http.StatusUnsupportedMediaType,
			error:         "",
		},
		{
			name:          "double content application/json error",
			contentType:   "application/json",
			contentTypes:  []ContentType{ContentTypeJson, ContentTypeJson},
			expectedValue: http.StatusOK,
			error:         "found duplicate request Content-Type config application/json",
		},
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			mux, err := Router(
				WithRequestContentType(val.contentTypes),
				WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
			)
			if err != nil {
				if strings.Contains(err.Error(), val.error) {
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

func TestWithResponseContentType(t *testing.T) {
	testTable := []struct {
		name        string
		contentType ContentType
		error       string
	}{
		{
			name:        "good contencontentType",
			contentType: ContentTypeXml,
		},
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			mux, err := Router(
				WithResponseContentType(val.contentType),
				WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
			)
			if err != nil {
				if strings.Contains(err.Error(), val.error) {
					return
				}
				t.Fatalf(err.Error())
			}

			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			ct := res.Header.Get("Content-Type")
			if ct != string(val.contentType) {
				t.Fatalf("Invalid response 'Content-Type': %s, expected %s", string(val.contentType), ct)
			}
		})
	}

}
