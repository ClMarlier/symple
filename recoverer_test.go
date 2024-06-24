package symple

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoverer(t *testing.T) {
	testTable := []struct {
		name           string
		writeError     bool
		expectedResult bool
	}{
		{
			name:           "recoverer writing error",
			writeError:     true,
			expectedResult: false,
		},
		{
			name:           "recoverer not writing error",
			writeError:     false,
			expectedResult: true,
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))
			mux, err := Router(
				WithRecoverer(val.writeError),
				WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) { panic("error") }),
			)
			if err != nil {
				t.Fatalf(err.Error())
			}

			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf(err.Error())
			}

			if (string(body) == "internal server error\n") != val.expectedResult {
				t.Fatalf("wrong body: %s", string(body))
			}
		})
	}
}
