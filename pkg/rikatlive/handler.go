package rikatlive

import (
	"encoding/json"
	"net/http"

	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
)

// DefaultLivePath is the standard HTTP endpoint where RikatLive requests are posted.
const DefaultLivePath = "/_rikat/live"

// RegisterRoutes registers the RikatLive endpoint on a Larapak application router.
func RegisterRoutes(app *larapak.App, customPath ...string) {
	path := DefaultLivePath
	if len(customPath) > 0 && customPath[0] != "" {
		path = customPath[0]
	}

	app.POST(path, func(c *larapak.Context) error {
		var req RequestPayload
		if err := c.BindJSON(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "rikatlive: invalid request payload: " + err.Error(),
			})
		}

		res, err := ProcessRequest(req)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, res)
	})
}

// HTTPHandler returns a standard Go net/http handler for RikatLive requests.
func HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RequestPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "rikatlive: invalid request json", http.StatusBadRequest)
			return
		}

		res, err := ProcessRequest(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(res)
	}
}

