package symple

type Middleware func(HandlerFunc) HandlerFunc

func chainMiddleware(h HandlerFunc, middlewareStack ...Middleware) HandlerFunc {
	if len(middlewareStack) == 0 {
		return h
	}
	for i := len(middlewareStack) - 1; i >= 0; i-- {
		h = middlewareStack[i](h)
	}
	return h
}

// WithMiddleware is used to add a custom middleware to the current Router
func (rs *routerState) WithMiddleware(middleware Middleware) routerOption {
	return func(rb *routerBuilder) error {
		rb.middlewareStack = append(rb.middlewareStack, middleware)
		return nil
	}
}
