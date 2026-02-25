package symple

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestId(t *testing.T) {
	rs := NewRouter(ErrFuncDefault)
	mux, err := rs.Router(
		rs.WithRequestId(),
		rs.WithRoute("GET /test", func(w http.ResponseWriter, r *http.Request) error { return nil }),
	)
	if err != nil {
		t.Fatal(err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	mux.ServeHTTP(recorder, req)

	res := recorder.Result()
	if requestId := res.Header.Get(RequestIdHeader); requestId == "" {
		t.Fatalf("'%s' header not found in response", RequestIdHeader)
	}
}
