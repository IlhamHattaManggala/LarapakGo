package larapak

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/session"
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

// MultipartForm parses and returns the multipart form (max 32MB by default).
func (c *Context) MultipartForm() (*multipart.Form, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	return c.Request.MultipartForm, nil
}

// FormFile returns the first file uploaded for the provided form key.
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Request.MultipartForm == nil {
		if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	f, fh, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	_ = f.Close()
	return fh, nil
}

// SaveUploadedFile saves an uploaded file header to the destination filepath.
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// FormValue retrieves a form value from POST body or query string.
func (c *Context) FormValue(key string) string {
	return c.Request.FormValue(key)
}

// PostForm retrieves a form value from the POST/PUT body.
func (c *Context) PostForm(key string) string {
	return c.Request.PostFormValue(key)
}

// PostFormDefault retrieves a POST form value with a fallback default.
func (c *Context) PostFormDefault(key, defaultValue string) string {
	val := c.PostForm(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// Redirect redirects the request to a given target URL with status code (default 302 Found).
func (c *Context) Redirect(code int, targetURL string) error {
	if code < 300 || code > 308 {
		code = http.StatusFound
	}
	c.statusCode = code
	http.Redirect(c.Writer, c.Request, targetURL, code)
	return nil
}

// AllInputs collects all request inputs (Query parameters, Form fields, and JSON body).
func (c *Context) AllInputs() map[string]any {
	inputs := make(map[string]any)

	// 1. URL Query parameters
	for k, v := range c.Request.URL.Query() {
		if len(v) == 1 {
			inputs[k] = v[0]
		} else {
			inputs[k] = v
		}
	}

	// 2. Form values (POST / PUT / multipart)
	_ = c.Request.ParseMultipartForm(32 << 20)
	for k, v := range c.Request.PostForm {
		if len(v) == 1 {
			inputs[k] = v[0]
		} else {
			inputs[k] = v
		}
	}

	// 3. JSON body if Content-Type is application/json
	if strings.Contains(c.Request.Header.Get("Content-Type"), "application/json") && c.Request.Body != nil {
		var jsonData map[string]any
		bodyBytes, err := io.ReadAll(c.Request.Body)
		if err == nil && len(bodyBytes) > 0 {
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			if json.Unmarshal(bodyBytes, &jsonData) == nil {
				for k, v := range jsonData {
					inputs[k] = v
				}
			}
		}
	}

	return inputs
}

// Session returns the active session container for this request.
func (c *Context) Session() *session.Session {
	return session.DefaultManager().Start(c.Writer, c.Request)
}

// FlashInputs stores all current request inputs into session flash for the next request.
func (c *Context) FlashInputs() {
	inputs := c.AllInputs()
	safeInputs := make(map[string]any, len(inputs))
	for k, v := range inputs {
		lower := strings.ToLower(k)
		if strings.Contains(lower, "password") || strings.Contains(lower, "sandi") || strings.Contains(lower, "secret") {
			continue
		}
		safeInputs[k] = v
	}
	if b, err := json.Marshal(safeInputs); err == nil {
		c.Session().Flash("_old_input", string(b))
	}
}

// AllOld retrieves all flashed old form inputs from the previous request.
func (c *Context) AllOld() map[string]any {
	raw := c.Session().GetFlash("_old_input")
	if raw == "" {
		return make(map[string]any)
	}
	var res map[string]any
	if err := json.Unmarshal([]byte(raw), &res); err != nil {
		return make(map[string]any)
	}
	return res
}

// Old retrieves a single flashed input value from previous request with optional default.
func (c *Context) Old(key string, fallback ...string) string {
	oldInputs := c.AllOld()
	if val, ok := oldInputs[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return ""
}

// Back redirects the user to the HTTP Referer URL, or to the provided fallback (default "/").
func (c *Context) Back(fallback ...string) error {
	ref := c.Request.Referer()
	if ref == "" {
		ref = "/"
		if len(fallback) > 0 && fallback[0] != "" {
			ref = fallback[0]
		}
	}
	return c.Redirect(http.StatusFound, ref)
}
