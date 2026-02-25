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
	extraInfo        map[int]routeExtra
	errorHandlerFunc ErrorHandlerFunc
}

type handlerFuncWithHost struct {
	host string
	fn   HandlerFunc
}

type preProcessor struct {
	options       map[string][]string
	sitemap       []string
	hostnameRoute map[string][]handlerFuncWithHost
}

func preProcess(rs *routerState, rd []routeDefinition) (preProcessor, error) {
	var pp = preProcessor{
		options:       make(map[string][]string),
		hostnameRoute: make(map[string][]handlerFuncWithHost),
	}
	for _, route := range rd {
		if val, ok := rs.getExtra(route.id); ok {
			path, methods, err := parsePattern(route.pattern)
			if err != nil {
				return pp, err
			}

			if val.options.value {
				var nextMethods []string = []string{}
				if val, ok := pp.options[path]; ok {
					nextMethods = val
				}

				for _, method := range methods {
					if !slices.Contains(nextMethods, method) {
						nextMethods = append(nextMethods, method)
					}
				}
				pp.options[path] = nextMethods
			}

			if val.sitemap.value {
				if !slices.Contains(pp.sitemap, path) {
					pp.sitemap = append(pp.sitemap, path)
				}
			}

			pp.hostnameRoute[route.pattern] = append(
				pp.hostnameRoute[route.pattern],
				handlerFuncWithHost{
					host: val.hostname,
					fn:   chainMiddleware(route.handler, route.middlewareStack...),
				},
			)
		}

	}
	return pp, nil
}

func NewRouter(handler ErrorHandlerFunc) *routerState {
	return &routerState{
		sequence:         0,
		extraInfo:        make(map[int]routeExtra),
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
	val, ok := rs.extraInfo[key]
	return val, ok
}

func (rs *routerState) setExtra(key int, value routeExtra) {
	rs.extraInfo[key] = value
}

type setBool struct {
	isSet bool
	value bool
}

type routeExtra struct {
	options  setBool
	sitemap  setBool
	hostname string
}

type routerBuilder struct {
	prefix          string
	middlewareStack []Middleware
	routeStack      []routeDefinition
	subRouter       []routerBuilder
	options         setBool
	sitemap         setBool
	hostname        string
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
//   - WithCacheFileServer()
//   - WithMiddleware()
//   - WithPprof()
//   - WithRecoverer()
//   - WithRequestContentType()
//   - WithResponseContentType()
//   - WithZeroLog()
func (rs *routerState) Router(opts ...routerOption) (*http.ServeMux, error) {
	router, err := rs.initRouter(opts...)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	preProcessed, err := preProcess(rs, router.routeStack)
	if err != nil {
		return nil, err
	}
	for key, value := range preProcessed.hostnameRoute {
		if len(value) == 1 && value[0].host == "" {
			mux.HandleFunc(key, rs.MakeHandlerFunc(value[0].fn))
		} else {
			mux.HandleFunc(key, rs.MakeHandlerFunc(hostHandlerBuild(value)))
		}
	}

	// Add handler with options methods to all registered routes
	for path, methods := range preProcessed.options {
		mux.HandleFunc(fmt.Sprintf("OPTIONS %s", path), optionHandler(methods))
	}

	// Add sitemap.xml handler to reference all listed routes
	if len(preProcessed.sitemap) > 0 {
		routePerHost := make(map[string][]string)
		for _, url := range preProcessed.sitemap {
			for _, item := range preProcessed.hostnameRoute[fmt.Sprintf("GET %s", url)] {
				routePerHost[item.host] = append(routePerHost[item.host], url)
			}
		}
		// sitemapBytes := generateSitemap(preProcessed.sitemap, rs.host)
		mux.HandleFunc(fmt.Sprintf("GET /sitemap.xml"), func(w http.ResponseWriter, r *http.Request) {
			hostname := getHostname(r)
			var sitemapBytes []byte
			if val, ok := routePerHost[hostname]; ok {
				sitemapBytes = generateSitemap(val, hostname)
			} else if val, ok := routePerHost[""]; ok {
				sitemapBytes = generateSitemap(val, "")
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("Not found"))
				return
			}
			w.Header().Set("Content-Type", fmt.Sprintf("%s; charset=UTF-8", ContentTypeXml))
			w.Write(sitemapBytes)
		})
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
		extraInfo, ok := rs.getExtra(route.id)
		if !ok {
			return routerBuilder{}, fmt.Errorf("couldn't load id %d for route %s", route.id, route.pattern)
		}

		if rb.hostname != "" {
			extraInfo.hostname = rb.hostname
		}
		if !extraInfo.options.isSet {
			extraInfo.options = rb.options
		}
		if !extraInfo.sitemap.isSet {
			extraInfo.sitemap = rb.sitemap
		}

		rs.setExtra(route.id, extraInfo)
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

// WithHostname is restricting children routes to the provided hostname
func (rs *routerState) WithHostname(hostname string) routerOption {
	return func(rb *routerBuilder) error {
		rb.hostname = hostname
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

func generateSitemap(urls []string, host string) []byte {
	sitemap := []byte{}
	sitemap = fmt.Append(sitemap, `<?xml version="1.0" encoding="UTF-8"?><urlset>`)
	for _, url := range urls {
		sitemap = fmt.Appendf(sitemap, "<url><loc>%s%s</loc></url>", host, url)
	}
	sitemap = fmt.Append(sitemap, "</urlset>")

	return sitemap
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

func hostHandlerBuild(items []handlerFuncWithHost) HandlerFunc {
	var noHost HandlerFunc
	for _, item := range items {
		if item.host == "" {
			noHost = item.fn
		}
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		hostname := getHostname(r)
		for _, item := range items {
			if item.host == hostname {
				return item.fn(w, r)
			}
		}
		if noHost != nil {
			return noHost(w, r)
		}
		return ErrNotFound
	}
}

func getHostname(r *http.Request) string {
	var hostname = r.Host
	if forwarded_host := r.Header.Get("X-Forwarded-Host"); forwarded_host != "" {
		hostname = forwarded_host
	}
	return hostname
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
