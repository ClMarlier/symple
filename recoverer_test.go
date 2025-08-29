package symple

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRecoverer(t *testing.T) {
	testTable := []struct {
		name           string
		writeError     bool
		expectedResult bool
	}{
		{
			name:           "writing error stacktrace",
			writeError:     true,
			expectedResult: false,
		},
		{
			name:           "not writing error stacktrace",
			writeError:     false,
			expectedResult: true,
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))

			rs := NewRouter(ErrFuncDefault)
			mux, err := rs.Router(
				rs.WithRecoverer(val.writeError),
				rs.WithRoute("POST /test", func(w http.ResponseWriter, r *http.Request) error {
					panic("triggered error")
				}),
			)
			if err != nil {
				t.Fatal(err.Error())
			}

			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err.Error())
			}

			if string(body[:len(ErrInternalServer.Error())]) != ErrInternalServer.Error() {
				t.Fatalf("wrong response body: %s", string(body))
			}
			if val.writeError && len(string(body)) == len(ErrInternalServer.Error())+1 {
				t.Fatalf("missing stacktrace in the body: %s", string(body))
			}
		})
	}
}
