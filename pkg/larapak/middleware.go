package larapak

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

func appendLogToFile(line string) {
	logDir := filepath.Join("storage", "logs")
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		_ = os.MkdirAll(logDir, 0755)
	}
	f, err := os.OpenFile(filepath.Join(logDir, "larapak.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		_, _ = f.WriteString(line + "\n")
	}
}

// Logger returns a request logging middleware with execution timing and file logging.
func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			start := time.Now()
			err := next(c)
			duration := time.Since(start)

			logMsg := fmt.Sprintf("[%s] [INFO] [LarapakGo] %s %s -> %d (%v) - IP: %s",
				start.Format("2006-01-02 15:04:05"),
				c.Method(),
				c.Path(),
				c.statusCode,
				duration,
				c.Request.RemoteAddr,
			)
			log.Print(logMsg)
			appendLogToFile(logMsg)
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
					panicMsg := fmt.Sprintf("[%s] [ERROR] [LarapakGo PANIC] %v\n%s",
						time.Now().Format("2006-01-02 15:04:05"),
						r,
						string(debug.Stack()),
					)
					log.Print(panicMsg)
					appendLogToFile(panicMsg)

					// Check if custom 500 error view exists
					if !c.IsAJAX() && !strings.HasPrefix(c.Path(), "/api/") {
						for _, p := range []string{
							filepath.Join("sawangan", "errors", "500.nview"),
							filepath.Join("resources", "views", "errors", "500.nview"),
							filepath.Join("tampilan", "errors", "500.nview"),
						} {
							if b, err := os.ReadFile(p); err == nil {
								html := strings.ReplaceAll(string(b), "{{ error }}", fmt.Sprintf("%v", r))
								html = strings.ReplaceAll(html, "{{ pesan }}", fmt.Sprintf("%v", r))
								_ = c.HTML(http.StatusInternalServerError, html)
								return
							}
						}
					}

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

