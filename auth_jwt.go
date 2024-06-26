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

func WithHS256(ajc *authJwtConfig) error {
	if slices.Contains(ajc.methods, "HS256") {
		return fmt.Errorf("duplicate signing method")
	}
	ajc.methods = append(ajc.methods, "HS256")
	return nil
}

func subFromToken(tokenString string, secret string, methods []string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	},
		jwt.WithValidMethods(methods),
		jwt.WithIssuedAt(),
	)

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		sub, err := claims.GetSubject()
		if err != nil {
			return "", err
		}
		return sub, nil
	} else {
		return "", err
	}
}

func (ac *authJwtConfig) authJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}
		splitedTokenString := strings.Split(tokenString, " ")
		if len(splitedTokenString) != 2 {
			http.Error(w, "Invalid format for Authorization Header ", http.StatusUnauthorized)
			return
		}
		if splitedTokenString[0] != "Bearer" {
			http.Error(w, "The Authorization Header should be a Bearer", http.StatusUnauthorized)
			return
		}
		sub, err := subFromToken(splitedTokenString[1], ac.secret, ac.methods)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		} else if sub != "" {
			ctx := context.WithValue(r.Context(), tokenSub{}, sub)
			r = r.WithContext(ctx)
		}
		next(w, r)
	}
}
