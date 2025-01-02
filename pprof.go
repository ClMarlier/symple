package symple

import (
	"net/http"
	"net/http/pprof"
)

type pprofItems struct {
	pattern string
	handler HandlerFunc
}

func (rs *routerState) WithPprof() routerOption {
	return func(rb *routerBuilder) error {
		pItems := []pprofItems{
			{pattern: "GET /debug/pprof",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Index(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/cmdline",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Cmdline(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/profile",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Profile(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/symbol",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Symbol(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/trace",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Trace(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/heap",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Handler("heap").ServeHTTP(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/goroutine",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Handler("goroutine").ServeHTTP(w, r)
					return nil
				},
			},
			{pattern: "GET /debug/pprof/block", handler: func(w http.ResponseWriter, r *http.Request) error {
				pprof.Handler("block").ServeHTTP(w, r)
				return nil
			},
			},
			{pattern: "GET /debug/pprof/threadcreate",
				handler: func(w http.ResponseWriter, r *http.Request) error {
					pprof.Handler("threadcreate").ServeHTTP(w, r)
					return nil
				},
			},
		}

		for _, item := range pItems {
			rb.routeStack = append(
				rb.routeStack,
				routeDefinition{
					id:              rs.getSequence(),
					pattern:         item.pattern,
					handler:         item.handler,
					middlewareStack: []Middleware{},
				},
			)
			rs.setExtra(rs.getSequence(), routeExtra{options: unset, sitemap: unset})
			rs.nextSequence()
		}
		return nil
	}
}
