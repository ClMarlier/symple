package symple

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAuthJWT(t *testing.T) {
	Startup()
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "symple",
		"sub": "clement",
		"exp": time.Now().UTC().Add(5 * time.Second).Unix(),
		"iat": time.Now().UTC().Unix(),
	})
	validTokenString, err := validToken.SignedString([]byte("1234"))
	if err != nil {
		t.Fatal(err.Error())
	}

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "symple",
		"sub": "clement",
		"exp": time.Now().UTC().Add(-5 * time.Second).Unix(),
		"iat": time.Now().UTC().Unix(),
	})
	expiredTokenString, err := expiredToken.SignedString([]byte("1234"))
	if err != nil {
		t.Fatal(err.Error())
	}

	wrongMethodToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss": "symple",
		"sub": "clement",
		"exp": time.Now().UTC().Add(5 * time.Second).Unix(),
		"iat": time.Now().UTC().Unix(),
	})
	wrongMethodTokenString, err := wrongMethodToken.SignedString([]byte("1234"))
	if err != nil {
		t.Fatal(err.Error())
	}

	noSubToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "symple",
		"exp": time.Now().UTC().Add(5 * time.Second).Unix(),
		"iat": time.Now().UTC().Unix(),
	})
	noSubTokenString, err := noSubToken.SignedString([]byte("1234"))
	if err != nil {
		t.Fatal(err.Error())
	}

	testTable := []struct {
		name               string
		authorization      string
		withAuthJwt        routerOption
		routerError        string
		expectedStatusCode int
		expectedResponse   string
	}{
		{
			name:               "missing secret",
			authorization:      "",
			withAuthJwt:        WithAuthJWT(WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "using AuthJWT with no secret is not allowed",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "",
		},
		{
			name:               "without signing method",
			authorization:      "",
			withAuthJwt:        WithAuthJWT(WithSecret("1234")),
			routerError:        "using AuthJWT with no signing method is not allowed",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "",
		},
		{
			name:          "with duplicate signign method",
			authorization: "wrong",
			withAuthJwt: WithAuthJWT(
				WithSecret("1234"),
				WithSigningMethod(jwt.SigningMethodHS256),
				WithSigningMethod(jwt.SigningMethodHS256),
			),
			routerError:        "duplicate signing method",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "",
		},
		{
			name:               "unauthorized",
			authorization:      "wrong",
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "",
		},
		{
			name:               "authorized",
			authorization:      fmt.Sprintf("Bearer %s", validTokenString),
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   "clement",
		},
		{
			name:               "missing auth header",
			authorization:      "",
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "missing authorization header",
		},
		{
			name:               "not a Bearer token",
			authorization:      fmt.Sprintf("NotBearer %s", expiredTokenString),
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "the authorization header should be a Bearer",
		},
		{
			name:               "expired",
			authorization:      fmt.Sprintf("Bearer %s", expiredTokenString),
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "token has invalid claims: token is expired",
		},
		{
			name:               "wrong signing method",
			authorization:      fmt.Sprintf("Bearer %s", wrongMethodTokenString),
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "token signature is invalid: signing method",
		},
		{
			name:               "no sub token",
			authorization:      fmt.Sprintf("Bearer %s", noSubTokenString),
			withAuthJwt:        WithAuthJWT(WithSecret("1234"), WithSigningMethod(jwt.SigningMethodHS256)),
			routerError:        "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   "'sub' claim is invalid",
		},
	}

	for _, val := range testTable {
		t.Run(val.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("body")))
			if val.authorization != "" {
				req.Header.Set("Authorization", val.authorization)
			}
			mux, err := Router(
				val.withAuthJwt,
				WithRoute(
					"POST /test",
					func(w http.ResponseWriter, r *http.Request) {
						sub, ok := r.Context().Value(tokenSub{}).(string)
						if !ok {
							http.Error(w, "could'nt get tokenSub from context", http.StatusInternalServerError)
						}
						w.Write([]byte(sub))
					}),
			)
			if err != nil {
				if err.Error() == val.routerError {
					return
				}
				t.Fatal(err.Error())
			}

			mux.ServeHTTP(recorder, req)

			res := recorder.Result()
			if res.StatusCode != val.expectedStatusCode {
				t.Fatalf("wrong response status code %d expected %d found", res.StatusCode, val.expectedStatusCode)
			}
			if val.expectedResponse != "" {
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err.Error())
				}

				if !strings.HasPrefix(string(body), val.expectedResponse) {
					t.Fatalf("invalid response: expected %s found '%s'", val.expectedResponse, string(body))
				}
			}
		})
	}
}
