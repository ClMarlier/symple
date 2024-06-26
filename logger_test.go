package symple

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testBody struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type logEntry struct {
	Time   string   `json:"time"`
	Level  string   `json:"level"`
	Msg    string   `json:"msg"`
	Status int      `json:"status"`
	Method string   `json:"method"`
	Path   string   `json:"path"`
	Body   testBody `json:"body"`
	User   *string  `json:"user"`
	Error  string   `json:"error"`
}

func TestWithStructLogger(t *testing.T) {
	testTable := []struct {
		name string
		body testBody
	}{
		{
			name: "succes json post",
			body: testBody{Name: "clement", Age: 99},
		},
	}

	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	mux, err := Router(
		WithStructLogger(handler),
		WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) {}),
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			byteBody, err := json.Marshal(val.body)
			if err != nil {
				t.Fatalf(err.Error())
			}
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader(byteBody))
			req.Header.Add("Content-Type", "application/json")
			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			if res.StatusCode != http.StatusOK {
				t.Fatalf("Wrong status code 200 expected, %d found", res.StatusCode)
			}
			var decodedLog logEntry
			json.Unmarshal(buf.Bytes(), &decodedLog)
			if err != nil {
				t.Fatal(err.Error())
			}
			if val.body.Name != decodedLog.Body.Name {
				t.Fatalf("wrong body logged, expecting %#v found %#v", val.body.Name, decodedLog.Body.Name)
			}
		})
	}

}
