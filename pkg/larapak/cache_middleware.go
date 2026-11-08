package larapak

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/cache"
)

// CachedResponse stores a snapshot of HTTP response for caching middleware.
type CachedResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type cacheResponseRecorder struct {
	http.ResponseWriter
	body        *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

func (r *cacheResponseRecorder) WriteHeader(code int) {
	if !r.wroteHeader {
		r.statusCode = code
		r.wroteHeader = true
		r.ResponseWriter.WriteHeader(code)
	}
}

func (r *cacheResponseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// CacheResponse creates middleware that caches GET/HEAD responses for the given TTL duration.
func CacheResponse(duration time.Duration) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			// Only cache GET and HEAD requests
			if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
				return next(c)
			}

			cacheKey := fmt.Sprintf("route_cache:%s:%s", c.Request.Method, c.Request.URL.RequestURI())
			cStore := cache.Default()

			if cachedVal, ok := cStore.Get(cacheKey); ok {
				if resp, ok := cachedVal.(CachedResponse); ok {
					for k, vList := range resp.Header {
						for _, v := range vList {
							c.Writer.Header().Set(k, v)
						}
					}
					c.Writer.Header().Set("X-Larapak-Cache", "HIT")
					c.Writer.WriteHeader(resp.StatusCode)
					_, err := c.Writer.Write(resp.Body)
					return err
				}
			}

			// Cache MISS: record response
			origWriter := c.Writer
			rec := &cacheResponseRecorder{
				ResponseWriter: origWriter,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}
			c.Writer = rec
			rec.Header().Set("X-Larapak-Cache", "MISS")

			err := next(c)
			if err != nil {
				return err
			}

			// Only cache successful 2xx responses
			if rec.statusCode >= 200 && rec.statusCode < 300 {
				toCache := CachedResponse{
					StatusCode: rec.statusCode,
					Header:     rec.Header().Clone(),
					Body:       rec.body.Bytes(),
				}
				cStore.Set(cacheKey, toCache, duration)
			}

			return nil
		}
	}
}

// SimpenCache is the Ngapak alias for CacheResponse.
func SimpenCache(durasi time.Duration) MiddlewareFunc {
	return CacheResponse(durasi)
}
