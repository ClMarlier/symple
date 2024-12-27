package symple

import (
	"net/http"
	"net/http/pprof"
)

type pprofItems struct {
	pattern string
	handler http.HandlerFunc
}

func WithPprof() routerOption {
	return func(rb *routerBuilder) error {
		pItems := []pprofItems{
			{pattern: "GET /debug/pprof", handler: pprof.Index},
			{pattern: "GET /debug/pprof/cmdline", handler: pprof.Cmdline},
			{pattern: "GET /debug/pprof/profile", handler: pprof.Profile},
			{pattern: "GET /debug/pprof/symbol", handler: pprof.Symbol},
			{pattern: "GET /debug/pprof/trace", handler: pprof.Trace},
			{pattern: "GET /debug/pprof/heap", handler: pprof.Handler("heap").ServeHTTP},
			{pattern: "GET /debug/pprof/goroutine", handler: pprof.Handler("goroutine").ServeHTTP},
			{pattern: "GET /debug/pprof/block", handler: pprof.Handler("block").ServeHTTP},
			{pattern: "GET /debug/pprof/threadcreate", handler: pprof.Handler("threadcreate").ServeHTTP},
		}

		for _, item := range pItems {
			rb.routeStack = append(
				rb.routeStack,
				routeDefinition{
					id:              getSequence(),
					pattern:         item.pattern,
					handler:         item.handler,
					middlewareStack: []Middleware{},
				},
			)
			setExtra(getSequence(), routeExtra{options: unset, sitemap: unset})
			nextSequence()
		}
		return nil
	}
}
