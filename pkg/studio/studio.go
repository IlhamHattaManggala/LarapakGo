package studio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// RouteInfo holds metadata about an individual registered route for Studio and OpenAPI.
type RouteInfo struct {
	Method string   `json:"method"`
	Path   string   `json:"path"`
	Target string   `json:"target"`
	Type   string   `json:"type"`
	Params []string `json:"params"`
}

// Studio provides the interactive Route & API Studio UI and OpenAPI 3.0 export.
type Studio struct {
	AppName string
	Version string
	AppEnv  string
	BaseURL string
	routes  []RouteInfo
}

// New creates a new Studio instance.
func New(appName, version, env, baseURL string) *Studio {
	if appName == "" {
		appName = "LarapakGo Application"
	}
	if version == "" {
		version = "1.2.2"
	}
	if env == "" {
		env = "local"
	}
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	return &Studio{
		AppName: appName,
		Version: version,
		AppEnv:  env,
		BaseURL: baseURL,
		routes:  make([]RouteInfo, 0),
	}
}

// RegisterRoute adds a route metadata item to Studio.
func (s *Studio) RegisterRoute(method, path, target string) {
	routeType := "API (JSON)"
	if strings.Contains(strings.ToLower(target), "view") || path == "/" {
		routeType = "VIEW (HTML)"
	}

	// Extract dynamic parameters
	var params []string
	segments := strings.Split(path, "/")
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			params = append(params, strings.TrimPrefix(seg, ":"))
		} else if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			params = append(params, strings.TrimSuffix(strings.TrimPrefix(seg, "{"), "}"))
		}
	}

	s.routes = append(s.routes, RouteInfo{
		Method: strings.ToUpper(method),
		Path:   path,
		Target: target,
		Type:   routeType,
		Params: params,
	})
}

// GenerateOpenAPI exports an OpenAPI 3.0.3 specification JSON.
func (s *Studio) GenerateOpenAPI() map[string]any {
	paths := make(map[string]any)

	for _, r := range s.routes {
		openApiPath := r.Path
		var openApiParams []map[string]any

		for _, p := range r.Params {
			openApiPath = strings.ReplaceAll(openApiPath, ":"+p, "{"+p+"}")
			openApiParams = append(openApiParams, map[string]any{
				"name":        p,
				"in":          "path",
				"required":    true,
				"description": fmt.Sprintf("Parameter dinamis '%s'", p),
				"schema":      map[string]any{"type": "string"},
			})
		}

		methodKey := strings.ToLower(r.Method)
		op := map[string]any{
			"summary":     fmt.Sprintf("Endpoint %s %s", r.Method, r.Path),
			"description": fmt.Sprintf("Ditangani dening: %s", r.Target),
			"operationId": fmt.Sprintf("%s_%s", methodKey, strings.ReplaceAll(strings.Trim(r.Path, "/"), "/", "_")),
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Respons sukses",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{"type": "object"},
						},
					},
				},
			},
		}

		if len(openApiParams) > 0 {
			op["parameters"] = openApiParams
		}

		if _, exists := paths[openApiPath]; !exists {
			paths[openApiPath] = make(map[string]any)
		}
		paths[openApiPath].(map[string]any)[methodKey] = op
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       fmt.Sprintf("%s - OpenAPI 3.0", s.AppName),
			"version":     s.Version,
			"description": "Dokumentasi API interaktif sing digawe kanthi otomatis dening LarapakGo Route & API Studio.",
		},
		"servers": []map[string]any{
			{"url": s.BaseURL, "description": fmt.Sprintf("Server Lingkungan %s", s.AppEnv)},
		},
		"paths": paths,
	}
}

// RenderHTML returns the complete Route & API Studio HTML document.
func (s *Studio) RenderHTML() string {
	routesJSON, _ := json.Marshal(s.routes)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>LarapakGo Route &amp; API Studio</title>
    <style>
        :root {
            --bg-base: #0B0C10;
            --bg-surface: #111218;
            --bg-elevated: #161822;
            --border-subtle: #1E202B;
            --border-highlight: #2A2D3D;
            --text-primary: #F8FAFC;
            --text-secondary: #94A3B8;
            --text-muted: #64748B;
            --color-amber: #F59E0B;
            --color-emerald: #10B981;
            --color-rose: #EF4444;
            --color-sky: #38BDF8;
            --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
            --font-sans: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: var(--font-sans); background: var(--bg-base); color: var(--text-primary); height: 100vh; display: flex; flex-direction: column; overflow: hidden; }
        .header { height: 56px; border-bottom: 1px solid var(--border-subtle); display: flex; align-items: center; justify-content: space-between; padding: 0 20px; background: var(--bg-surface); }
        .brand { display: flex; align-items: center; gap: 12px; }
        .logo-box { width: 28px; height: 28px; border-radius: 6px; background: rgba(245,158,11,0.15); border: 1px solid rgba(245,158,11,0.3); display: flex; align-items: center; justify-content: center; color: var(--color-amber); }
        .title { font-size: 14px; font-weight: 600; display: flex; align-items: center; gap: 8px; }
        .badge { font-size: 11px; padding: 2px 8px; border-radius: 9999px; font-weight: 500; font-family: var(--font-mono); }
        .badge-ver { background: #1E202B; color: var(--text-secondary); border: 1px solid var(--border-highlight); }
        .badge-env { background: rgba(16,185,129,0.12); color: var(--color-emerald); border: 1px solid rgba(16,185,129,0.25); display: flex; align-items: center; gap: 5px; }
        .dot { width: 6px; height: 6px; border-radius: 50%%; background: var(--color-emerald); }
        .main-layout { display: flex; flex: 1; overflow: hidden; }
        .sidebar { width: 340px; border-right: 1px solid var(--border-subtle); background: var(--bg-surface); display: flex; flex-direction: column; }
        .search-box { padding: 12px; border-bottom: 1px solid var(--border-subtle); }
        .search-input { width: 100%%; background: var(--bg-base); border: 1px solid var(--border-subtle); border-radius: 6px; padding: 8px 12px; color: var(--text-primary); font-size: 12px; outline: none; }
        .route-list { flex: 1; overflow-y: auto; padding: 8px; }
        .route-card { padding: 10px 12px; border-radius: 8px; cursor: pointer; border: 1px solid transparent; margin-bottom: 4px; transition: all 0.15s; }
        .route-card:hover { background: var(--bg-elevated); border-color: var(--border-subtle); }
        .route-card.active { background: rgba(245,158,11,0.08); border-color: rgba(245,158,11,0.3); }
        .route-card-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
        .method-badge { font-family: var(--font-mono); font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; }
        .m-GET { background: rgba(16,185,129,0.15); color: var(--color-emerald); }
        .m-POST { background: rgba(56,189,248,0.15); color: var(--color-sky); }
        .m-PUT, .m-PATCH { background: rgba(245,158,11,0.15); color: var(--color-amber); }
        .m-DELETE { background: rgba(239,68,68,0.15); color: var(--color-rose); }
        .route-path { font-family: var(--font-mono); font-size: 12px; font-weight: 500; color: var(--text-primary); }
        .route-target { font-size: 11px; color: var(--text-muted); }
        .content { flex: 1; display: flex; flex-direction: column; background: var(--bg-base); }
        .url-bar { padding: 16px 20px; border-bottom: 1px solid var(--border-subtle); display: flex; gap: 10px; align-items: center; background: var(--bg-surface); }
        .url-input { flex: 1; background: var(--bg-base); border: 1px solid var(--border-subtle); border-radius: 8px; padding: 10px 14px; color: var(--text-primary); font-family: var(--font-mono); font-size: 13px; outline: none; }
        .btn-send { background: var(--color-amber); color: #000; font-weight: 600; font-size: 13px; border: none; padding: 10px 18px; border-radius: 8px; cursor: pointer; display: flex; align-items: center; gap: 8px; }
        .btn-send:hover { opacity: 0.9; }
        .workspace { flex: 1; display: flex; overflow: hidden; }
        .panel { flex: 1; display: flex; flex-direction: column; border-right: 1px solid var(--border-subtle); }
        .panel-header { height: 42px; border-bottom: 1px solid var(--border-subtle); display: flex; align-items: center; padding: 0 16px; gap: 12px; background: var(--bg-surface); }
        .tab-btn { background: none; border: none; color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; padding: 6px 10px; border-radius: 4px; }
        .tab-btn.active { color: var(--text-primary); background: var(--bg-elevated); }
        .panel-body { flex: 1; padding: 20px; overflow-y: auto; }
        .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%%; color: var(--text-muted); text-align: center; gap: 12px; }
        .viewport-switch { display: flex; gap: 4px; margin-left: auto; }
        .vp-btn { background: var(--bg-base); border: 1px solid var(--border-subtle); color: var(--text-secondary); padding: 3px 8px; border-radius: 4px; font-size: 11px; cursor: pointer; }
        .vp-btn.active { border-color: var(--color-amber); color: var(--text-primary); }
    </style>
</head>
<body>
    <header class="header">
        <div class="brand">
            <div class="logo-box">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2">
                    <circle cx="18" cy="5" r="3"></circle>
                    <circle cx="6" cy="12" r="3"></circle>
                    <circle cx="18" cy="19" r="3"></circle>
                    <line x1="8.59" y1="13.51" x2="15.42" y2="17.49"></line>
                    <line x1="15.41" y1="6.51" x2="8.59" y2="10.49"></line>
                </svg>
            </div>
            <div class="title">
                LarapakGo Route &amp; API Studio
                <span class="badge badge-ver">v%s</span>
                <span class="badge badge-env"><span class="dot"></span>%s</span>
            </div>
        </div>
        <div>
            <a href="/_studio/openapi.json" target="_blank" style="color: var(--text-secondary); text-decoration: none; font-size: 12px; display: flex; align-items: center; gap: 6px;">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>
                OpenAPI 3.0
            </a>
        </div>
    </header>

    <div class="main-layout">
        <aside class="sidebar">
            <div class="search-box">
                <input type="text" id="filter" class="search-input" placeholder="Cari rute, path, method..." oninput="filterRoutes()">
            </div>
            <div class="route-list" id="routeList"></div>
        </aside>

        <main class="content">
            <div class="url-bar">
                <span class="method-badge m-GET" id="curMethod">GET</span>
                <input type="text" class="url-input" id="curUrl" value="%s/">
                <button class="btn-send" onclick="sendReq()">
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polygon points="5 3 19 12 5 21 5 3"></polygon></svg>
                    Kirim Permintaan
                </button>
            </div>

            <div class="workspace">
                <div class="panel">
                    <div class="panel-header">
                        <button class="tab-btn active">Parameter &amp; Query</button>
                        <button class="tab-btn">Headers</button>
                        <button class="tab-btn">Body</button>
                    </div>
                    <div class="panel-body" id="reqConfig">
                        <p style="font-size: 13px; color: var(--text-muted);">Pilih salah satu rute di sebelah kiwa kanggo miwiti simulasi request.</p>
                    </div>
                </div>

                <div class="panel">
                    <div class="panel-header">
                        <button class="tab-btn active">Pratinjau Live</button>
                        <button class="tab-btn">Raw Response</button>
                        <button class="tab-btn">Headers</button>
                        <div class="viewport-switch">
                            <button class="vp-btn active">Desktop</button>
                            <button class="vp-btn">Tablet</button>
                            <button class="vp-btn">Mobile</button>
                        </div>
                    </div>
                    <div class="panel-body">
                        <div class="empty-state">
                            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
                            <div style="font-weight: 500; color: var(--text-primary); font-size: 14px;">Pilih Rute &amp; Kirim Permintaan</div>
                            <div style="font-size: 12px; max-width: 320px;">Klik tombol Kirim Permintaan ing ndhuwur kanggo ndeleng respons live data utawa tampilan HTML.</div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <script>
        const routes = %s;
        function renderRoutes(list) {
            const el = document.getElementById('routeList');
            el.innerHTML = '';
            list.forEach((r, idx) => {
                const card = document.createElement('div');
                card.className = 'route-card' + (idx === 0 ? ' active' : '');
                card.innerHTML = `+"`"+`
                    <div class="route-card-top">
                        <span class="method-badge m-${r.method}">${r.method}</span>
                        <span style="font-size:10px; color:var(--text-muted);">${r.type}</span>
                    </div>
                    <div class="route-path">${r.path}</div>
                    <div class="route-target">${r.target}</div>
                `+"`"+`;
                card.onclick = () => selectRoute(r, card);
                el.appendChild(card);
            });
        }
        function selectRoute(r, card) {
            document.querySelectorAll('.route-card').forEach(c => c.classList.remove('active'));
            card.classList.add('active');
            document.getElementById('curMethod').textContent = r.method;
            document.getElementById('curMethod').className = 'method-badge m-' + r.method;
            document.getElementById('curUrl').value = '%s' + r.path;
        }
        function filterRoutes() {
            const q = document.getElementById('filter').value.toLowerCase();
            const filtered = routes.filter(r => r.path.toLowerCase().includes(q) || r.method.toLowerCase().includes(q) || r.target.toLowerCase().includes(q));
            renderRoutes(filtered);
        }
        function sendReq() {
            alert('Mengirim permintaan menyang rute: ' + document.getElementById('curUrl').value);
        }
        renderRoutes(routes);
    </script>
</body>
</html>`, s.Version, s.AppEnv, s.BaseURL, string(routesJSON), s.BaseURL)
}

// Handler returns the HTTP handler function for Studio requests.
func (s *Studio) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_studio/openapi.json" || r.URL.Path == "/studio/openapi.json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(s.GenerateOpenAPI())
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(s.RenderHTML()))
	}
}
