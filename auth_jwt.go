package symple

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type tokenSub struct{}

type authJwtConfig struct {
	secret  string
	methods []string
}

type authJwtOption func(*authJwtConfig) error

// WithAuthJWT restricts the access to the current router and all it's
// children subrouters with a valid JWT token
func WithAuthJWT(opts ...authJwtOption) routerOption {
	return func(rb *routerBuilder) error {
		config := &authJwtConfig{
			secret:  "",
			methods: []string{},
		}

		for _, option := range opts {
			if err := option(config); err != nil {
				return err
			}
		}
		if len(config.methods) == 0 {
			return fmt.Errorf("using AuthJWT with no signing method is not allowed")
		}
		if config.secret == "" {
			return fmt.Errorf("using AuthJWT with no secret is not allowed")
		}
		rb.middlewareStack = append(rb.middlewareStack, config.authJWT)
		return nil
	}
}

func WithSecret(secret string) authJwtOption {
	return func(ajc *authJwtConfig) error {
		ajc.secret = secret
		return nil
	}
}

func WithSigningMethod(method jwt.SigningMethod) authJwtOption {
	return func(ajc *authJwtConfig) error {
		if slices.Contains(ajc.methods, method.Alg()) {
			return fmt.Errorf("duplicate signing method")
		}
		ajc.methods = append(ajc.methods, method.Alg())
		return nil
	}
}

func subFromToken(tokenString string, secret string, methods []string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	},
		jwt.WithExpirationRequired(),
		jwt.WithValidMethods(methods),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		return "", err
	}

	sub, err := token.Claims.GetSubject()
	if err != nil || sub == "" {
		return "", fmt.Errorf("'sub' claim is invalid")
	}
	return sub, nil
}

func (ac *authJwtConfig) authJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			ErrorResponse(
				w,
				fmt.Errorf("missing authorization header"),
				http.StatusUnauthorized)
			return
		}
		splitedTokenString := strings.Split(tokenString, " ")
		if len(splitedTokenString) != 2 {
			ErrorResponse(
				w,
				fmt.Errorf("invalid format for authorization header "),
				http.StatusUnauthorized)
			return
		}
		if splitedTokenString[0] != "Bearer" {
			ErrorResponse(
				w,
				fmt.Errorf("the authorization header should be a Bearer"),
				http.StatusUnauthorized)
			return
		}
		sub, err := subFromToken(splitedTokenString[1], ac.secret, ac.methods)
		if err != nil {
			ErrorResponse(
				w,
				err,
				http.StatusUnauthorized)
			return
		} else if sub != "" {
			ctx := context.WithValue(r.Context(), tokenSub{}, sub)
			r = r.WithContext(ctx)
		}
		next(w, r)
	}
}
