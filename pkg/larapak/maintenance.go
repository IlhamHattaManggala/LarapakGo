package larapak

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MaintenanceData holds state and configuration when the application is down.
type MaintenanceData struct {
	Time    int64  `json:"time"`
	Message string `json:"message"`
	Secret  string `json:"secret"`
	Retry   int    `json:"retry"`
}

// DefaultDownFilePath returns the standard location of the maintenance lock file.
func DefaultDownFilePath() string {
	return filepath.Join("storage", "framework", "down")
}

// IsDown checks if the application is currently in maintenance mode.
func IsDown(customPath ...string) (bool, *MaintenanceData) {
	p := DefaultDownFilePath()
	if len(customPath) > 0 && customPath[0] != "" {
		p = customPath[0]
	}

	bytes, err := os.ReadFile(p)
	if err != nil {
		return false, nil
	}

	var data MaintenanceData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return true, &MaintenanceData{
			Time:    time.Now().Unix(),
			Message: "Aplikasi lagi didandani (maintenance mode).",
			Retry:   60,
		}
	}
	return true, &data
}

// Down enables maintenance mode by writing the down lock file.
func Down(secret, message string, retry int, customPath ...string) (*MaintenanceData, error) {
	p := DefaultDownFilePath()
	if len(customPath) > 0 && customPath[0] != "" {
		p = customPath[0]
	}

	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}

	if secret == "" {
		bytes := make([]byte, 16)
		if _, err := rand.Read(bytes); err == nil {
			secret = hex.EncodeToString(bytes)
		}
	}

	if message == "" {
		message = "Aplikasi lagi didandani (maintenance mode)."
	}
	if retry <= 0 {
		retry = 60
	}

	data := &MaintenanceData{
		Time:    time.Now().Unix(),
		Message: message,
		Secret:  secret,
		Retry:   retry,
	}

	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return data, os.WriteFile(p, encoded, 0644)
}

// Up disables maintenance mode by removing the down lock file.
func Up(customPath ...string) error {
	p := DefaultDownFilePath()
	if len(customPath) > 0 && customPath[0] != "" {
		p = customPath[0]
	}

	err := os.Remove(p)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Maintenance returns a middleware that blocks incoming HTTP requests when the app is in maintenance mode.
func Maintenance(customPath ...string) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			isDown, data := IsDown(customPath...)
			if !isDown {
				return next(c)
			}

			// Check bypass token in query param or cookie
			bypassToken := c.Query("bypass")
			if bypassToken == "" {
				bypassToken = c.Query("secret")
			}

			cookieVal := ""
			if cookie, err := c.Request.Cookie("larapak_maintenance"); err == nil {
				cookieVal = cookie.Value
			}

			// If valid secret provided, allow bypass
			if data.Secret != "" && (bypassToken == data.Secret || cookieVal == data.Secret) {
				// Set cookie if bypassing via URL query parameter
				if bypassToken == data.Secret && cookieVal != data.Secret {
					http.SetCookie(c.Writer, &http.Cookie{
						Name:     "larapak_maintenance",
						Value:    data.Secret,
						Path:     "/",
						HttpOnly: true,
						MaxAge:   86400,
					})
				}
				return next(c)
			}

			// Respond with 503 Service Unavailable
			retrySeconds := data.Retry
			if retrySeconds <= 0 {
				retrySeconds = 60
			}
			c.Writer.Header().Set("Retry-After", fmt.Sprintf("%d", retrySeconds))

			// Check if API / JSON request
			accept := c.Request.Header.Get("Accept")
			if strings.Contains(accept, "application/json") || strings.HasPrefix(c.Path(), "/api") {
				return c.JSON(http.StatusServiceUnavailable, map[string]any{
					"error":       "Service Unavailable",
					"message":     data.Message,
					"retry_after": retrySeconds,
				})
			}

			// Clean, high-end HTML 503 page with SVG icon (zero emojis)
			htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>503 - Sedang Didandani | LarapakGo</title>
    <style>
        :root {
            --bg-base: #0B0C10;
            --bg-surface: #12131A;
            --border-subtle: #1E202B;
            --text-primary: #F8FAFC;
            --text-secondary: #94A3B8;
            --color-amber: #F59E0B;
            --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background-color: var(--bg-base);
            color: var(--text-primary);
            font-family: var(--font-sans);
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            padding: 24px;
        }
        .card {
            background: var(--bg-surface);
            border: 1px solid var(--border-subtle);
            border-radius: 12px;
            max-width: 480px;
            width: 100%%;
            padding: 36px 32px;
            text-align: center;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.4);
        }
        .icon-wrap {
            width: 56px;
            height: 56px;
            background: rgba(245, 158, 11, 0.12);
            border: 1px solid rgba(245, 158, 11, 0.25);
            border-radius: 12px;
            display: inline-flex;
            align-items: center;
            justify-content: center;
            color: var(--color-amber);
            margin-bottom: 20px;
        }
        h1 {
            font-size: 20px;
            font-weight: 700;
            margin-bottom: 10px;
            letter-spacing: -0.02em;
        }
        p {
            color: var(--text-secondary);
            font-size: 14px;
            line-height: 1.6;
            margin-bottom: 24px;
        }
        .badge {
            display: inline-block;
            font-family: var(--font-mono);
            font-size: 11px;
            font-weight: 600;
            padding: 4px 10px;
            border-radius: 9999px;
            background: #1E202B;
            color: var(--text-secondary);
            border: 1px solid #2A2D3D;
        }
    </style>
</head>
<body>
    <div class="card">
        <div class="icon-wrap">
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"></path>
            </svg>
        </div>
        <h1>503 Sedang Diopeni</h1>
        <p>%s</p>
        <span class="badge">HTTP 503 &bull; Retry in %ds</span>
    </div>
</body>
</html>`, data.Message, retrySeconds)

			c.Writer.WriteHeader(http.StatusServiceUnavailable)
			c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = c.Writer.Write([]byte(htmlContent))
			return nil
		}
	}
}

