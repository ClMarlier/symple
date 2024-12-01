package symple

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type Body struct {
	Name []string `json:"name"`
	Age  []string `json:"age"`
}

type logEntry struct {
	Time   string  `json:"time"`
	Level  string  `json:"level"`
	Msg    string  `json:"msg"`
	Status int     `json:"status"`
	Method string  `json:"method"`
	Path   string  `json:"path"`
	Body   Body    `json:"body"`
	User   *string `json:"user"`
	Error  string  `json:"error"`
}

func TestWithStructLogger(t *testing.T) {
	testTable := []struct {
		name           string
		cType          string
		body           Body
		responseStatus int
		writeBody      bool
		error          string
		panic          string
	}{
		{
			name:           "succes json",
			cType:          "json",
			body:           Body{Name: []string{"clement"}, Age: []string{"99"}},
			responseStatus: http.StatusOK,
			writeBody:      true,
			error:          "",
			panic:          "",
		},
		{
			name:           "error form-url-encoded",
			cType:          "form-encoded",
			body:           Body{Name: []string{"clement"}, Age: []string{"99"}},
			responseStatus: http.StatusUnprocessableEntity,
			writeBody:      true,
			error:          "error from http handler",
			panic:          "",
		},
		{
			name:           "panic multipart/form-data",
			cType:          "multipart",
			body:           Body{Name: []string{"clement"}, Age: []string{"99"}},
			responseStatus: http.StatusInternalServerError,
			writeBody:      true,
			error:          "",
			panic:          "panic from http handler",
		},
		{
			name:           "succes bad content type",
			cType:          "none",
			body:           Body{Name: []string{"clement"}, Age: []string{"99"}},
			responseStatus: http.StatusOK,
			writeBody:      true,
			error:          "",
			panic:          "",
		},
	}

	for _, val := range testTable {
		var buf bytes.Buffer
		handler := slog.NewJSONHandler(&buf, nil)
		mux, err := Router(
			WithStructLogger(handler),
			WithRecoverer(val.writeBody),
			WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {
				if val.error != "" {
					ErrorResponse(w, fmt.Errorf(val.error), http.StatusUnprocessableEntity)
				}
				if val.panic != "" {
					panic(val.panic)
				}
			}),
		)
		if err != nil {
			t.Fatal(err.Error())
		}

		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			switch val.cType {
			case "json":
				byteBody, err := json.Marshal(val.body)
				if err != nil {
					t.Fatal(err.Error())
				}
				req := httptest.NewRequest("POST", "/test", bytes.NewReader(byteBody))
				req.Header.Add("Content-Type", "application/json")
				mux.ServeHTTP(recorder, req)
			case "form-encoded":
				form := url.Values{}
				form.Add("name", val.body.Name[0])
				form.Add("age", val.body.Age[0])
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
				byteBody, err := json.Marshal(val.body)
				if err != nil {
					t.Fatal(err.Error())
				}
				req := httptest.NewRequest("POST", "/test", bytes.NewReader(byteBody))
				req.Header.Add("Content-Type", "text/html")
				mux.ServeHTTP(recorder, req)
			}

			res := recorder.Result()
			if res.StatusCode != val.responseStatus {
				t.Fatalf("Wrong status code 200 expected, %d found", res.StatusCode)
			}
			var decodedLog logEntry
			json.Unmarshal(buf.Bytes(), &decodedLog)
			if err != nil {
				t.Fatal(err.Error())
			}
			if val.cType != "none" {
				if val.body.Name[0] != decodedLog.Body.Name[0] {
					t.Fatalf("wrong body logged, expecting %#v found %#v", val.body, decodedLog.Body)
				}
				if val.body.Age[0] != decodedLog.Body.Age[0] {
					t.Fatalf("wrong body logged, expecting %#v found %#v", val.body, decodedLog.Body)
				}
			} else {
				if len(decodedLog.Body.Name) != 0 || len(decodedLog.Body.Age) != 0 {
					t.Fatalf("the body should'nt be logged, found %#v", decodedLog.Body)
				}
			}
			if val.error != "" && !strings.HasPrefix(decodedLog.Error, val.error) {
				t.Fatalf("wrong error value expecting %s, found %s", val.error, decodedLog.Error)
			}
			if val.panic != "" && !strings.Contains(decodedLog.Error, val.panic) {
				t.Fatalf("wrong error value expecting %s, found %s", val.panic, decodedLog.Error)
			}
		})
	}
}
