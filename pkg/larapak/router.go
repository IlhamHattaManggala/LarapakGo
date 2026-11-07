package larapak

import (
	"net/http"
	"strings"
)

type routeNode struct {
	pattern  string
	method   string
	handler  HandlerFunc
	handlers []MiddlewareFunc
}

// Router provides path routing with dynamic parameters and middleware support.
type Router struct {
	routes       []routeNode
	middlewares  []MiddlewareFunc
	prefix       string
	notFound     HandlerFunc
	methodNotAll HandlerFunc
}

// NewRouter constructs a clean Router instance.
func NewRouter() *Router {
	return &Router{
		routes:      make([]routeNode, 0),
		middlewares: make([]MiddlewareFunc, 0),
		notFound: func(c *Context) error {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Not Found",
				"message": "Rute ora ditemokna: " + c.Path(),
				"status":  http.StatusNotFound,
			})
		},
	}
}

// Use registers global or group middlewares.
func (r *Router) Use(m ...MiddlewareFunc) {
	r.middlewares = append(r.middlewares, m...)
}

// AddMiddlewareToLastRoute appends a middleware to the most recently registered route.
func (r *Router) AddMiddlewareToLastRoute(m MiddlewareFunc) {
	if len(r.routes) > 0 {
		r.routes[len(r.routes)-1].handlers = append(r.routes[len(r.routes)-1].handlers, m)
	}
}

// SetNotFoundHandler overrides the default 404 handler.
func (r *Router) SetNotFoundHandler(h HandlerFunc) {
	if h != nil {
		r.notFound = h
	}
}

// Group creates a nested route group with a shared URL prefix.
func (r *Router) Group(prefix string, fn func(*Router)) *Router {
	sub := &Router{
		routes:      nil,
		middlewares: append([]MiddlewareFunc(nil), r.middlewares...),
		prefix:      r.prefix + prefix,
		notFound:    r.notFound,
	}
	fn(sub)
	// Append sub-routes back to root router
	r.routes = append(r.routes, sub.routes...)
	return sub
}

// Handle registers a route with method, path pattern, and handler.
func (r *Router) Handle(method, path string, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	fullPath := r.prefix + path
	if !strings.HasPrefix(fullPath, "/") {
		fullPath = "/" + fullPath
	}

	allMiddlewares := append([]MiddlewareFunc(nil), r.middlewares...)
	allMiddlewares = append(allMiddlewares, middlewares...)

	r.routes = append(r.routes, routeNode{
		pattern:  fullPath,
		method:   strings.ToUpper(method),
		handler:  handler,
		handlers: allMiddlewares,
	})
}

// GET convenience shortcut.
func (r *Router) GET(path string, handler HandlerFunc, m ...MiddlewareFunc) {
	r.Handle(http.MethodGet, path, handler, m...)
}

// POST convenience shortcut.
func (r *Router) POST(path string, handler HandlerFunc, m ...MiddlewareFunc) {
	r.Handle(http.MethodPost, path, handler, m...)
}

// PUT convenience shortcut.
func (r *Router) PUT(path string, handler HandlerFunc, m ...MiddlewareFunc) {
	r.Handle(http.MethodPut, path, handler, m...)
}

// DELETE convenience shortcut.
func (r *Router) DELETE(path string, handler HandlerFunc, m ...MiddlewareFunc) {
	r.Handle(http.MethodDelete, path, handler, m...)
}

// PATCH convenience shortcut.
func (r *Router) PATCH(path string, handler HandlerFunc, m ...MiddlewareFunc) {
	r.Handle(http.MethodPatch, path, handler, m...)
}

// Match searches for a matching route and extracts parameters.
func (r *Router) match(method, path string) (*routeNode, map[string]string) {
	reqSegments := cleanSegments(path)

	for _, route := range r.routes {
		if route.method != method && !(method == http.MethodHead && route.method == http.MethodGet) {
			continue
		}

		routeSegments := cleanSegments(route.pattern)
		if len(reqSegments) != len(routeSegments) && !strings.Contains(route.pattern, "*") {
			continue
		}

		params := make(map[string]string)
		matched := true

		for i, rSeg := range routeSegments {
			if strings.HasPrefix(rSeg, "*") {
				paramName := strings.TrimPrefix(rSeg, "*")
				params[paramName] = strings.Join(reqSegments[i:], "/")
				matched = true
				break
			}

			if i >= len(reqSegments) {
				matched = false
				break
			}

			if strings.HasPrefix(rSeg, ":") {
				paramName := strings.TrimPrefix(rSeg, ":")
				params[paramName] = reqSegments[i]
			} else if strings.HasPrefix(rSeg, "{") && strings.HasSuffix(rSeg, "}") {
				paramName := strings.TrimSuffix(strings.TrimPrefix(rSeg, "{"), "}")
				params[paramName] = reqSegments[i]
			} else if rSeg != reqSegments[i] {
				matched = false
				break
			}
		}

		if matched {
			return &route, params
		}
	}

	return nil, nil
}

func cleanSegments(path string) []string {
	parts := strings.Split(path, "/")
	var res []string
	for _, p := range parts {
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

// RouteInfo provides metadata about a registered route.
type RouteInfo struct {
	Method  string
	Pattern string
}

// Routes returns a snapshot of all registered routes.
func (r *Router) Routes() []RouteInfo {
	res := make([]RouteInfo, len(r.routes))
	for i, rt := range r.routes {
		res[i] = RouteInfo{
			Method:  rt.method,
			Pattern: rt.pattern,
		}
	}
	return res
}

