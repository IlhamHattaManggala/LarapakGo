package larapak

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// HandlerFunc defines the request handler signature in LarapakGo.
type HandlerFunc func(*Context) error

// MiddlewareFunc defines middleware signature wrapping next handler.
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// Context wraps http.ResponseWriter, http.Request, and helper methods.
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     map[string]string
	statusCode int
	store      map[string]any
}

// NewContext creates a new Context instance.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:     w,
		Request:    r,
		Params:     make(map[string]string),
		statusCode: http.StatusOK,
		store:      make(map[string]any),
	}
}

// Param retrieves a URL path parameter (e.g. :id or {slug}).
func (c *Context) Param(name string) string {
	return c.Params[name]
}

// Query retrieves a URL query string value.
func (c *Context) Query(key string) string {
	return c.Request.URL.Query().Get(key)
}

// QueryDefault retrieves a URL query string value with a fallback.
func (c *Context) QueryDefault(key, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// QueryParams returns parsed query values.
func (c *Context) QueryParams() url.Values {
	return c.Request.URL.Query()
}

// Status sets the HTTP response status code.
func (c *Context) Status(code int) *Context {
	c.statusCode = code
	return c
}

// Header sets a response header.
func (c *Context) Header(key, value string) *Context {
	c.Writer.Header().Set(key, value)
	return c
}

// Set stores a key-value pair in the request context lifecycle.
func (c *Context) Set(key string, val any) {
	c.store[key] = val
}

// Get retrieves a key from the request context lifecycle.
func (c *Context) Get(key string) (any, bool) {
	val, ok := c.store[key]
	return val, ok
}

// String sends a plain text response with the current status code.
func (c *Context) String(code int, format string, values ...any) error {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(code)
	c.statusCode = code
	_, err := fmt.Fprintf(c.Writer, format, values...)
	return err
}

// JSON sends a JSON response with status code and indentation.
func (c *Context) JSON(code int, data any) error {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Writer.WriteHeader(code)
	c.statusCode = code
	encoder := json.NewEncoder(c.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// HTML sends an HTML response.
func (c *Context) HTML(code int, htmlContent string) error {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(code)
	c.statusCode = code
	_, err := c.Writer.Write([]byte(htmlContent))
	return err
}

// BindJSON decodes the request body into a struct pointer.
func (c *Context) BindJSON(dest any) error {
	if c.Request.Body == nil {
		return fmt.Errorf("request body is empty")
	}
	defer c.Request.Body.Close()
	return json.NewDecoder(c.Request.Body).Decode(dest)
}

// Method returns the HTTP method.
func (c *Context) Method() string {
	return c.Request.Method
}

// Path returns the requested URL path.
func (c *Context) Path() string {
	return c.Request.URL.Path
}

// IsAJAX checks if the request is an XMLHttpRequest / fetch.
func (c *Context) IsAJAX() bool {
	return strings.ToLower(c.Request.Header.Get("X-Requested-With")) == "xmlhttprequest"
}
