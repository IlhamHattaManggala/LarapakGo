package larapak

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// Logger returns a request logging middleware with execution timing.
func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			log.Printf("[LarapakGo] %s %s -> %d (%v)",
				c.Method(),
				c.Path(),
				c.statusCode,
				duration,
			)
			return err
		}
	}
}

// Recover handles panic exceptions gracefully without crashing the server.
func Recover() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[LarapakGo PANIC] %v\n%s", r, string(debug.Stack()))
					c.JSON(http.StatusInternalServerError, map[string]any{
						"error":   "Internal Server Error",
						"message": fmt.Sprintf("Kliru server internal: %v", r),
						"status":  http.StatusInternalServerError,
					})
				}
			}()
			return next(c)
		}
	}
}

// CORS returns a customizable Cross-Origin Resource Sharing middleware.
func CORS() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-CSRF-Token")

			if c.Method() == http.MethodOptions {
				c.Status(http.StatusNoContent)
				return nil
			}

			return next(c)
		}
	}
}

