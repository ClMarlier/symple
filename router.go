package symple

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"
)

type routerBuilder struct {
	router          *http.ServeMux
	prefix          string
	middlewareStack []Middleware
	routeStack      []routeDefinition
	subRouter       []*http.ServeMux
	routeMap        map[string][]string
	option          bool
}

type muxOption func(*routerBuilder) error

type routeDefinition struct {
	pattern string
	handler http.HandlerFunc
}

// Router is the entrypoint to build the http.ServeMux. Be careful Router is
// intended to be the root router, if you want nested router use WithRouter
//
// Available muxOption:
//
//   - WithOption()
//   - WithPrefix()
//   - WithRoute()
//   - WithRouter()
//   - WithAuthJWT()
//   - WithStructLogger()
//   - WithRecoverer
func Router(opts ...muxOption) *http.ServeMux {
	return routerWithPrefix("", opts...)
}

func routerWithPrefix(prefix string, opts ...muxOption) *http.ServeMux {
	rb := &routerBuilder{
		router:          http.NewServeMux(),
		prefix:          prefix,
		middlewareStack: []Middleware{},
		subRouter:       []*http.ServeMux{},
		routeMap:        map[string][]string{},
		option:          false,
	}

	f, err := os.OpenFile("openapi.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// Load options into routerBuilder struct
	for _, option := range opts {
		if err := option(rb); err != nil {
			log.Fatal(err.Error())
		}
	}

	router := http.NewServeMux()
	stack := createStack(rb.middlewareStack...)

	// Load and append all the subrouters routes to this router
	for _, rh := range rb.subRouter {
		rb.router.Handle("/", rh)
	}

	// Append the routes to this router
	for _, route := range rb.routeStack {
		f.Write([]byte(applyPrefix(route.pattern, rb.prefix) + "\n"))
		rb.router.HandleFunc(applyPrefix(route.pattern, rb.prefix), route.handler)
	}

	// If activated add a handler for each route with method OPTION
	if rb.option == true {
		for path, methods := range rb.routeMap {
			fmt.Println(fmt.Sprintf("%s: %s", path, methods))
			router.HandleFunc(fmt.Sprintf("%s %s", "OPTIONS", path), optionHandler(methods))
		}
	}

	// Wrap this router with the middleware stack
	router.Handle("/", stack(rb.router))

	return router
}

func WithRouter(opts ...muxOption) muxOption {
	return func(rb *routerBuilder) error {
		router := routerWithPrefix(rb.prefix, opts...)
		rb.subRouter = append(rb.subRouter, router)

		return nil
	}
}

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

// WithOption is adding if set to true a handler for OPTION method for every child
// route created. You can deactivate this behaviour in child SubRouter by
// setting it to false
func WithOption(active bool) muxOption {
	return func(rb *routerBuilder) error {
		rb.option = active
		return nil
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

func optionHandler(methods []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Accept", strings.Join(methods, ", "))
	}
}

func WithRoute(pattern string, handler http.HandlerFunc) muxOption {
	return func(rb *routerBuilder) error {
		rb.routeStack = append(
			rb.routeStack,
			routeDefinition{
				pattern: pattern,
				handler: handler},
		)

		if len(strings.Split(pattern, " ")) == 1 {
			rb.routeMap[applyPrefix(pattern, rb.prefix)] = []string{"GET",
				"HEAD",
				"POST",
				"PUT",
				"DELETE",
				"CONNECT",
				"OPTIONS",
				"TRACE",
				"PATCH"}
			return nil
		}
		splittedPath := strings.Split(pattern, " ")
		method := splittedPath[0]
		path := applyPrefix(splittedPath[1], rb.prefix)

		if _, ok := rb.routeMap[path]; !ok {
			rb.routeMap[path] = []string{method}
		} else if !slices.Contains(rb.routeMap[path], method) {
			rb.routeMap[path] = append(rb.routeMap[path], method)
		}
		return nil
	}
}

func BeforeStart() {
	fmt.Println("")
	fmt.Println("\033[0;30m\033[102m  ____                            _        \033[0m")
	fmt.Println("\033[0;30m\033[102m / ___|  _   _  _ __ ___   _ __  | |  ___  \033[0m")
	fmt.Println("\033[0;30m\033[102m \\___ \\ | | | || '_ ` _ \\ | '_ \\ | | / _ \\ \033[0m")
	fmt.Println("\033[0;30m\033[102m  ___) || |_| || | | | | || |_) || ||  __/ \033[0m")
	fmt.Println("\033[0;30m\033[102m |____/  \\__, ||_| |_| |_|| .__/ |_| \\___| \033[0m")
	fmt.Println("\033[0;30m\033[102m         |___/            |_|              \033[0m")
	fmt.Println("\033[0;30m\033[102m                                           \033[0m")

}
