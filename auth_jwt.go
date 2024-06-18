package symple

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type authJwtConfig struct {
	secret  string
	methods []string
}

type authJwtOption func(*authJwtConfig) error

// WithAuthJWT add a middleware that restrict the access to the route define in
// the current router and all the child subrouters.
func WithAuthJWT(opts ...authJwtOption) muxOption {
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
			return fmt.Errorf("Using AuthJWT with no signing method is not allowed")
		}
		if config.secret == "" {
			return fmt.Errorf("Using AuthJWT with no secret is not allowed")
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
	ajc.methods = append(ajc.methods, "HS256")
	return nil
}

func subFromToken(tokenString string, secret string, methods []string) (int, error) {
	if tokenString == "" {
		return 0, nil
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return 0, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods(methods),
	)

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		sub, err := claims.GetSubject()
		if err != nil {
			return 0, err
		}
		user, err := strconv.Atoi(sub)
		if err != nil {
			return 0, err
		}
		return user, nil
	} else {
		return 0, err
	}
}

func (ac *authJwtConfig) authJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}
		splitedTokenString := strings.Split(tokenString, " ")
		if len(splitedTokenString) != 2 {
			http.Error(w, "Invalid format for Authorization Header", http.StatusUnauthorized)
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
		} else if sub != 0 {
			ctx := context.WithValue(r.Context(), "user", sub)
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}
