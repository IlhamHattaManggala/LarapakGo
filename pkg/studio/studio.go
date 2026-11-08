package studio

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/IlhamHattaManggala/larapak-go/pkg/database"
)

// RouteInfo holds metadata about an individual registered route for Studio and OpenAPI.
type RouteInfo struct {
	Method string   `json:"method"`
	Path   string   `json:"path"`
	Target string   `json:"target"`
	Type   string   `json:"type"`
	Params []string `json:"params"`
}

// Studio provides the interactive Route & API Studio UI, Database Inspector, and OpenAPI 3.0 export.
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
		version = "1.0.0"
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
			"description": "Dokumentasi API interaktif sing digawe otomatis dening LarapakGo Route & API Studio.",
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
        .nav-switch { display: flex; background: var(--bg-base); border: 1px solid var(--border-subtle); border-radius: 8px; padding: 3px; gap: 4px; }
        .nav-btn { background: none; border: none; color: var(--text-secondary); padding: 5px 12px; border-radius: 6px; font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.15s; }
        .nav-btn:hover { color: var(--text-primary); }
        .nav-btn.active { background: var(--bg-elevated); color: var(--color-amber); font-weight: 600; border: 1px solid var(--border-highlight); }
        .main-layout { display: flex; flex: 1; overflow: hidden; }
        .sidebar { width: 340px; border-right: 1px solid var(--border-subtle); background: var(--bg-surface); display: flex; flex-direction: column; }
        .search-box { padding: 12px; border-bottom: 1px solid var(--border-subtle); }
        .search-input { width: 100%%; background: var(--bg-base); border: 1px solid var(--border-subtle); border-radius: 6px; padding: 8px 12px; color: var(--text-primary); font-size: 12px; outline: none; }
        .route-list, .table-list { flex: 1; overflow-y: auto; padding: 8px; }
        .route-card, .table-card { padding: 10px 12px; border-radius: 8px; cursor: pointer; border: 1px solid transparent; margin-bottom: 4px; transition: all 0.15s; }
        .route-card:hover, .table-card:hover { background: var(--bg-elevated); border-color: var(--border-subtle); }
        .route-card.active, .table-card.active { background: rgba(245,158,11,0.08); border-color: rgba(245,158,11,0.3); }
        .route-card-top, .table-card-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px; }
        .method-badge { font-family: var(--font-mono); font-size: 10px; font-weight: 700; padding: 2px 6px; border-radius: 4px; }
        .m-GET { background: rgba(16,185,129,0.15); color: var(--color-emerald); }
        .m-POST { background: rgba(56,189,248,0.15); color: var(--color-sky); }
        .m-PUT, .m-PATCH { background: rgba(245,158,11,0.15); color: var(--color-amber); }
        .m-DELETE { background: rgba(239,68,68,0.15); color: var(--color-rose); }
        .route-path { font-family: var(--font-mono); font-size: 12px; font-weight: 500; color: var(--text-primary); }
        .route-target { font-size: 11px; color: var(--text-muted); }
        .content { flex: 1; display: flex; flex-direction: column; background: var(--bg-base); overflow: hidden; }
        .url-bar { padding: 14px 20px; border-bottom: 1px solid var(--border-subtle); display: flex; gap: 10px; align-items: center; background: var(--bg-surface); }
        .url-input { flex: 1; background: var(--bg-base); border: 1px solid var(--border-subtle); border-radius: 8px; padding: 9px 14px; color: var(--text-primary); font-family: var(--font-mono); font-size: 13px; outline: none; }
        .btn-send { background: var(--color-amber); color: #000; font-weight: 600; font-size: 13px; border: none; padding: 9px 18px; border-radius: 8px; cursor: pointer; display: flex; align-items: center; gap: 8px; }
        .btn-send:hover { opacity: 0.9; }
        .workspace { flex: 1; display: flex; overflow: hidden; }
        .panel { flex: 1; display: flex; flex-direction: column; border-right: 1px solid var(--border-subtle); }
        .panel-header { height: 42px; border-bottom: 1px solid var(--border-subtle); display: flex; align-items: center; padding: 0 16px; gap: 12px; background: var(--bg-surface); }
        .tab-btn { background: none; border: none; color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; padding: 6px 10px; border-radius: 4px; }
        .tab-btn.active { color: var(--text-primary); background: var(--bg-elevated); }
        .panel-body { flex: 1; padding: 20px; overflow-y: auto; }
        .empty-state { display: flex; flex-direction: column; align-items: center; justify-content: center; height: 100%%; color: var(--text-muted); text-align: center; gap: 12px; }
        .data-table-wrap { flex: 1; overflow: auto; padding: 16px; }
        table.db-table { width: 100%%; border-collapse: collapse; font-size: 12px; text-align: left; }
        table.db-table th { background: var(--bg-surface); color: var(--text-secondary); padding: 10px 12px; border: 1px solid var(--border-subtle); font-family: var(--font-mono); font-weight: 600; position: sticky; top: 0; }
        table.db-table td { padding: 8px 12px; border: 1px solid var(--border-subtle); color: var(--text-primary); font-family: var(--font-mono); }
        table.db-table tr:hover td { background: var(--bg-surface); }
        .db-header-bar { padding: 14px 20px; border-bottom: 1px solid var(--border-subtle); display: flex; align-items: center; justify-content: space-between; background: var(--bg-surface); }
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

        <div class="nav-switch">
            <button class="nav-btn active" id="btnModeRoutes" onclick="switchView('routes')">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"></polyline></svg>
                Route &amp; API Tester
            </button>
            <button class="nav-btn" id="btnModeDB" onclick="switchView('database')">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path></svg>
                Database Inspector
            </button>
            <button class="nav-btn" id="btnModeProfiler" onclick="switchView('profiler')">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>
                SQL Profiler
            </button>
        </div>

        <div>
            <a href="/_studio/openapi.json" target="_blank" style="color: var(--text-secondary); text-decoration: none; font-size: 12px; display: flex; align-items: center; gap: 6px;">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline><line x1="16" y1="13" x2="8" y2="13"></line><line x1="16" y1="17" x2="8" y2="17"></line></svg>
                OpenAPI 3.0
            </a>
        </div>
    </header>

    <!-- 1. VIEW: ROUTE & API TESTER -->
    <div class="main-layout" id="viewRoutes">
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
                    </div>
                    <div class="panel-body" id="resPanel">
                        <div class="empty-state">
                            <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"></circle><polygon points="16.24 7.76 14.12 14.12 7.76 16.24 9.88 9.88 16.24 7.76"></polygon></svg>
                            <div style="font-weight: 500; color: var(--text-primary); font-size: 14px;">Pilih Rute &amp; Kirim Request</div>
                            <div style="font-size: 12px; max-width: 320px;">Klik tombol Kirim Permintaan neng ndhuwur nggo ndeleng respons live data utawa tampilan HTML.</div>
                        </div>
                    </div>
                </div>
            </div>
        </main>
    </div>

    <!-- 2. VIEW: DATABASE INSPECTOR -->
    <div class="main-layout" id="viewDatabase" style="display: none;">
        <aside class="sidebar">
            <div class="search-box">
                <input type="text" id="dbFilter" class="search-input" placeholder="Cari tabel database..." oninput="filterTables()">
            </div>
            <div class="table-list" id="tableList">
                <div style="padding: 12px; font-size: 12px; color: var(--text-muted);">Ngunggah daftar tabel...</div>
            </div>
        </aside>

        <main class="content">
            <div class="db-header-bar">
                <div style="display: flex; align-items: center; gap: 10px;">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-amber);"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path></svg>
                    <span id="curTableName" style="font-weight: 600; font-size: 14px;">Pilih Tabel Database</span>
                    <span class="badge badge-ver" id="curTableRowCount">0 baris</span>
                </div>
                <div>
                    <button class="nav-btn" onclick="refreshTableData()" style="background: var(--bg-surface); border: 1px solid var(--border-subtle);">
                        <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                        Muat Ulang
                    </button>
                </div>
            </div>

            <div class="data-table-wrap" id="dbTableArea">
                <div class="empty-state">
                    <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><ellipse cx="12" cy="5" rx="9" ry="3"></ellipse><path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3"></path><path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path></svg>
                    <div style="font-weight: 500; color: var(--text-primary); font-size: 14px;">Pilih Tabel neng Sebelah Kiwa</div>
                    <div style="font-size: 12px; max-width: 340px;">Pilih salah siji tabel database nggo ndeleng struktur kolom lan isi data rekaman kanthi visual.</div>
                </div>
            </div>
        </main>
    </div>

    <!-- 3. VIEW: SQL PROFILER -->
    <div class="main-layout" id="viewProfiler" style="display: none; flex-direction: column;">
        <div class="db-header-bar">
            <div style="display: flex; align-items: center; gap: 10px;">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color: var(--color-sky);"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>
                <span style="font-weight: 600; font-size: 14px;">Database Query Profiler</span>
                <span class="badge badge-ver" id="profQueryCount">0 queries</span>
                <span class="badge badge-env" id="profTotalTime">0.00 ms</span>
            </div>
            <div style="display: flex; gap: 8px;">
                <button class="nav-btn" onclick="clearProfilerData()" style="background: var(--bg-surface); border: 1px solid var(--border-subtle); color: var(--color-rose);">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path></svg>
                    Resiki Log
                </button>
                <button class="nav-btn" onclick="loadProfilerData()" style="background: var(--bg-surface); border: 1px solid var(--border-subtle);">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"></polyline><polyline points="1 20 1 14 7 14"></polyline><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                    Muat Ulang
                </button>
            </div>
        </div>

        <div class="data-table-wrap" id="profTableArea">
            <div class="empty-state">
                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>
                <div style="font-weight: 500; color: var(--text-primary); font-size: 14px;">Durung Ana Query SQL sing Dicathet</div>
                <div style="font-size: 12px; max-width: 360px;">Saben query SELECT, INSERT, UPDATE, lan DELETE bakal otomatis dicathet lan diukur wektu eksekusine neng kene.</div>
            </div>
        </div>
    </div>

    <script>
        const routes = %s;
        let dbTables = [];
        let selectedTable = '';

        function switchView(mode) {
            document.getElementById('viewRoutes').style.display = 'none';
            document.getElementById('viewDatabase').style.display = 'none';
            document.getElementById('viewProfiler').style.display = 'none';
            document.getElementById('btnModeRoutes').classList.remove('active');
            document.getElementById('btnModeDB').classList.remove('active');
            document.getElementById('btnModeProfiler').classList.remove('active');

            if (mode === 'routes') {
                document.getElementById('viewRoutes').style.display = 'flex';
                document.getElementById('btnModeRoutes').classList.add('active');
            } else if (mode === 'database') {
                document.getElementById('viewDatabase').style.display = 'flex';
                document.getElementById('btnModeDB').classList.add('active');
                loadTables();
            } else if (mode === 'profiler') {
                document.getElementById('viewProfiler').style.display = 'flex';
                document.getElementById('btnModeProfiler').classList.add('active');
                loadProfilerData();
            }
        }

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
            const url = document.getElementById('curUrl').value;
            const method = document.getElementById('curMethod').textContent;
            const resPanel = document.getElementById('resPanel');
            resPanel.innerHTML = '<div style="font-size:12px; color:var(--text-muted);">Ngirim panjalukan...</div>';
            
            fetch(url, { method: method })
                .then(r => r.text().then(t => ({ status: r.status, ok: r.ok, text: t, type: r.headers.get('content-type') || '' })))
                .then(res => {
                    resPanel.innerHTML = `+"`"+`
                        <div style="margin-bottom:12px; display:flex; gap:8px;">
                            <span class="badge" style="background:${res.ok ? 'rgba(16,185,129,0.15)' : 'rgba(239,68,68,0.15)'}; color:${res.ok ? 'var(--color-emerald)' : 'var(--color-rose)'}">Status: ${res.status}</span>
                        </div>
                        <pre style="background:var(--bg-base); padding:12px; border-radius:6px; font-family:var(--font-mono); font-size:12px; overflow:auto; max-height:450px;">${escapeHtml(res.text)}</pre>
                    `+"`"+`;
                })
                .catch(err => {
                    resPanel.innerHTML = '<div style="color:var(--color-rose); font-size:12px;">Kliru: ' + err.message + '</div>';
                });
        }

        function loadTables() {
            fetch('/_studio/api/database/tables')
                .then(r => r.json())
                .then(d => {
                    dbTables = d.tables || [];
                    renderTables(dbTables);
                    if (dbTables.length > 0 && !selectedTable) {
                        selectTable(dbTables[0].name);
                    }
                })
                .catch(err => {
                    document.getElementById('tableList').innerHTML = '<div style="padding:12px; font-size:12px; color:var(--color-rose);">Gagal ngunggah tabel: ' + err.message + '</div>';
                });
        }

        function renderTables(list) {
            const el = document.getElementById('tableList');
            el.innerHTML = '';
            if (list.length === 0) {
                el.innerHTML = '<div style="padding:12px; font-size:12px; color:var(--text-muted);">Laka tabel database sing ditemokna.</div>';
                return;
            }
            list.forEach(t => {
                const card = document.createElement('div');
                card.className = 'table-card' + (t.name === selectedTable ? ' active' : '');
                card.innerHTML = `+"`"+`
                    <div class="table-card-top">
                        <span style="font-weight:600; font-size:13px; font-family:var(--font-mono); color:var(--text-primary);">${t.name}</span>
                        <span class="badge badge-ver">${t.rows} baris</span>
                    </div>
                `+"`"+`;
                card.onclick = () => {
                    document.querySelectorAll('.table-card').forEach(c => c.classList.remove('active'));
                    card.classList.add('active');
                    selectTable(t.name);
                };
                el.appendChild(card);
            });
        }

        function filterTables() {
            const q = document.getElementById('dbFilter').value.toLowerCase();
            renderTables(dbTables.filter(t => t.name.toLowerCase().includes(q)));
        }

        function selectTable(tableName) {
            selectedTable = tableName;
            document.getElementById('curTableName').textContent = tableName;
            refreshTableData();
        }

        function refreshTableData() {
            if (!selectedTable) return;
            const area = document.getElementById('dbTableArea');
            area.innerHTML = '<div style="font-size:12px; color:var(--text-muted);">Ngunggah data tabel...</div>';

            fetch('/_studio/api/database/table?name=' + encodeURIComponent(selectedTable))
                .then(r => r.json())
                .then(d => {
                    document.getElementById('curTableRowCount').textContent = (d.total || (d.rows ? d.rows.length : 0)) + ' baris';
                    if (!d.rows || d.rows.length === 0) {
                        area.innerHTML = '<div class="empty-state"><div style="color:var(--text-muted); font-size:13px;">Tabel kosong (durung ana data).</div></div>';
                        return;
                    }
                    let html = '<table class="db-table"><thead><tr>';
                    d.columns.forEach(col => { html += `+"`"+`<th>${col}</th>`+"`"+`; });
                    html += '</tr></thead><tbody>';
                    d.rows.forEach(r => {
                        html += '<tr>';
                        d.columns.forEach(col => {
                            const val = r[col] !== undefined && r[col] !== null ? escapeHtml(String(r[col])) : '<span style="color:var(--text-muted); font-style:italic;">NULL</span>';
                            html += `+"`"+`<td>${val}</td>`+"`"+`;
                        });
                        html += '</tr>';
                    });
                    html += '</tbody></table>';
                    area.innerHTML = html;
                })
                .catch(err => {
                    area.innerHTML = '<div style="color:var(--color-rose); font-size:12px;">Gagal maca isi tabel: ' + err.message + '</div>';
                });
        }

        function loadProfilerData() {
            const area = document.getElementById('profTableArea');
            fetch('/_studio/api/queries')
                .then(r => r.json())
                .then(d => {
                    const queries = d.queries || [];
                    document.getElementById('profQueryCount').textContent = (d.count || queries.length) + ' queries';
                    document.getElementById('profTotalTime').textContent = (d.total_ms ? d.total_ms.toFixed(2) : '0.00') + ' ms';

                    if (queries.length === 0) {
                        area.innerHTML = `+"`"+`
                            <div class="empty-state">
                                <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"></path></svg>
                                <div style="font-weight: 500; color: var(--text-primary); font-size: 14px;">Durung Ana Query SQL sing Dicathet</div>
                                <div style="font-size: 12px; max-width: 360px;">Saben query SELECT, INSERT, UPDATE, lan DELETE bakal otomatis dicathet lan diukur wektu eksekusine neng kene.</div>
                            </div>
                        `+"`"+`;
                        return;
                    }

                    let html = '<table class="db-table"><thead><tr>';
                    html += '<th style="width: 100px;">Duration</th><th style="width: 130px;">Wektu</th><th>SQL Query</th><th style="width: 180px;">Bindings</th><th style="width: 80px;">Status</th>';
                    html += '</tr></thead><tbody>';

                    queries.forEach(q => {
                        const durColor = q.duration_ms > 50 ? 'var(--color-rose)' : (q.duration_ms > 10 ? 'var(--color-amber)' : 'var(--color-emerald)');
                        const statusBadge = q.error ? '<span class="badge" style="background:rgba(239,68,68,0.15);color:var(--color-rose);">ERROR</span>' : '<span class="badge" style="background:rgba(16,185,129,0.15);color:var(--color-emerald);">OK</span>';
                        const timeStr = q.timestamp ? q.timestamp.substring(11, 19) : '-';
                        const bindingsStr = q.bindings && q.bindings.length > 0 ? escapeHtml(JSON.stringify(q.bindings)) : '<span style="color:var(--text-muted);">-</span>';

                        html += '<tr>';
                        html += '<td style="font-weight:600; color:' + durColor + ';">' + (q.duration_ms ? q.duration_ms.toFixed(2) : '0.00') + ' ms</td>';
                        html += '<td style="color:var(--text-secondary);">' + timeStr + '</td>';
                        html += '<td style="color:var(--text-primary); word-break:break-all;"><code>' + escapeHtml(q.sql) + '</code></td>';
                        html += '<td style="color:var(--text-secondary); font-size:11px;">' + bindingsStr + '</td>';
                        html += '<td>' + statusBadge + '</td>';
                        html += '</tr>';
                    });

                    html += '</tbody></table>';
                    area.innerHTML = html;
                })
                .catch(err => {
                    area.innerHTML = '<div style="color:var(--color-rose); font-size:12px;">Gagal ngunggah data query profiler: ' + err.message + '</div>';
                });
        }

        function clearProfilerData() {
            fetch('/_studio/api/queries/clear', { method: 'POST' })
                .then(r => r.json())
                .then(() => {
                    loadProfilerData();
                })
                .catch(err => {
                    alert('Gagal ngresiki data profiler: ' + err.message);
                });
        }

        function escapeHtml(str) {
            return str.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
        }

        renderRoutes(routes);
    </script>
</body>
</html>`, s.Version, s.AppEnv, s.BaseURL, string(routesJSON), s.BaseURL)
}

// Handler returns the HTTP handler function for Studio requests and API endpoints.
func (s *Studio) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/_studio/openapi.json" || path == "/studio/openapi.json" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			_ = json.NewEncoder(w).Encode(s.GenerateOpenAPI())
			return
		}

		if path == "/_studio/api/queries" || path == "/studio/api/queries" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			prof := database.GetProfiler()
			count, totalMS := prof.Stats()
			_ = json.NewEncoder(w).Encode(map[string]any{
				"queries":  prof.Queries(),
				"count":    count,
				"total_ms": totalMS,
			})
			return
		}

		if (path == "/_studio/api/queries/clear" || path == "/studio/api/queries/clear") && r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			database.GetProfiler().Clear()
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
			return
		}

		if path == "/_studio/api/database/tables" || path == "/studio/api/database/tables" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			tables, err := InspectTables()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error(), "tables": []any{}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"tables": tables})
			return
		}

		if path == "/_studio/api/database/table" || path == "/studio/api/database/table" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			tableName := r.URL.Query().Get("name")
			if tableName == "" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "parameter 'name' dibutuhna"})
				return
			}
			data, err := InspectTableData(tableName, 50, 0)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": err.Error()})
				return
			}
			_ = json.NewEncoder(w).Encode(data)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(s.RenderHTML()))
	}
}
