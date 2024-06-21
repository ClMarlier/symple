package symple

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

type routerBuilder struct {
	prefix          string
	middlewareStack []Middleware
	routeStack      []routeDefinition
	subRouter       []routerBuilder
	option          bool
}

type muxOption func(*routerBuilder) error

type routeDefinition struct {
	pattern         string
	handler         http.HandlerFunc
	middlewareStack []Middleware
}

// Router is the entrypoint to build the http.ServeMux. Be careful Router is
// intended to be at the root, if you want nested router use WithRouter instead
//
// Available muxOption:
//
//   - WithOption()
//   - WithPrefix()
//   - WithRoute()
//   - WithRouter()
//
// Available middleware:
//
//   - WithAuthJWT()
//   - WithStructLogger()
//   - WithMiddleware()
//   - WithRecoverer
func Router(opts ...muxOption) (*http.ServeMux, error) {
	router, err := initRouter(opts...)
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	for _, route := range router.routeStack {
		mux.HandleFunc(route.pattern, chainMiddleware(route.handler, route.middlewareStack...))
	}
	beforeStart()
	return mux, nil
}

func initRouter(opts ...muxOption) (routerBuilder, error) {
	rb := routerBuilder{
		prefix:          "",
		middlewareStack: []Middleware{},
		routeStack:      []routeDefinition{},
		subRouter:       []routerBuilder{},
		option:          false,
	}

	// Load options
	for _, opt := range opts {
		if err := opt(&rb); err != nil {
			return routerBuilder{}, err
		}
	}

	// add routes for HTTP OPTION method
	routeMethods := map[string][]string{}
	for _, route := range rb.routeStack {
		splitted := strings.Split(route.pattern, " ")
		if len(splitted) == 1 {
			routeMethods[applyPrefix(route.pattern, rb.prefix)] = []string{
				"GET",
				"HEAD",
				"POST",
				"PUT",
				"DELETE",
				"CONNECT",
				"OPTIONS",
				"TRACE",
				"PATCH"}
			continue
		}

		if len(splitted) == 2 {
			method := splitted[0]
			path := applyPrefix(splitted[1], rb.prefix)
			if _, ok := routeMethods[path]; !ok {
				routeMethods[path] = []string{method}
			} else if !slices.Contains(routeMethods[path], method) {
				routeMethods[path] = append(routeMethods[path], method)
			}
		}
	}
	for path, methods := range routeMethods {
		rb.routeStack = append(rb.routeStack, routeDefinition{
			pattern:         fmt.Sprintf("OPTION %s", path),
			handler:         optionHandler(methods),
			middlewareStack: []Middleware{},
		})
	}

	// add subrouters route to routeStack
	for _, subRouter := range rb.subRouter {
		rb.routeStack = append(rb.routeStack, subRouter.routeStack...)
	}

	// add middlewares
	for i := range rb.routeStack {
		fmt.Println(rb.routeStack[i].pattern)
		rb.routeStack[i].pattern = applyPrefix(rb.routeStack[i].pattern, rb.prefix)
		rb.routeStack[i].middlewareStack = append(rb.middlewareStack, rb.routeStack[i].middlewareStack...)
	}

	return rb, nil
}

// WithRouter adds a subrouter to the current router
func WithRouter(opts ...muxOption) muxOption {
	return func(rb *routerBuilder) error {
		router, err := initRouter(opts...)
		if err != nil {
			return err
		}
		rb.subRouter = append(rb.subRouter, router)

		return nil
	}
}

// WithPrefix set the prefix path for the current router. Keep in mind that
// the current router also inherit from all it's parents prefixes
func WithPrefix(prefix string) muxOption {
	return func(rb *routerBuilder) error {
		rb.prefix = fmt.Sprintf("%s%s", rb.prefix, prefix)
		if prefix == "" {
			return nil
		}
		reg := `^\/[A-Za-z0-9\-._~!$&'()*+,;=:@%]+[A-Za-z0-9\-._~!$&'()*+,;=:@%]$`
		re := regexp.MustCompile(reg)
		success := re.MatchString(prefix)
		if !success {
			return fmt.Errorf("'%s' is not a valid prefix", prefix)
		}
		return nil
	}
}

// WithRoute adds a new route to the current router
func WithRoute(pattern string, handler http.HandlerFunc) muxOption {
	return func(rb *routerBuilder) error {
		rb.routeStack = append(
			rb.routeStack,
			routeDefinition{
				pattern:         pattern,
				handler:         handler,
				middlewareStack: []Middleware{},
			},
		)
		return nil
	}
}

// WithOption is adding if set to true a handler for OPTION method for every child
// route created. You can deactivate this behaviour in child SubRouter by
// setting it to false
func WithOption(active bool) muxOption {
	return func(rb *routerBuilder) error {
		rb.option = active
		return nil
	}
}

func optionHandler(methods []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept", strings.Join(methods, ", "))
	}
}

func applyPrefix(pattern string, prefix string) string {
	if prefix == "" {
		return pattern
	}
	splited := strings.Split(pattern, " ")
	newPath := ""
	if len(splited) == 1 {
		newPath = fmt.Sprintf("%s%s", prefix, splited[0])
	} else {
		newPath = fmt.Sprintf("%s %s%s", splited[0], prefix, splited[1])
	}
	return newPath
}

func beforeStart() {
	fmt.Println("")
	fmt.Println("\033[0;30m\033[102m  ____                            _        \033[0m")
	fmt.Println("\033[0;30m\033[102m / ___|  _   _  _ __ ___   _ __  | |  ___  \033[0m")
	fmt.Println("\033[0;30m\033[102m \\___ \\ | | | || '_ ` _ \\ | '_ \\ | | / _ \\ \033[0m")
	fmt.Println("\033[0;30m\033[102m  ___) || |_| || | | | | || |_) || ||  __/ \033[0m")
	fmt.Println("\033[0;30m\033[102m |____/  \\__, ||_| |_| |_|| .__/ |_| \\___| \033[0m")
	fmt.Println("\033[0;30m\033[102m         |___/            |_|              \033[0m")
	fmt.Println("\033[0;30m\033[102m                                           \033[0m")

}
