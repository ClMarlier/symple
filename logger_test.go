package symple

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type Body struct {
	Name []string `json:"name"`
	Age  []string `json:"age"`
}

type logEntry struct {
	Time      string `json:"time"`
	Level     string `json:"level"`
	Verb      string `json:"verb"`
	Path      string `json:"path"`
	ErrorDesc string `json:"desc"`
}

func TestWithZeroLog(t *testing.T) {
	body := Body{Name: []string{"clement"}, Age: []string{"99"}}
	testTable := []struct {
		name           string
		cType          string
		responseStatus int
		error          error
	}{
		{
			name:           "log info success",
			cType:          "json",
			responseStatus: http.StatusOK,
			error:          nil,
		},
		{
			name:           "first",
			cType:          "json",
			responseStatus: http.StatusUnauthorized,
			error:          ErrUnauthorized,
		},
	}

	for _, val := range testTable {
		var buf bytes.Buffer
		rs := NewRouter(ErrFuncDefault)
		mux, err := rs.Router(
			rs.WithZeroLog(&buf),
			rs.WithRecoverer(false),
			rs.WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) error {
				return val.error
			}),
		)
		if err != nil {
			t.Fatal(err.Error())
		}

		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			switch val.cType {
			case "json":
				byteBody, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err.Error())
				}
				req := httptest.NewRequest("POST", "/test", bytes.NewReader(byteBody))
				req.Header.Add("Content-Type", "application/json")
				mux.ServeHTTP(recorder, req)
			case "form-encoded":
				form := url.Values{}
				form.Add("name", body.Name[0])
				form.Add("age", body.Age[0])
				req := httptest.NewRequest("POST", "/test", strings.NewReader(form.Encode()))
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				mux.ServeHTTP(recorder, req)
			case "multipart":
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				err := writer.WriteField("name", "clement")
				if err != nil {
					t.Fatalf("Error writing form field: %v", err)
				}
				err = writer.WriteField("age", "99")
				if err != nil {
					t.Fatalf("Error writing form field: %v", err)
				}
				writer.Close()
				req := httptest.NewRequest("POST", "/test", body)
				req.Header.Add("Content-Type", writer.FormDataContentType())
				mux.ServeHTTP(recorder, req)
			default:
				byteBody, err := json.Marshal(body)
				if err != nil {
					t.Fatal(err.Error())
				}
				req := httptest.NewRequest("POST", "/test", bytes.NewReader(byteBody))
				req.Header.Add("Content-Type", "text/html")
				mux.ServeHTTP(recorder, req)
			}

			res := recorder.Result()
			if res.StatusCode != val.responseStatus {
				t.Fatalf("Wrong status code %d expected, %d found", res.StatusCode, val.responseStatus)
			}
			var decodedLog logEntry
			json.Unmarshal(buf.Bytes(), &decodedLog)
			if err != nil {
				t.Fatal(err.Error())
			}

			if val.error == nil {
				if decodedLog.Level != "info" {
					t.Fatalf("level should be info when no error, %s found", decodedLog.Level)
				}
			} else {
				if decodedLog.Level != "error" {
					t.Fatalf("level should be error when there is an error, %s found", decodedLog.Level)
				}
				if decodedLog.ErrorDesc != val.error.Error() {
					t.Fatalf("invalid error type found %s, %s expected", decodedLog.ErrorDesc, val.error.Error())
				}
			}
			if decodedLog.Verb != "POST" {
				t.Fatalf("invalid http verb logger POST expected, %s found", decodedLog.Verb)
			}
			if _, err := time.ParseDuration(decodedLog.Time); err != nil {
				t.Fatalf("invalid time duration %s", decodedLog.Time)
			}
			if decodedLog.Path != "/test" {
				t.Fatalf("invalid path value found %s, /test expected", decodedLog.Path)
			}
		})
	}
}
