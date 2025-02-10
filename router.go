package symple

import (
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

type routerState struct {
	sequence         int
	extra            map[int]routeExtra
	errorHandlerFunc ErrorHandlerFunc
}

func NewRouter(handler ErrorHandlerFunc) *routerState {
	return &routerState{
		sequence:         0,
		extra:            make(map[int]routeExtra),
		errorHandlerFunc: handler,
	}
}

func (rs *routerState) nextSequence() {
	rs.sequence += 1
}

func (rs *routerState) getSequence() int {
	return rs.sequence
}

func (rs *routerState) getExtra(key int) (routeExtra, bool) {
	val, ok := rs.extra[key]
	return val, ok
}

func (rs *routerState) setExtra(key int, value routeExtra) {
	rs.extra[key] = value
}

type setBool struct {
	isSet bool
	value bool
}

type routeExtra struct {
	options setBool
	sitemap setBool
}

type routerBuilder struct {
	prefix          string
	middlewareStack []Middleware
	routeStack      []routeDefinition
	subRouter       []routerBuilder
	options         setBool
	sitemap         setBool
}

type routerOption func(*routerBuilder) error

type routeDefinition struct {
	id              int
	pattern         string
	handler         HandlerFunc
	middlewareStack []Middleware
}

// Router is the entrypoint to build the http.ServeMux. Be careful Router is
// intended to be at the root, if you want nested router use WithRouter instead
//
// Available routerOption:
//
//   - WithOptions()
//   - WithPrefix()
//   - WithRoute()
//   - WithRouter()
//
// Available middleware:
//
//   - WithAuthJWT()
//   - WithStructLogger()
//   - WithMiddleware()
//   - WithRecoverer()
//   - WithRequestContentType()
//   - WithResponseContentType()
func (rs *routerState) Router(opts ...routerOption) (*http.ServeMux, error) {
	router, err := rs.initRouter(opts...)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	options := make(map[string][]string)
	sitemap := make([]string, 0)
	for _, route := range router.routeStack {
		mux.HandleFunc(route.pattern, rs.MakeHandlerFunc(chainMiddleware(route.handler, route.middlewareStack...)))

		if val, ok := rs.getExtra(route.id); ok {
			path, methods, err := parsePattern(route.pattern)
			if err != nil {
				return nil, err
			}

			if val.options.value {
				var nextMethods []string = []string{}
				if val, ok := options[path]; ok {
					nextMethods = val
				}

				for _, method := range methods {
					if !slices.Contains(nextMethods, method) {
						nextMethods = append(nextMethods, method)
					}
				}
				options[path] = nextMethods
			}

			if val.sitemap.value {
				if !slices.Contains(sitemap, path) {
					sitemap = append(sitemap, path)
				}
			}
		}
	}

	// Add handler with options methods to all registered routes
	for path, methods := range options {
		mux.HandleFunc(fmt.Sprintf("OPTIONS %s", path), optionHandler(methods))
	}

	// Add sitemap.xml handler if needed
	if len(sitemap) > 0 {
		fmt.Println(sitemap)
	}
	return mux, nil
}

func (rs *routerState) initRouter(opts ...routerOption) (routerBuilder, error) {
	rb := routerBuilder{
		prefix:          "",
		middlewareStack: []Middleware{},
		routeStack:      []routeDefinition{},
		subRouter:       []routerBuilder{},
		options:         setBool{},
	}

	// Load options
	for _, opt := range opts {
		if err := opt(&rb); err != nil {
			return routerBuilder{}, err
		}
	}

	// add subrouters route to routeStack
	for _, subRouter := range rb.subRouter {
		rb.routeStack = append(rb.routeStack, subRouter.routeStack...)
	}

	// add middlewares
	for i := range rb.routeStack {
		rb.routeStack[i].pattern = applyPrefix(rb.routeStack[i].pattern, rb.prefix)
		rb.routeStack[i].middlewareStack = append(rb.middlewareStack, rb.routeStack[i].middlewareStack...)
	}

	// handle options, sitemap and all route
	for _, route := range rb.routeStack {
		ext, ok := rs.getExtra(route.id)
		if !ok {
			return routerBuilder{}, fmt.Errorf("couldn't load id %d for route %s", route.id, route.pattern)
		}
		if !ext.options.isSet {
			ext.options = rb.options
			rs.setExtra(route.id, ext)
		}
	}

	return rb, nil
}

// WithRouter adds a subrouter to the current router
func (rs *routerState) WithRouter(opts ...routerOption) routerOption {
	return func(rb *routerBuilder) error {
		router, err := rs.initRouter(opts...)
		if err != nil {
			return err
		}
		rb.subRouter = append(rb.subRouter, router)

		return nil
	}
}

// WithPrefix set the prefix path for the current router. Keep in mind that
// the current router also inherit from all it's parents prefixes
func (rs *routerState) WithPrefix(prefix string) routerOption {
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
func (rs *routerState) WithRoute(pattern string, handler HandlerFunc) routerOption {
	return func(rb *routerBuilder) error {
		rb.routeStack = append(
			rb.routeStack,
			routeDefinition{
				id:              rs.getSequence(),
				pattern:         pattern,
				handler:         handler,
				middlewareStack: []Middleware{},
			},
		)
		rs.setExtra(rs.getSequence(), routeExtra{options: setBool{}, sitemap: setBool{}})
		rs.nextSequence()

		return nil
	}
}

// WithOptions is adding if set to true a handler for OPTION method for every child
// route created. You can deactivate this behaviour in child SubRouters by
// setting the active value to false
func (rs *routerState) WithOptions(active bool) routerOption {
	return func(rb *routerBuilder) error {
		rb.options = setBool{isSet: true, value: active}
		return nil
	}
}

// WithSitemap is adding all the child route to the sitemap. You can reverse this
// behaviour in SubRouters by setting the active value to false
func (rs *routerState) WithSitemap(active bool) routerOption {
	return func(rb *routerBuilder) error {
		rb.sitemap = setBool{isSet: true, value: active}
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

func parsePattern(pattern string) (string, []string, error) {
	splitted := strings.Split(pattern, " ")
	if len(splitted) == 1 {
		return pattern,
			[]string{
				"GET",
				"HEAD",
				"POST",
				"PUT",
				"DELETE",
				"CONNECT",
				"OPTIONS",
				"TRACE",
				"PATCH"},
			nil
	}

	if len(splitted) == 2 {
		return splitted[1], []string{splitted[0]}, nil
	}
	return "", []string{}, fmt.Errorf("malformated handler pattern %s", pattern)
}

func (rs *routerState) Startup() {
	fmt.Println("")
	fmt.Println("\033[0;30m\033[102m  ____                            _        \033[0m")
	fmt.Println("\033[0;30m\033[102m / ___|  _   _  _ __ ___   _ __  | |  ___  \033[0m")
	fmt.Println("\033[0;30m\033[102m \\___ \\ | | | || '_ ` _ \\ | '_ \\ | | / _ \\ \033[0m")
	fmt.Println("\033[0;30m\033[102m  ___) || |_| || | | | | || |_) || ||  __/ \033[0m")
	fmt.Println("\033[0;30m\033[102m |____/  \\__, ||_| |_| |_|| .__/ |_| \\___| \033[0m")
	fmt.Println("\033[0;30m\033[102m         |___/            |_|              \033[0m")
	fmt.Println("\033[0;30m\033[102m                                           \033[0m")
}
