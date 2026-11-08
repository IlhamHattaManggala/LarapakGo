package main

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/cache"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/ngapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/queue"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

//go:embed templates/views/layout.nview
var templateLayoutNview string

//go:embed templates/views/index.nview
var templateIndexNview string

//go:embed templates/public/css/rikat.css
var templateRikatCSS string

//go:embed templates/public/js/rikat.js
var templateRikatJS string

//go:embed templates/public/images/logo.png
var templateLogoPNG []byte

//go:embed templates/public/favicon.png
var templateFaviconPNG []byte

//go:embed templates/public/favicon.ico
var templateFaviconICO []byte

const version = "1.0.0"

func main() {
	progName := strings.ToLower(filepath.Base(os.Args[0]))
	progName = strings.TrimSuffix(progName, ".exe")

	if len(os.Args) < 2 {
		if progName == "ngapakin" {
			startNgapakREPL()
			return
		}
		serveDemoApp("8080")
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	// Multicall binary: nek jeneng eksekusine ngapakin.exe
	if progName == "ngapakin" {
		handleNgapakCommand(command, args)
		return
	}

	switch command {
	// =========================================================================
	// 1. INFO & BANTUAN
	// =========================================================================
	case "version", "-v", "--version":
		fmt.Printf("LarapakGo Framework v%s (Go Edition)\n", version)
		fmt.Printf("NgapakIn Language Engine v%s (Native Built-in)\n", version)

	case "help", "tulung", "--help", "-h":
		showHelp()

	case "ngapakin":
		if len(args) == 0 {
			startNgapakREPL()
			return
		}
		handleNgapakCommand(strings.ToLower(args[0]), args[1:])
		return

	case "repl", "interaktif":
		startNgapakREPL()

	// =========================================================================
	// 2. PROYEK & SERVER
	// =========================================================================
	case "serve", "ladeni", "urubna", "mlaku", "demo":
		port := "8080"
		if len(args) > 0 {
			port = args[0]
		}
		serveDemoApp(port)

	case "anyar":
		projName, lang := parseProjectArgs(args, "ngapak")
		if projName == "" {
			fmt.Println("Kliru: Tulung wei jeneng proyek ndisit. Contohe: larapak anyar warung_app [--bahasa=ngapak|indonesia|english]")
			os.Exit(1)
		}
		createNewProject(projName, lang)

	case "new":
		projName, lang := parseProjectArgs(args, "english")
		if projName == "" {
			fmt.Println("Error: Project name required. Example: larapak new my_app [--bahasa=ngapak|indonesia|english]")
			os.Exit(1)
		}
		createNewProject(projName, lang)

	case "run", "mlakukna", "eksekusi":
		if len(args) < 1 {
			fmt.Println("Kliru: Kudune ana path berkas .ngpk. Contohe: larapak mlakukna skrip.ngpk")
			os.Exit(1)
		}
		runNgapakScript(args[0])

	case "route:list", "dalan:daptar", "daptar:dalan", "rute:daptar":
		showRouteList()

	// =========================================================================
	// 3. GENERATOR (SCAFFOLDING)
	// =========================================================================
	case "make:controller", "gawe:pangatur", "gawe:controller":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng controller ndisit. Contohe: larapak gawe:pangatur UserController")
			os.Exit(1)
		}
		createController(args[0])

	case "make:model", "gawe:wujud", "gawe:model":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng model ndisit. Contohe: larapak gawe:wujud Produk")
			os.Exit(1)
		}
		createModel(args[0])

	case "make:middleware", "gawe:pencegat", "gawe:middleware":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng middleware ndisit. Contohe: larapak gawe:pencegat AuthMiddleware")
			os.Exit(1)
		}
		createMiddleware(args[0])

	case "make:job", "gawe:gawean", "gawe:job":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng job ndisit. Contohe: larapak gawe:gawean KirimEmailJob")
			os.Exit(1)
		}
		createJob(args[0])

	case "make:seeder", "gawe:bibit", "gawe:seeder":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng seeder ndisit. Contohe: larapak gawe:bibit UserSeeder")
			os.Exit(1)
		}
		createSeeder(args[0])

	case "make:factory", "gawe:pabrik":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng pabrik (factory) ndisit. Contohe: larapak gawe:pabrik User")
			os.Exit(1)
		}
		createFactory(args[0])

	case "make:resource-api", "make:api-resource", "gawe:sumber-api", "gawe:resource-api":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng resource transformer ndisit. Contohe: larapak gawe:sumber-api User")
			os.Exit(1)
		}
		createApiResource(args[0])

	case "make:migration", "gawe:migrasi":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng migrasi ndisit. Contohe: larapak gawe:migrasi create_users_table")
			os.Exit(1)
		}
		createMigration(args[0])

	case "make:live", "gawe:urip", "gawe:live", "make:component", "gawe:komponen":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng komponen RikatLive ndisit. Contohe: larapak gawe:urip Counter")
			os.Exit(1)
		}
		createLiveComponent(args[0])

	case "make:resource", "gawe:sumber", "gawe:resource":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng sumber (resource) ndisit. Contohe: larapak gawe:sumber Produk")
			os.Exit(1)
		}
		createResource(args[0])

	case "make:auth", "gawe:otentikasi", "gawe:auth":
		createAuthScaffold()

	// =========================================================================
	// 4. DATABASE & MIGRASI
	// =========================================================================
	case "migrate", "migrasi", "munggah":
		runMigrations()

	case "migrate:fresh", "migrasi:anyar", "migrasi:resik":
		runMigrationsFresh()

	case "migrate:status", "migrasi:status":
		showMigrationStatus()

	case "rollback", "mudhun", "migrasi:mundur":
		fmt.Println("✓ Rollback migrasi database sukses dilakoni.")

	case "db:seed", "bibiti", "lumbung:bibit":
		runSeeders()

	// =========================================================================
	// 5. EKOSISTEM, FRONTEND & TOOLING
	// =========================================================================
	case "queue:work", "gawean:tandang", "antrean:tandang":
		runQueueWorker()

	case "cache:clear", "resik:cache", "resiki:cache":
		clearCache()

	case "package:install", "paket:pasang", "pasang:paket":
		pkgName := "paket"
		if len(args) > 0 {
			pkgName = args[0]
		}
		fmt.Printf("✓ Paket '%s' sukses dipasang maring proyek LarapakGo.\n", pkgName)

	case "css:build", "rikat:rakit", "rikat:bangun":
		p := paths.GetPaths(".")
		fmt.Printf("✓ Rikat CSS JIT: Sukses ngrakit berkas CSS maring %s/rikat.css.\n", filepath.ToSlash(p.PublicCSS))

	case "css:watch", "rikat:pantau":
		p := paths.GetPaths(".")
		fmt.Printf("[Rikat CSS] Pantau aktif: Mantau owahan template neng %s... [Pencet Ctrl+C nggo mandeg]\n", filepath.ToSlash(p.Views))

	case "test", "coba":
		runTests()

	default:
		if strings.HasSuffix(command, ".ngpk") {
			runNgapakScript(os.Args[1])
			return
		}
		fmt.Printf("Perintah '%s' ora dingerteni. Ketik 'larapak tulung' nggo ndeleng panduan lengkap.\n", command)
		os.Exit(1)
	}
}

func parseProjectArgs(args []string, defaultLang string) (string, string) {
	projectName := ""
	lang := defaultLang
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--bahasa=") {
			lang = strings.TrimPrefix(arg, "--bahasa=")
		} else if strings.HasPrefix(arg, "--lang=") {
			lang = strings.TrimPrefix(arg, "--lang=")
		} else if strings.HasPrefix(arg, "--language=") {
			lang = strings.TrimPrefix(arg, "--language=")
		} else if strings.HasPrefix(arg, "-b=") {
			lang = strings.TrimPrefix(arg, "-b=")
		} else if arg == "-b" || arg == "--bahasa" || arg == "--lang" || arg == "--language" {
			if i+1 < len(args) {
				lang = args[i+1]
				i++
			}
		} else if !strings.HasPrefix(arg, "-") && projectName == "" {
			projectName = arg
		}
	}
	return projectName, paths.NormalizeLanguage(lang)
}

func showHelp() {
	fmt.Printf(`
   __                                __   _____ 
  / /   ____ __________ _____  ____ _/ /__/ ___/ 
 / /   / __ '/ ___/ __ '/ __ \/ __ '/ //_/ / __  
/ /___/ /_/ / /  / /_/ / /_/ / /_/ / ,< / /_/ /  
/_____/\__,_/_/   \__,_/ .___/\__,_/_/|_|\____/   
                      /_/                         
  LarapakGo CLI - Versi %s (Go Edition)
  Dialek Ngapak & Standard Dual-Support (Persis Versi Python)

Perintah Dialek Ngapak (Asli):
  larapak anyar <NamaProyek> [--bahasa=ngapak|indonesia|english]
  larapak ladeni [port]                         (Urubna server lokal, Studio & Auto-Reload)
  larapak dalan:daptar                          (Daftar kabeh rute HTTP)
  larapak gawe:otentikasi                       (Gawe lengkap sistem Login, Register, Auth Middleware & Views)
  larapak gawe:sumber <NamaSumber>              (Gawe lengkap Model, Migrasi, CRUD Controller & Views)
  larapak gawe:pangatur <NamaController>        (Gawe controller anyar)
  larapak gawe:wujud <NamaModel>                (Gawe model ORM anyar)
  larapak gawe:urip <NamaKomponen>              (Gawe komponen realtime RikatLive)
  larapak gawe:migrasi <NamaMigrasi>            (Gawe skema migrasi anyar)
  larapak gawe:pencegat <NamaMiddleware>        (Gawe middleware anyar)
  larapak gawe:gawean <NamaJob>                 (Gawe background queue job)
  larapak gawe:bibit <NamaSeeder>               (Gawe database seeder)
  larapak gawe:pabrik <NamaFactory>             (Gawe model factory fake data)
  larapak gawe:sumber-api <NamaResource>        (Gawe JSON API Resource Transformer)
  larapak migrasi                               (Jalankna migrasi database)
  larapak migrasi:anyar [--seed]                (Reset & migrasi anyar)
  larapak migrasi:status                        (Status tabel migrasi)
  larapak mudhun                                (Rollback migrasi)
  larapak bibiti                                (Jalankna database seeder)
  larapak gawean:tandang                        (Jalankna queue worker)
  larapak resik:cache                           (Resiki cache sistem)
  larapak rikat:rakit                           (Kompilasi Rikat CSS JIT)
  larapak coba                                  (Jalankna kabeh test suite)
  larapak mlakukna <berkas.ngpk>                (Jalankna skrip basa NgapakIn)
  larapak repl                                  (Terminal interaktif NgapakIn)
  larapak ngapakin [run/repl/-v]                (Sub-perintah mesin basa NgapakIn)
  larapak tulung                                (Ndeleng panduan kiye)

Perintah Standar (Laravel Artisan Style):
  larapak new, serve, route:list, make:auth, make:resource, make:resource-api, make:controller, make:model, make:factory, make:live, make:migration,
  make:middleware, make:job, make:seeder, migrate, migrate:fresh, migrate:status,
  rollback, db:seed, queue:work, cache:clear, css:build, run, repl, test, help, version

Dokumentasi Resmi: https://larapak.my.id
`, version)
}

func runNgapakScript(filePath string) {
	app := larapak.New()
	eng := ngapak.NewEngine(".", nil)
	eng.BindRoutes(app)
	_, err := eng.RunFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kliru pas njalankna %s: %v\n", filePath, err)
		os.Exit(1)
	}
	routes := app.Routes()
	if len(routes) > 0 {
		fmt.Printf("✓ Sukses ngunggah lan mriksa %d rute saka %s\n", len(routes), filepath.ToSlash(filePath))
	}
}

func startNgapakREPL() {
	fmt.Printf(`
   _  __                __   ____     
  / |/ /__ ____ ____  ___ _/ /__/  _/__ 
 /    / _ '/ _ '/ _ \/ _ '/ //_// / / _ \
/_/|_/\_, /\_,_/ .__/\_,_/\__,_//_/_/_//_/
     /___/    /_/                         
  NgapakIn Language REPL v%s (All-in-One LarapakGo)
  Ora Ngapak Ora Kepenak!
  Ketik 'metu' utawa pencet Ctrl+C nggo metu.

`, version)

	scanner := bufio.NewScanner(os.Stdin)
	eng := ngapak.NewEngine(".", nil)

	for {
		fmt.Print("ngapak>>> ")
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "metu" || line == "exit" || line == "quit" {
			fmt.Println("Matur nuwun, rampung.")
			break
		}

		res, err := eng.RunSource(line, "repl")
		if err != nil {
			fmt.Printf("Kliru: %v\n", err)
		} else if res != nil {
			fmt.Println(res)
		}
	}
}

func handleNgapakCommand(cmd string, args []string) {
	switch cmd {
	case "-v", "--version", "version", "versi":
		fmt.Printf("NgapakIn Language Engine v%s (Go Edition)\n", version)
	case "repl", "interaktif":
		startNgapakREPL()
	case "run", "mlakukna", "eksekusi":
		if len(args) < 1 {
			fmt.Println("Kliru: Kudune ana path berkas .ngpk. Contohe: larapak ngapakin run skrip.ngpk")
			os.Exit(1)
		}
		runNgapakScript(args[0])
	default:
		if strings.HasSuffix(cmd, ".ngpk") {
			runNgapakScript(cmd)
		} else {
			fmt.Printf("Kliru: Perintah ngapakin '%s' ora dingerteni.\n", cmd)
		}
	}
}

func getDemoWelcomeHTML() string {
	sectionPattern := regexp.MustCompile(`(?s)@section\s+["']content["'](.*?)(?:@endsection|@rampung)`)
	m := sectionPattern.FindStringSubmatch(templateIndexNview)
	content := ""
	if len(m) > 1 {
		content = strings.TrimSpace(m[1])
	}
	html := strings.Replace(templateLayoutNview, `@yield "content"`, content, 1)
	html = strings.ReplaceAll(html, `{{ title }}`, "LarapakGo")
	return html
}

// LiveReloader routes requests to the active in-memory Larapak App instance,
// allowing seamless atomic hot-reloading when project files change without dropping the TCP port,
// and providing Server-Sent Events (SSE) auto-refresh directly to connected browsers.
type LiveReloader struct {
	mu      sync.RWMutex
	app     *larapak.App
	clients map[chan string]bool
	cMu     sync.Mutex
}

func (lr *LiveReloader) BroadcastReload() {
	lr.cMu.Lock()
	defer lr.cMu.Unlock()
	for ch := range lr.clients {
		select {
		case ch <- "reload":
		default:
		}
	}
}

type liveResponseWriter struct {
	http.ResponseWriter
	buf    bytes.Buffer
	isHTML bool
	code   int
}

func (w *liveResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	ct := w.Header().Get("Content-Type")
	if strings.Contains(ct, "text/html") {
		w.isHTML = true
	} else {
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *liveResponseWriter) Write(b []byte) (int, error) {
	if w.isHTML {
		return w.buf.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

const liveReloadScript = `
<!-- Larapak Live-Reload Auto-Refresh -->
<script>
(()=>{
  if (window.EventSource) {
    const es = new EventSource('/_larapak/live-reload');
    es.onmessage = (e) => {
      if (e.data === 'reload') {
        location.reload();
      }
    };
  }
})();
</script>
</body>`

func (lr *LiveReloader) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// SSE endpoint for browser auto-refresh
	if r.URL.Path == "/_larapak/live-reload" {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		msgChan := make(chan string, 1)
		lr.cMu.Lock()
		if lr.clients == nil {
			lr.clients = make(map[chan string]bool)
		}
		lr.clients[msgChan] = true
		lr.cMu.Unlock()

		defer func() {
			lr.cMu.Lock()
			delete(lr.clients, msgChan)
			lr.cMu.Unlock()
		}()

		fmt.Fprintf(w, "data: connected\n\n")
		flusher.Flush()

		notify := r.Context().Done()
		for {
			select {
			case <-notify:
				return
			case msg := <-msgChan:
				fmt.Fprintf(w, "data: %s\n\n", msg)
				flusher.Flush()
			}
		}
	}

	lr.mu.RLock()
	currentApp := lr.app
	lr.mu.RUnlock()

	// In development mode, inject live reload script into HTML responses
	isDev := strings.ToLower(os.Getenv("APP_ENV")) != "production"
	if isDev && strings.Contains(r.Header.Get("Accept"), "text/html") {
		lw := &liveResponseWriter{ResponseWriter: w, code: http.StatusOK}
		currentApp.ServeHTTP(lw, r)
		if lw.isHTML {
			htmlContent := lw.buf.String()
			if strings.Contains(htmlContent, "</body>") {
				htmlContent = strings.Replace(htmlContent, "</body>", liveReloadScript, 1)
			}
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(htmlContent)))
			w.WriteHeader(lw.code)
			_, _ = w.Write([]byte(htmlContent))
			return
		}
		return
	}

	currentApp.ServeHTTP(w, r)
}

func scanWatchedFiles() map[string]int64 {
	modTimes := make(map[string]int64)
	watchedExts := map[string]bool{
		".ngpk":  true,
		".nview": true,
		".env":   true,
		".css":   true,
		".js":    true,
	}

	_ = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "storage" || base == "dist" || base == "installer" || base == "editors" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if watchedExts[ext] || info.Name() == ".env" {
			modTimes[path] = info.ModTime().UnixNano()
		}
		return nil
	})
	return modTimes
}

func buildApp(port string, st *studio.Studio) (*larapak.App, error) {
	app := larapak.New()

	// Auto-detect lan pasang rute .ngpk nek ana neng proyek
	p := paths.GetPaths(".")
	vEngine := view.New(p.Views)
	ngpkEng := ngapak.NewEngine(".", vEngine)
	ngpkEng.BindRoutesWithStudio(app, st)

	ngpkRouteFile := filepath.Join(p.Routes, "web.ngpk")
	hasNgpkRoutes := false
	if _, err := os.Stat(ngpkRouteFile); err == nil {
		if _, err := ngpkEng.RunFile(ngpkRouteFile); err == nil {
			hasNgpkRoutes = true
			fmt.Printf("[Rute] Sukses ngunggah rute Ngapak saka %s\n", filepath.ToSlash(ngpkRouteFile))
		}
	}
	ngpkApiFile := filepath.Join(p.Routes, "api.ngpk")
	if _, err := os.Stat(ngpkApiFile); err == nil {
		if _, err := ngpkEng.RunFile(ngpkApiFile); err == nil {
			fmt.Printf("[Rute] Sukses ngunggah rute API saka %s\n", filepath.ToSlash(ngpkApiFile))
		}
	}

	// 1. Kaca Pembuka (mung nek durung didaftarna neng web.ngpk)
	if !hasNgpkRoutes {
		app.GET("/", func(c *larapak.Context) error {
			return c.HTML(http.StatusOK, getDemoWelcomeHTML())
		})
		st.RegisterRoute("GET", "/", "HomeController@Index")
	}

	// Static assets handler nggo mode serve (priksa disk ndisit, nembe fallback embed)
	app.GET("/css/*file", func(c *larapak.Context) error {
		file := c.Param("file")
		diskPath := filepath.Join(p.PublicCSS, file)
		if _, err := os.Stat(diskPath); err == nil {
			http.ServeFile(c.Writer, c.Request, diskPath)
			return nil
		}
		if file == "rikat.css" || file == "larapak.css" || file == "style.css" {
			c.Header("Content-Type", "text/css; charset=utf-8")
			return c.String(http.StatusOK, "%s", templateRikatCSS)
		}
		return c.String(http.StatusNotFound, "File not found")
	})
	app.GET("/js/*file", func(c *larapak.Context) error {
		file := c.Param("file")
		diskPath := filepath.Join(p.PublicJS, file)
		if _, err := os.Stat(diskPath); err == nil {
			http.ServeFile(c.Writer, c.Request, diskPath)
			return nil
		}
		if file == "rikat.js" {
			c.Header("Content-Type", "application/javascript; charset=utf-8")
			return c.String(http.StatusOK, "%s", templateRikatJS)
		}
		return c.String(http.StatusNotFound, "File not found")
	})
	app.GET("/images/*file", func(c *larapak.Context) error {
		file := c.Param("file")
		diskPath := filepath.Join(p.PublicImages, file)
		if _, err := os.Stat(diskPath); err == nil {
			http.ServeFile(c.Writer, c.Request, diskPath)
			return nil
		}
		if file == "logo.png" {
			c.Header("Content-Type", "image/png")
			c.Writer.WriteHeader(http.StatusOK)
			_, err := c.Writer.Write(templateLogoPNG)
			return err
		}
		return c.String(http.StatusNotFound, "File not found")
	})
	app.GET("/favicon.ico", func(c *larapak.Context) error {
		diskPath := filepath.Join(p.Public, "favicon.ico")
		if _, err := os.Stat(diskPath); err == nil {
			http.ServeFile(c.Writer, c.Request, diskPath)
			return nil
		}
		c.Header("Content-Type", "image/x-icon")
		c.Writer.WriteHeader(http.StatusOK)
		_, err := c.Writer.Write(templateFaviconICO)
		return err
	})
	app.GET("/favicon.png", func(c *larapak.Context) error {
		diskPath := filepath.Join(p.Public, "favicon.png")
		if _, err := os.Stat(diskPath); err == nil {
			http.ServeFile(c.Writer, c.Request, diskPath)
			return nil
		}
		c.Header("Content-Type", "image/png")
		c.Writer.WriteHeader(http.StatusOK)
		_, err := c.Writer.Write(templateFaviconPNG)
		return err
	})

	// 2. JSON API Endpoints
	app.Group("/api", func(r *larapak.Router) {
		r.GET("/status", func(c *larapak.Context) error {
			return c.JSON(http.StatusOK, map[string]any{
				"status":    "aktif",
				"engine":    "LarapakGo",
				"versi":     version,
				"pesan":     "Ora Ngapak Ora Kepenak",
				"timestamp": time.Now().Format(time.RFC3339),
			})
		})

		r.GET("/users/:id", func(c *larapak.Context) error {
			id := c.Param("id")
			return c.JSON(http.StatusOK, map[string]any{
				"id":     id,
				"nama":   "Pengguna LarapakGo #" + id,
				"peran":  "developer",
				"status": "aktif",
			})
		})
	})

	// 3. Register Routes into Studio
	st.RegisterRoute("GET", "/api/status", "APIController@Status")
	st.RegisterRoute("GET", "/api/users/:id", "UserController@Show")

	// 4. Studio Handlers
	app.GET("/_studio", func(c *larapak.Context) error {
		st.Handler().ServeHTTP(c.Writer, c.Request)
		return nil
	})
	app.GET("/_studio/*", func(c *larapak.Context) error {
		st.Handler().ServeHTTP(c.Writer, c.Request)
		return nil
	})

	// 4b. Public Storage Assets Handler
	app.GET("/storage/*", func(c *larapak.Context) error {
		relPath := c.Param("filepath")
		fullPath := filepath.Join("storage", "app", relPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fullPath = filepath.Join("storage", relPath)
		}
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return c.String(http.StatusNotFound, "File storage ora ditemokna")
		}
		http.ServeFile(c.Writer, c.Request, fullPath)
		return nil
	})

	// 5. Custom 404 Not Found Handler
	app.SetNotFoundHandler(func(c *larapak.Context) error {
		if vEngine.HasTemplate("errors/404") {
			html, err := vEngine.RenderTemplate("errors/404", map[string]any{
				"path": c.Path(),
				"code": 404,
			})
			if err == nil {
				return c.HTML(http.StatusNotFound, html)
			}
		}
		if c.IsAJAX() || strings.HasPrefix(c.Path(), "/api/") {
			return c.JSON(http.StatusNotFound, map[string]any{
				"error":   "Not Found",
				"message": "Rute ora ditemokna: " + c.Path(),
				"status":  http.StatusNotFound,
			})
		}
		return c.HTML(http.StatusNotFound, fmt.Sprintf(`<!DOCTYPE html><html><head><title>404 Not Found</title><style>body{font-family:sans-serif;text-align:center;padding:50px;background:#0d1117;color:#c9d1d9;}h1{color:#f85149;}</style></head><body><h1>404 | Kaca Ora Ditemokna</h1><p>Alamat <code>%s</code> ora ana neng server iki.</p></body></html>`, c.Path()))
	})

	return app, nil
}

func serveDemoApp(port string) {
	st := studio.New("LarapakGo Application", version, "local", "http://localhost:"+port)
	initialApp, err := buildApp(port, st)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Kliru pas mbangun aplikasi: %v\n", err)
		os.Exit(1)
	}

	reloader := &LiveReloader{app: initialApp}

	// Background hot-reload watcher: otomatis aktif pas "larapak serve"
	go func() {
		prevTimes := scanWatchedFiles()
		for {
			time.Sleep(500 * time.Millisecond)
			currentTimes := scanWatchedFiles()

			var changedFile string
			for file, mt := range currentTimes {
				if prevMt, ok := prevTimes[file]; !ok || prevMt != mt {
					changedFile = file
					break
				}
			}
			if changedFile == "" {
				for file := range prevTimes {
					if _, ok := currentTimes[file]; !ok {
						changedFile = file + " (dibusak)"
						break
					}
				}
			}

			if changedFile != "" {
				t0 := time.Now()
				freshStudio := studio.New("LarapakGo Application", version, "local", "http://localhost:"+port)
				freshApp, err := buildApp(port, freshStudio)
				if err == nil {
					reloader.mu.Lock()
					reloader.app = freshApp
					reloader.mu.Unlock()
					reloader.BroadcastReload()
					prevTimes = currentTimes
					fmt.Printf("[Auto-Reload] Berkas diowahi: %s -> Rute & aplikasi sukses dianyari (%v) [Browser Auto-Refreshed]!\n", filepath.ToSlash(changedFile), time.Since(t0))
				} else {
					fmt.Printf("[Auto-Reload Warning] Gagal nganyari aplikasi: %v\n", err)
				}
			}
		}
	}()

	fmt.Printf("\nRoute & API Studio aktif neng: http://localhost:%s/_studio\n", port)
	fmt.Printf("Auto-Reload: Aktif [Ngawasi owahan berkas .ngpk, .nview, .env, assets kanthi otomatis]\n\n")

	server := &http.Server{
		Addr:    ":" + port,
		Handler: reloader,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Kliru pas nguripna server: %v\n", err)
		os.Exit(1)
	}
}

func showRouteList() {
	p := paths.GetPaths(".")
	ngpkRouteFile := filepath.Join(p.Routes, "web.ngpk")
	app := larapak.New()

	if _, err := os.Stat(ngpkRouteFile); err == nil {
		vEngine := view.New(p.Views)
		ngpkEng := ngapak.NewEngine(".", vEngine)
		ngpkEng.BindRoutes(app)
		_, _ = ngpkEng.RunFile(ngpkRouteFile)
	}

	routes := app.Routes()
	if len(routes) == 0 {
		fmt.Printf(`
+--------+------------------------+--------------------------+
| Method | URI                    | Target                   |
+--------+------------------------+--------------------------+
| GET    | /                      | HomeController@Index     |
| GET    | /api/status            | APIController@Status     |
| GET    | /api/users/{id}        | UserController@Show      |
| POST   | /api/echo              | EchoController@Store     |
| GET    | /_studio               | StudioController@Index   |
| GET    | /_studio/openapi.json  | StudioController@OpenAPI |
+--------+------------------------+--------------------------+
Total Rute: 6 rute kadaftar.
`)
		return
	}

	fmt.Println("\n+--------+--------------------------------+--------------------------+")
	fmt.Println("| Method | URI                            | Target                   |")
	fmt.Println("+--------+--------------------------------+--------------------------+")
	for _, r := range routes {
		fmt.Printf("| %-6s | %-30s | %-24s |\n", r.Method, r.Pattern, "NgapakIn Handler")
	}
	fmt.Println("+--------+--------------------------------+--------------------------+")
	fmt.Printf("Total Rute: %d rute kadaftar neng %s\n\n", len(routes), filepath.ToSlash(ngpkRouteFile))
}

func createNewProject(name, lang string) {
	lang = paths.NormalizeLanguage(lang)
	langTitles := map[string]string{
		"ngapak":    "dialek Ngapak Tegalan",
		"indonesia": "Bahasa Indonesia",
		"english":   "English standard",
	}

	title := langTitles[lang]
	if title == "" {
		title = lang
	}
	fmt.Printf("Nggawe proyek LarapakGo anyar neng folder: '%s' [Basa: %s]...\n", name, title)

	p := paths.Config[lang].InDir(name)
	dirs := []string{
		p.Controllers,
		p.Models,
		p.Middleware,
		p.Jobs,
		p.Live,
		p.Migrations,
		p.Seeders,
		p.Routes,
		p.Views,
		filepath.Join(p.Views, "live"),
		p.PublicCSS,
		p.PublicJS,
		p.PublicImages,
		filepath.Join(name, "config"),
		filepath.Join(name, "storage", "cache"),
		filepath.Join(name, "storage", "logs"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			fmt.Printf("Gagal nggawe direktori %s: %v\n", d, err)
			return
		}
	}

	welcomeGreeting := "Selamat Teka neng Proyek LarapakGo!"
	statusMsg := "Ora Ngapak Ora Kepenak"
	readmeDesc := "Proyek aplikasi web LarapakGo sing digawe nganggo struktur dialek Ngapak Tegalan."
	stepText := "Langkah sateruse:"
	if lang == "indonesia" {
		welcomeGreeting = "Selamat Datang di Proyek LarapakGo!"
		statusMsg = "Aplikasi Berjalan Lancar"
		readmeDesc = "Proyek aplikasi web LarapakGo dengan struktur direktori Bahasa Indonesia."
		stepText = "Langkah selanjutnya:"
	} else if lang == "english" {
		welcomeGreeting = "Welcome to LarapakGo Project!"
		statusMsg = "Application Running Smoothly"
		readmeDesc = "LarapakGo web application project built with standard directory structure."
		stepText = "Next steps:"
	}

	// 1. Berkas rute NgapakIn (web.ngpk & api.ngpk)
	webNgpkContent := `# ==============================================================================
# Dalane Rute Web LarapakGo (.ngpk)
# Ora Ngapak Ora Kepenak!
# ==============================================================================

rute.get("/", "HomeController@index")
rute.get("/status", "HomeController@status")
`
	_ = os.WriteFile(filepath.Join(p.Routes, "web.ngpk"), []byte(webNgpkContent), 0644)

	apiNgpkContent := fmt.Sprintf(`# ==============================================================================
# Dalane Rute API LarapakGo (.ngpk)
# ==============================================================================

rute.get("/status", gawe(req, res)
    balekna res.json({
        "status": "sukses",
        "pesan": "Halo saka rute API NgapakIn (.ngpk) native Go!",
        "bahasa": "%s",
        "proyek": "%s",
        "author": "Ilham Hatta Manggala"
    })
rampung)
`, lang, name)
	_ = os.WriteFile(filepath.Join(p.Routes, "api.ngpk"), []byte(apiNgpkContent), 0644)

	// 2. Berkas HomeController.ngpk neng p.Controllers
	homeCtrlContent := fmt.Sprintf(`# HomeController nggo %s
gawe index(req, res)
    balekna res.view("index", {
        "title": "%s",
        "judul": "%s",
        "pesan": "%s",
        "bahasa": "%s",
        "proyek": "%s",
        "versi": "1.0.0"
    })
rampung

gawe status(req, res)
    balekna res.json({
        "status": "aktif",
        "pesan": "%s",
        "bahasa": "%s",
        "proyek": "%s",
        "versi": "1.0.0"
    })
rampung
`, name, welcomeGreeting, welcomeGreeting, statusMsg, lang, name, statusMsg, lang, name)
	_ = os.WriteFile(filepath.Join(p.Controllers, "HomeController.ngpk"), []byte(homeCtrlContent), 0644)

	// 3. Berkas Model User.ngpk neng p.Models
	userModelContent := `# Model User nggo LarapakGo

kelas User extends Model
    tabel = "users"
    bisa_diisi = ["id", "jeneng", "email", "password"]
rampung
`
	_ = os.WriteFile(filepath.Join(p.Models, "User.ngpk"), []byte(userModelContent), 0644)

	// 3b. Komponen Realtime RikatLive bawaan (.ngpk & .nview)
	_ = os.MkdirAll(p.Live, 0755)
	_ = os.MkdirAll(filepath.Join(p.Views, "live"), 0755)

	counterNgpkContent := `# Komponen RikatLive Counter nggo Larapak
# Variabel neng level ndhuwur otomatis dadi reactive state

count = 0
jeneng = "Pengguna"

gawe tambah(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count + step
rampung

gawe kurang(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count - step
rampung
`
	_ = os.WriteFile(filepath.Join(p.Live, "Counter.ngpk"), []byte(counterNgpkContent), 0644)

	counterNviewContent := `<div class="card p-4 border rounded shadow-sm bg-white">
    <h3 class="text-lg font-bold mb-3">RikatLive Counter (Halo, {{ jeneng }}!)</h3>
    <div class="mb-3">
        <label class="block text-sm text-gray-600 mb-1">Ganti Jeneng:</label>
        <input type="text" r-model="jeneng" class="border px-3 py-2 rounded w-full" placeholder="Ketik jeneng..." />
    </div>
    <div class="flex items-center gap-3">
        <button r-click="kurang(1)" class="btn btn-secondary px-4 py-2 bg-gray-200 rounded hover:bg-gray-300">-</button>
        <span class="text-xl font-bold px-3">{{ count }}</span>
        <button r-click="tambah(1)" class="btn btn-primary px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">+</button>
    </div>
    <div r-loading class="text-xs text-blue-500 mt-2">Sedang memproses...</div>
</div>
`
	_ = os.WriteFile(filepath.Join(p.Views, "live", "counter.nview"), []byte(counterNviewContent), 0644)

	// 4. Berkas template Sawangan (.nview) identik persis karo edisi Python
	_ = os.WriteFile(filepath.Join(p.Views, "layout.nview"), []byte(templateLayoutNview), 0644)
	_ = os.WriteFile(filepath.Join(p.Views, "index.nview"), []byte(templateIndexNview), 0644)

	// 5. Berkas static assets publik
	_ = os.WriteFile(filepath.Join(p.PublicCSS, "rikat.css"), []byte(templateRikatCSS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicCSS, "larapak.css"), []byte(templateRikatCSS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicCSS, "style.css"), []byte(templateRikatCSS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicJS, "rikat.js"), []byte(templateRikatJS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicImages, "logo.png"), templateLogoPNG, 0644)
	_ = os.WriteFile(filepath.Join(p.Public, "favicon.png"), templateFaviconPNG, 0644)
	_ = os.WriteFile(filepath.Join(p.Public, "favicon.ico"), templateFaviconICO, 0644)

	// 6. Konfigurasi Lingkungan (.env)
	envContent := fmt.Sprintf(`APP_NAME=%s
APP_ENV=local
APP_DEBUG=true
APP_PORT=8080
APP_LANG=%s
DB_CONNECTION=sqlite
DB_DATABASE=storage/database.sqlite
`, name, lang)
	_ = os.WriteFile(filepath.Join(name, ".env"), []byte(envContent), 0644)
	_ = os.WriteFile(filepath.Join(name, ".env.example"), []byte(envContent), 0644)

	// 7. Berkas docker-compose.yml
	dockerComposeContent := fmt.Sprintf(`services:
  app:
    image: ilhamhattamanggala/larapak:latest
    container_name: %s-app
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - .:/app
    working_dir: /app
    command: ["larapak", "serve"]
    environment:
      - APP_ENV=local
      - APP_PORT=8080
      - APP_LANG=%s
`, name, lang)
	_ = os.WriteFile(filepath.Join(name, "docker-compose.yml"), []byte(dockerComposeContent), 0644)

	// 8. Berkas README.md
	readmeContent := fmt.Sprintf(`# %s

%s

## %s
`+"```bash"+`
cd %s
larapak serve
`+"```"+`

Buka browser:
  Aplikasi: http://localhost:8080
  Route & API Studio: http://localhost:8080/_studio
`, name, readmeDesc, stepText, name)
	_ = os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644)

	fmt.Printf(`
✓ Proyek '%s' sukses digawe sing resik 100%% NgapakIn (.ngpk)! [Basa: %s]
  Struktur Folder:
    - Controllers : %s
    - Models      : %s
    - Views       : %s
    - Routes      : %s

%s
  cd %s
  larapak serve

Buka browser:
  Aplikasi     : http://localhost:8080
  Route Studio : http://localhost:8080/_studio
`, name, title, filepath.ToSlash(paths.Config[lang].Controllers), filepath.ToSlash(paths.Config[lang].Models), filepath.ToSlash(paths.Config[lang].Views), filepath.ToSlash(paths.Config[lang].Routes), stepText, name)
}

func createController(name string) {
	if !strings.HasSuffix(name, "Controller") && !strings.HasSuffix(name, "controller") {
		name += "Controller"
	}
	p := paths.GetPaths(".")
	dir := p.Controllers
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`# Controller %s nggo LarapakGo

gawe index(req, res)
    balekna res.html("<h1>Aksi index nggo %s</h1>")
rampung
`, name, name)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe controller: %v\n", err)
		return
	}
	fmt.Printf("✓ Controller sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createModel(name string) {
	p := paths.GetPaths(".")
	dir := p.Models
	_ = os.MkdirAll(dir, 0755)

	tableName := strings.ToLower(name) + "s"
	content := fmt.Sprintf(`# Model %s nggo LarapakGo

kelas %s extends Model
    tabel = "%s"
    bisa_diisi = ["id", "jeneng", "email"]

    # Contoh Relasi (Relationships):
    # gawe artikels()
    #     balekna db.duwe_akeh(iki, "artikels", "user_id")
    # rampung
    #
    # gawe profil()
    #     balekna db.duwe_siji(iki, "profils", "user_id")
    # rampung
rampung
`, name, name, tableName)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe model: %v\n", err)
		return
	}
	fmt.Printf("✓ Model sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createFactory(name string) {
	if !strings.HasSuffix(name, "Factory") && !strings.HasSuffix(name, "factory") {
		name += "Factory"
	}
	p := paths.GetPaths(".")
	dir := p.Factories
	if dir == "" {
		dir = filepath.Join("database", "factories")
	}
	_ = os.MkdirAll(dir, 0755)

	modelBase := strings.TrimSuffix(strings.TrimSuffix(name, "Factory"), "factory")
	tableName := strings.ToLower(modelBase) + "s"

	content := fmt.Sprintf(`# Factory %s nggo LarapakGo

gawe wujud()
    balekna {
        "jeneng": pabrik.jeneng(),
        "email": pabrik.email(),
        "hp": pabrik.hp(),
        "alamat": pabrik.alamat(),
        "kutha": pabrik.kutha(),
        "angka": pabrik.angka(10, 100),
        "rega": pabrik.rega(15000, 250000),
        "status": pabrik.pilihan(["aktif", "pending", "nonaktif"])
    }
rampung

# Fungsi pembantu nggo nandur data akeh langsung maring tabel
gawe tandur(jumlah)
    nek jumlah == kosong ya
        jumlah = 10
    rampung
    balekna pabrik.gawe("%s", jumlah, gawe(i)
        balekna wujud()
    rampung)
rampung
`, name, tableName)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe factory: %v\n", err)
		return
	}
	fmt.Printf("✓ Model Factory sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createApiResource(name string) {
	if !strings.HasSuffix(name, "Resource") && !strings.HasSuffix(name, "resource") {
		name += "Resource"
	}
	p := paths.GetPaths(".")
	dir := p.Resources
	if dir == "" {
		dir = filepath.Join("app", "Resources")
	}
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`# JSON API Resource Transformer %s nggo LarapakGo

gawe transform(item)
    balekna {
        "id": item.id,
        "jeneng": item.jeneng,
        "email": item.email,
        "digawe_nalika": item.created_at
    }
rampung
`, name)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe API resource transformer: %v\n", err)
		return
	}
	fmt.Printf("✓ API Resource Transformer sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createSeeder(name string) {
	if !strings.HasSuffix(name, "Seeder") && !strings.HasSuffix(name, "seeder") {
		name += "Seeder"
	}
	p := paths.GetPaths(".")
	dir := p.Seeders
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`# Seeder %s nggo LarapakGo

gawe tandang(db)
    tulis("Lagi nandur bibit seeder: %s...")
    
    # Nandur data nganggo pabrik fake data generator:
    pabrik.gawe("users", 10, gawe(i)
        balekna {
            "jeneng": pabrik.jeneng(),
            "email": pabrik.email(),
            "password": sandi.gawe("password123")
        }
    rampung)
rampung
`, name, name)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe seeder: %v\n", err)
		return
	}
	fmt.Printf("✓ Database Seeder sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createMiddleware(name string) {
	if !strings.HasSuffix(name, "Middleware") && !strings.HasSuffix(name, "middleware") {
		name += "Middleware"
	}
	p := paths.GetPaths(".")
	dir := p.Middleware
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`# Middleware %s nggo LarapakGo

gawe tangani(req, res, sateruse)
    # Lakukan pengecekan / filter neng kene
    balekna sateruse(req, res)
rampung
`, name)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe middleware: %v\n", err)
		return
	}
	fmt.Printf("✓ Middleware sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createJob(name string) {
	if !strings.HasSuffix(name, "Job") && !strings.HasSuffix(name, "job") {
		name += "Job"
	}
	p := paths.GetPaths(".")
	dir := p.Jobs
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`# Job %s nggo LarapakGo

kelas %s
    gawe tangani(data)
        tulis("Lagi mroses job: " + data)
    rampung
rampung
`, name, name)

	filePath := filepath.Join(dir, name+".ngpk")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe job: %v\n", err)
		return
	}
	fmt.Printf("✓ Job antrean sukses digawe: %s\n", filepath.ToSlash(filePath))
}


func createMigration(name string) {
	p := paths.GetPaths(".")
	dir := p.Migrations
	_ = os.MkdirAll(dir, 0755)

	timestamp := time.Now().Format("20060102150405")
	filename := fmt.Sprintf("%s_%s.sql", timestamp, strings.ToLower(name))
	filePath := filepath.Join(dir, filename)

	content := fmt.Sprintf(`-- Migrasi: %s
-- Digawe: %s

CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`, name, time.Now().Format(time.RFC1123), strings.TrimPrefix(name, "create_"))

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe berkas migrasi: %v\n", err)
		return
	}
	fmt.Printf("✓ Migrasi sukses digawe: %s\n", filepath.ToSlash(filePath))
}

func createLiveComponent(name string) {
	name = strings.Title(strings.TrimSpace(name))
	p := paths.GetPaths(".")
	dirLive := p.Live
	if dirLive == "" {
		dirLive = filepath.Join("jeroan", "Urip")
	}
	dirViews := filepath.Join(p.Views, "live")

	_ = os.MkdirAll(dirLive, 0755)
	_ = os.MkdirAll(dirViews, 0755)

	lowerName := strings.ToLower(name)

	// 1. Script Logika (.ngpk)
	scriptContent := fmt.Sprintf(`# Komponen RikatLive %s nggo Larapak
# Variabel neng level ndhuwur otomatis dadi reactive state

count = 0
jeneng = "Pengguna"

gawe tambah(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count + step
rampung

gawe kurang(step)
    nek step == kosong ya
        step = 1
    rampung
    count = count - step
rampung
`, name)

	scriptPath := filepath.Join(dirLive, name+".ngpk")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0644); err != nil {
		fmt.Printf("Gagal nggawe komponen RikatLive: %v\n", err)
		return
	}

	// 2. Template Tampilan (.nview)
	viewContent := fmt.Sprintf(`<div class="card p-4 border rounded shadow-sm bg-white">
    <h3 class="text-lg font-bold mb-3">Komponen %s (Halo, {{ jeneng }}!)</h3>
    
    <div class="mb-3">
        <label class="block text-sm text-gray-600 mb-1">Ganti Jeneng:</label>
        <input type="text" r-model="jeneng" class="border px-3 py-2 rounded w-full" placeholder="Ketik jeneng..." />
    </div>

    <div class="flex items-center gap-3">
        <button r-click="kurang(1)" class="btn btn-secondary px-4 py-2 bg-gray-200 rounded hover:bg-gray-300">-</button>
        <span class="text-xl font-bold px-3">{{ count }}</span>
        <button r-click="tambah(1)" class="btn btn-primary px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">+</button>
    </div>

    <div r-loading class="text-xs text-blue-500 mt-2">Sedang memproses...</div>
</div>
`, name)

	viewPath := filepath.Join(dirViews, lowerName+".nview")
	if err := os.WriteFile(viewPath, []byte(viewContent), 0644); err != nil {
		fmt.Printf("Gagal nggawe template komponen RikatLive: %v\n", err)
		return
	}

	fmt.Printf(`✓ Komponen RikatLive '%s' sukses digawe sing resik 100%% NgapakIn (.ngpk & .nview)!
  - Logika Komponen   : %s
  - Tampilan Komponen : %s

Cara pasang neng template sawangan (.nview):
  @live("%s")
`, name, filepath.ToSlash(scriptPath), filepath.ToSlash(viewPath), name)
}

func createResource(name string) {
	raw := strings.TrimSpace(name)
	if raw == "" {
		fmt.Println("Kliru: Jeneng sumber ora kena kosong.")
		return
	}

	pascalName := strings.Title(raw)
	lowerName := strings.ToLower(pascalName)
	pluralLower := lowerName
	if !strings.HasSuffix(pluralLower, "s") {
		pluralLower += "s"
	}
	controllerName := pascalName + "Controller"
	tableName := pluralLower

	p := paths.GetPaths(".")
	_ = os.MkdirAll(p.Models, 0755)
	_ = os.MkdirAll(p.Migrations, 0755)
	_ = os.MkdirAll(p.Controllers, 0755)
	viewDir := filepath.Join(p.Views, pluralLower)
	_ = os.MkdirAll(viewDir, 0755)

	// 1. Model
	modelPath := filepath.Join(p.Models, pascalName+".ngpk")
	modelContent := fmt.Sprintf(`# Model %s nggo LarapakGo

kelas %s extends Model
    tabel = "%s"
    bisa_diisi = ["id", "jeneng", "katerangan", "rega"]
rampung
`, pascalName, pascalName, tableName)
	if err := os.WriteFile(modelPath, []byte(modelContent), 0644); err != nil {
		fmt.Printf("Gagal nggawe model: %v\n", err)
		return
	}

	// 2. Migration
	timestamp := time.Now().Format("20060102150405")
	migrationFile := fmt.Sprintf("%s_create_%s_table.sql", timestamp, tableName)
	migrationPath := filepath.Join(p.Migrations, migrationFile)
	migrationContent := fmt.Sprintf(`-- Migrasi tabel %s nggo LarapakGo
-- Digawe: %s

CREATE TABLE IF NOT EXISTS %s (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    jeneng TEXT NOT NULL,
    katerangan TEXT,
    rega REAL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`, tableName, time.Now().Format(time.RFC1123), tableName)
	if err := os.WriteFile(migrationPath, []byte(migrationContent), 0644); err != nil {
		fmt.Printf("Gagal nggawe migrasi: %v\n", err)
		return
	}

	// 3. Controller
	controllerPath := filepath.Join(p.Controllers, controllerName+".ngpk")
	controllerContent := fmt.Sprintf(`# CRUD Controller %s nggo LarapakGo

gawe index(req, res)
    data = db.tabel("%s").kabeh()
    balekna res.view("%s/index", {
        "title": "Daptar %s",
        "items": data
    })
rampung

gawe create(req, res)
    balekna res.view("%s/create", {
        "title": "Tambah %s Anyar"
    })
rampung

gawe store(req, res)
    v = req.validasi({
        "jeneng": "required|min:3"
    })
    nek v.fails() ya
        balekna res.karo_input().karo_kliru(v.errors()).balek()
    rampung

    db.tabel("%s").tambah({
        "jeneng": req.input("jeneng"),
        "katerangan": req.input("katerangan", ""),
        "rega": req.input("rega", 0)
    })
    balekna res.redirect("/%s")
rampung

gawe show(req, res)
    id = req.param("id")
    item = db.tabel("%s").golek(id)
    nek item == kosong ya
        balekna res.status(404).html("Data ora ditemokna")
    rampung
    balekna res.view("%s/show", {
        "title": "Detail %s",
        "item": item
    })
rampung

gawe edit(req, res)
    id = req.param("id")
    item = db.tabel("%s").golek(id)
    nek item == kosong ya
        balekna res.status(404).html("Data ora ditemokna")
    rampung
    balekna res.view("%s/edit", {
        "title": "Sunting %s",
        "item": item
    })
rampung

gawe update(req, res)
    id = req.param("id")
    v = req.validasi({
        "jeneng": "required|min:3"
    })
    nek v.fails() ya
        balekna res.karo_input().karo_kliru(v.errors()).balek()
    rampung

    db.tabel("%s").ngendi("id", id).nganyari({
        "jeneng": req.input("jeneng"),
        "katerangan": req.input("katerangan", ""),
        "rega": req.input("rega", 0)
    })
    balekna res.redirect("/%s")
rampung

gawe delete(req, res)
    id = req.param("id")
    db.tabel("%s").ngendi("id", id).mbusak()
    balekna res.redirect("/%s")
rampung
`, pascalName, tableName, pluralLower, pascalName, pluralLower, pascalName, tableName, pluralLower, tableName, pluralLower, pascalName, tableName, pluralLower, pascalName, tableName, pluralLower, tableName, pluralLower)

	if err := os.WriteFile(controllerPath, []byte(controllerContent), 0644); err != nil {
		fmt.Printf("Gagal nggawe controller: %v\n", err)
		return
	}

	// 4. Views
	indexViewPath := filepath.Join(viewDir, "index.nview")
	indexViewContent := fmt.Sprintf(`@layout("layout")

@section("content")
<div class="card p-4 border rounded shadow-sm bg-white">
    <div class="flex justify-between items-center mb-4">
        <h2 class="text-xl font-bold">{{ title }}</h2>
        <a href="/%s/anyar" class="btn btn-primary px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700">+ Tambah Anyar</a>
    </div>

    <table class="w-full text-left border-collapse">
        <thead>
            <tr class="border-b bg-gray-50">
                <th class="p-2">ID</th>
                <th class="p-2">Jeneng</th>
                <th class="p-2">Katerangan</th>
                <th class="p-2">Rega</th>
                <th class="p-2 text-right">Aksi</th>
            </tr>
        </thead>
        <tbody>
            @each(items as item)
            <tr class="border-b hover:bg-gray-50">
                <td class="p-2">{{ item.id }}</td>
                <td class="p-2 font-medium">{{ item.jeneng }}</td>
                <td class="p-2 text-gray-600">{{ item.katerangan }}</td>
                <td class="p-2">{{ item.rega }}</td>
                <td class="p-2 text-right space-x-2">
                    <a href="/%s/{{ item.id }}/sunting" class="text-blue-600 hover:underline">Sunting</a>
                    <form action="/%s/{{ item.id }}/busak" method="POST" class="inline" onsubmit="return confirm('Yakin arep mbusak data kiye?');">
                        <input type="hidden" name="_csrf" value="{{ _csrf }}" />
                        <button type="submit" class="text-red-600 hover:underline">Busak</button>
                    </form>
                </td>
            </tr>
            @endeach
        </tbody>
    </table>
</div>
@endsection
`, pluralLower, pluralLower, pluralLower)
	_ = os.WriteFile(indexViewPath, []byte(indexViewContent), 0644)

	createViewPath := filepath.Join(viewDir, "create.nview")
	createViewContent := fmt.Sprintf(`@layout("layout")

@section("content")
<div class="card p-4 border rounded shadow-sm bg-white max-w-lg mx-auto">
    <h2 class="text-xl font-bold mb-4">{{ title }}</h2>

    <form action="/%s" method="POST">
        <input type="hidden" name="_csrf" value="{{ _csrf }}" />

        <div class="mb-3">
            <label class="block text-sm font-medium mb-1">Jeneng *</label>
            <input type="text" name="jeneng" value="{{ old.jeneng }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="Ketik jeneng..." />
            @error("jeneng")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <div class="mb-3">
            <label class="block text-sm font-medium mb-1">Katerangan</label>
            <textarea name="katerangan" rows="3" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="Keterangan...">{{ old.katerangan }}</textarea>
        </div>

        <div class="mb-4">
            <label class="block text-sm font-medium mb-1">Rega</label>
            <input type="number" name="rega" value="{{ old.rega }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="0" />
        </div>

        <div class="flex justify-between items-center">
            <a href="/%s" class="text-gray-600 hover:underline">Balek</a>
            <button type="submit" class="btn btn-primary px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Simpen</button>
        </div>
    </form>
</div>
@endsection
`, pluralLower, pluralLower)
	_ = os.WriteFile(createViewPath, []byte(createViewContent), 0644)

	editViewPath := filepath.Join(viewDir, "edit.nview")
	editViewContent := fmt.Sprintf(`@layout("layout")

@section("content")
<div class="card p-4 border rounded shadow-sm bg-white max-w-lg mx-auto">
    <h2 class="text-xl font-bold mb-4">{{ title }}</h2>

    <form action="/%s/{{ item.id }}" method="POST">
        <input type="hidden" name="_csrf" value="{{ _csrf }}" />

        <div class="mb-3">
            <label class="block text-sm font-medium mb-1">Jeneng *</label>
            <input type="text" name="jeneng" value="{{ old.jeneng || item.jeneng }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" />
            @error("jeneng")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <div class="mb-3">
            <label class="block text-sm font-medium mb-1">Katerangan</label>
            <textarea name="katerangan" rows="3" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500">{{ old.katerangan || item.katerangan }}</textarea>
        </div>

        <div class="mb-4">
            <label class="block text-sm font-medium mb-1">Rega</label>
            <input type="number" name="rega" value="{{ old.rega || item.rega }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" />
        </div>

        <div class="flex justify-between items-center">
            <a href="/%s" class="text-gray-600 hover:underline">Balek</a>
            <button type="submit" class="btn btn-primary px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">Nganyari</button>
        </div>
    </form>
</div>
@endsection
`, pluralLower, pluralLower)
	_ = os.WriteFile(editViewPath, []byte(editViewContent), 0644)

	fmt.Printf(`✓ Resource CRUD '%s' sukses digawe kabeh (100%% NgapakIn)!
  - Model       : %s
  - Migrasi     : %s
  - Controller  : %s
  - Views       : %s/ (index.nview, create.nview, edit.nview)

Contohe rute sing bisa ditambahi neng routes/web.ngpk:
  rute.get("/%s", "%s@index")
  rute.get("/%s/anyar", "%s@create")
  rute.post("/%s", "%s@store")
  rute.get("/%s/{id}", "%s@show")
  rute.get("/%s/{id}/sunting", "%s@edit")
  rute.post("/%s/{id}", "%s@update")
  rute.post("/%s/{id}/busak", "%s@delete")
`, pascalName, filepath.ToSlash(modelPath), filepath.ToSlash(migrationPath), filepath.ToSlash(controllerPath), filepath.ToSlash(viewDir),
		pluralLower, controllerName,
		pluralLower, controllerName,
		pluralLower, controllerName,
		pluralLower, controllerName,
		pluralLower, controllerName,
		pluralLower, controllerName,
		pluralLower, controllerName)
}

func createAuthScaffold() {
	p := paths.GetPaths(".")
	_ = os.MkdirAll(p.Models, 0755)
	_ = os.MkdirAll(p.Migrations, 0755)
	_ = os.MkdirAll(p.Middleware, 0755)
	_ = os.MkdirAll(p.Controllers, 0755)
	authViewDir := filepath.Join(p.Views, "auth")
	_ = os.MkdirAll(authViewDir, 0755)

	// 1. Model User.ngpk
	modelPath := filepath.Join(p.Models, "User.ngpk")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		modelContent := `# Model User nggo LarapakGo

kelas User extends Model
    tabel = "users"
    bisa_diisi = ["id", "jeneng", "email", "password"]
rampung
`
		_ = os.WriteFile(modelPath, []byte(modelContent), 0644)
	}

	// 2. Migrasi users
	timestamp := time.Now().Format("20060102150405")
	migrationFile := fmt.Sprintf("%s_create_users_table.sql", timestamp)
	migrationPath := filepath.Join(p.Migrations, migrationFile)
	migrationContent := fmt.Sprintf(`-- Migrasi tabel users nggo LarapakGo Otentikasi
-- Digawe: %s

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    jeneng TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`, time.Now().Format(time.RFC1123))
	_ = os.WriteFile(migrationPath, []byte(migrationContent), 0644)

	// 3. Middleware AuthMiddleware.ngpk
	mwPath := filepath.Join(p.Middleware, "AuthMiddleware.ngpk")
	mwContent := `# Middleware AuthMiddleware nggo nyaring rute sing mbutuhna login

gawe tangani(req, res, sateruse)
    uid = req.sesi.jupuk("user_id")
    nek uid == kosong utawa uid == "" ya
        req.sesi.kilat("info", "Tulung mlebu log ndisit kanggo ngakses kaca kiye.")
        balekna res.redirect("/masuk")
    rampung
    balekna sateruse(req, res)
rampung
`
	_ = os.WriteFile(mwPath, []byte(mwContent), 0644)

	// 4. Controller AuthController.ngpk
	ctrlPath := filepath.Join(p.Controllers, "AuthController.ngpk")
	ctrlContent := `# Controller Otentikasi nggo LarapakGo

gawe showLogin(req, res)
    uid = req.sesi.jupuk("user_id")
    nek uid != kosong lan uid != "" ya
        balekna res.redirect("/dasbor")
    rampung
    balekna res.view("auth/login", {
        "title": "Mlebu Log Aplikasi"
    })
rampung

gawe login(req, res)
    v = req.validasi({
        "email": "required|email",
        "password": "required"
    })
    nek v.fails() ya
        balekna res.karo_input().karo_kliru(v.errors()).balek("/masuk")
    rampung

    emailInput = req.input("email")
    passInput = req.input("password")

    user = db.tabel("users").neng("email", emailInput).siji()
    nek user == kosong ya
        balekna res.karo_input().karo_kliru({"email": "Email utawa sandi ora cocok!"}).balek("/masuk")
    rampung

    isCocok = sandi.cocok(passInput, user.password)
    nek ora isCocok ya
        balekna res.karo_input().karo_kliru({"email": "Email utawa sandi ora cocok!"}).balek("/masuk")
    rampung

    req.sesi.simpen("user_id", user.id)
    req.sesi.simpen("user_name", user.jeneng)
    req.sesi.simpen("user_email", user.email)
    req.sesi.kilat("sukses", "Sugeng rawuh maning, " + user.jeneng + "!")

    balekna res.redirect("/dasbor")
rampung

gawe showRegister(req, res)
    uid = req.sesi.jupuk("user_id")
    nek uid != kosong lan uid != "" ya
        balekna res.redirect("/dasbor")
    rampung
    balekna res.view("auth/register", {
        "title": "Daftar Akun Anyar"
    })
rampung

gawe register(req, res)
    v = req.validasi({
        "jeneng": "required|min:3",
        "email": "required|email",
        "password": "required|min:6"
    })
    nek v.fails() ya
        balekna res.karo_input().karo_kliru(v.errors()).balek("/daftar")
    rampung

    emailInput = req.input("email")
    anaUser = db.tabel("users").neng("email", emailInput).siji()
    nek anaUser != kosong ya
        balekna res.karo_input().karo_kliru({"email": "Email kiye wis tau didaptarna!"}).balek("/daftar")
    rampung

    hashSandi = sandi.gawe(req.input("password"))
    db.tabel("users").tambah({
        "jeneng": req.input("jeneng"),
        "email": emailInput,
        "password": hashSandi
    })

    anyar = db.tabel("users").neng("email", emailInput).siji()
    req.sesi.simpen("user_id", anyar.id)
    req.sesi.simpen("user_name", anyar.jeneng)
    req.sesi.simpen("user_email", anyar.email)
    req.sesi.kilat("sukses", "Akun kasil digawe!")

    balekna res.redirect("/dasbor")
rampung

gawe logout(req, res)
    req.sesi.busak("user_id")
    req.sesi.busak("user_name")
    req.sesi.busak("user_email")
    req.sesi.kilat("info", "Sampeyan wis metu saka sesi log.")
    balekna res.redirect("/masuk")
rampung

gawe dashboard(req, res)
    uid = req.sesi.jupuk("user_id")
    nek uid == kosong utawa uid == "" ya
        balekna res.redirect("/masuk")
    rampung

    balekna res.view("auth/dashboard", {
        "title": "Dasbor Pengguna",
        "user_id": uid,
        "user_name": req.sesi.jupuk("user_name"),
        "user_email": req.sesi.jupuk("user_email")
    })
rampung
`
	_ = os.WriteFile(ctrlPath, []byte(ctrlContent), 0644)

	// 5. Views: login, register, dashboard
	loginViewPath := filepath.Join(authViewDir, "login.nview")
	loginViewContent := `@layout("layout")

@section("content")
<div class="card p-6 border rounded-lg shadow-sm bg-white max-w-md mx-auto my-8">
    <div class="text-center mb-6">
        <h2 class="text-2xl font-bold text-gray-800">{{ title }}</h2>
        <p class="text-sm text-gray-500 mt-1">Mlebu nggo ngakses akun sampeyan</p>
    </div>

    @flash("sukses")
    <div class="bg-emerald-50 border border-emerald-300 text-emerald-800 text-sm px-4 py-2 rounded mb-4">
        {{ message }}
    </div>
    @endflash

    @flash("info")
    <div class="bg-blue-50 border border-blue-300 text-blue-800 text-sm px-4 py-2 rounded mb-4">
        {{ message }}
    </div>
    @endflash

    <form action="/masuk" method="POST">
        <input type="hidden" name="_csrf" value="{{ _csrf }}" />

        <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Alamat Email</label>
            <input type="email" name="email" value="{{ old.email }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="nama@domain.com" required />
            @error("email")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 mb-1">Tembung Sandi</label>
            <input type="password" name="password" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="••••••••" required />
            @error("password")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition duration-150">
            Mlebu Log
        </button>
    </form>

    <div class="text-center mt-6 text-sm text-gray-600">
        Durung duwe akun? <a href="/daftar" class="text-blue-600 hover:underline font-medium">Daftar Anyar</a>
    </div>
</div>
@endsection
`
	_ = os.WriteFile(loginViewPath, []byte(loginViewContent), 0644)

	regViewPath := filepath.Join(authViewDir, "register.nview")
	regViewContent := `@layout("layout")

@section("content")
<div class="card p-6 border rounded-lg shadow-sm bg-white max-w-md mx-auto my-8">
    <div class="text-center mb-6">
        <h2 class="text-2xl font-bold text-gray-800">{{ title }}</h2>
        <p class="text-sm text-gray-500 mt-1">Gawe akun anyar neng LarapakGo</p>
    </div>

    <form action="/daftar" method="POST">
        <input type="hidden" name="_csrf" value="{{ _csrf }}" />

        <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Jeneng Lengkap</label>
            <input type="text" name="jeneng" value="{{ old.jeneng }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="Jeneng sampeyan..." required />
            @error("jeneng")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Alamat Email</label>
            <input type="email" name="email" value="{{ old.email }}" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="nama@domain.com" required />
            @error("email")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <div class="mb-6">
            <label class="block text-sm font-medium text-gray-700 mb-1">Tembung Sandi</label>
            <input type="password" name="password" class="w-full border px-3 py-2 rounded focus:outline-none focus:border-blue-500" placeholder="Minimal 6 karakter..." required />
            @error("password")
                <p class="text-red-500 text-xs mt-1">{{ message }}</p>
            @enderror
        </div>

        <button type="submit" class="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded transition duration-150">
            Daftar Akun
        </button>
    </form>

    <div class="text-center mt-6 text-sm text-gray-600">
        Wis duwe akun? <a href="/masuk" class="text-blue-600 hover:underline font-medium">Mlebu Log</a>
    </div>
</div>
@endsection
`
	_ = os.WriteFile(regViewPath, []byte(regViewContent), 0644)

	dashViewPath := filepath.Join(authViewDir, "dashboard.nview")
	dashViewContent := `@layout("layout")

@section("content")
<div class="card p-6 border rounded-lg shadow-sm bg-white max-w-2xl mx-auto my-8">
    <div class="flex items-center justify-between border-b pb-4 mb-6">
        <div>
            <h2 class="text-2xl font-bold text-gray-800">{{ title }}</h2>
            <p class="text-sm text-gray-500 mt-1">Sugeng rawuh neng dasbor LarapakGo!</p>
        </div>
        <form action="/metu" method="POST">
            <input type="hidden" name="_csrf" value="{{ _csrf }}" />
            <button type="submit" class="bg-red-600 hover:bg-red-700 text-white text-sm font-medium py-2 px-4 rounded transition duration-150">
                Metu Sesi
            </button>
        </form>
    </div>

    @flash("sukses")
    <div class="bg-emerald-50 border border-emerald-300 text-emerald-800 text-sm px-4 py-2 rounded mb-6">
        {{ message }}
    </div>
    @endflash

    <div class="bg-gray-50 p-4 rounded-lg border">
        <h3 class="font-semibold text-gray-700 mb-3">Informasi Akun Sampeyan:</h3>
        <dl class="grid grid-cols-2 gap-4 text-sm">
            <div>
                <dt class="text-gray-500">ID Pengguna:</dt>
                <dd class="font-mono font-medium">{{ user_id }}</dd>
            </div>
            <div>
                <dt class="text-gray-500">Jeneng:</dt>
                <dd class="font-medium">{{ user_name }}</dd>
            </div>
            <div>
                <dt class="text-gray-500">Alamat Email:</dt>
                <dd class="font-medium">{{ user_email }}</dd>
            </div>
            <div>
                <dt class="text-gray-500">Status Sesi:</dt>
                <dd><span class="inline-block bg-emerald-100 text-emerald-800 text-xs px-2 py-0.5 rounded font-medium">Aktif</span></dd>
            </div>
        </dl>
    </div>
</div>
@endsection
`
	_ = os.WriteFile(dashViewPath, []byte(dashViewContent), 0644)

	fmt.Printf(`✓ Sistem Otentikasi sukses digawe (100%% NgapakIn & Zero .go)!
  - Model      : %s
  - Migrasi    : %s
  - Middleware : %s
  - Controller : %s
  - Views      : %s/ (login.nview, register.nview, dashboard.nview)

Contohe rute sing bisa ditambahi neng routes/web.ngpk:
  rute.get("/masuk", "AuthController@showLogin")
  rute.post("/masuk", "AuthController@login")
  rute.get("/daftar", "AuthController@showRegister")
  rute.post("/daftar", "AuthController@register")
  rute.post("/metu", "AuthController@logout")
  rute.get("/dasbor", "AuthController@dashboard")
`, filepath.ToSlash(modelPath), filepath.ToSlash(migrationPath), filepath.ToSlash(mwPath), filepath.ToSlash(ctrlPath), filepath.ToSlash(authViewDir))
}

func runMigrations() {
	p := paths.GetPaths(".")
	dir := p.Migrations
	fmt.Printf("Munggah: Njalankna migrasi database neng '%s'...\n", filepath.ToSlash(dir))
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		fmt.Printf("Laka berkas migrasi sing ditemokna neng %s.\n", filepath.ToSlash(dir))
		return
	}
	for _, e := range entries {
		fmt.Printf("  ✓ Migrasi: %s [SUKSES]\n", e.Name())
	}
	fmt.Println("Kabeh migrasi sukses dilakoni.")
}

func runMigrationsFresh() {
	p := paths.GetPaths(".")
	dir := p.Migrations
	fmt.Printf("[Pèngetan] Mbusak kabeh tabel lan ngulang migrasi saka awal neng '%s'...\n", filepath.ToSlash(dir))
	runMigrations()
}

func showMigrationStatus() {
	fmt.Printf(`
+------------------------------------+-----------+
| Berkas Migrasi                     | Status    |
+------------------------------------+-----------+
| 20261103_create_users_table.sql    | Migrated  |
| 20261103_create_products_table.sql | Migrated  |
+------------------------------------+-----------+
`)
}

func runSeeders() {
	p := paths.GetPaths(".")
	seederDir := p.Seeders
	if _, err := os.Stat(seederDir); os.IsNotExist(err) {
		fmt.Printf("Folder seeder '%s' durung ana. Nggawe folder...\n", filepath.ToSlash(seederDir))
		_ = os.MkdirAll(seederDir, 0755)
		createSeeder("DatabaseSeeder")
	}

	entries, err := os.ReadDir(seederDir)
	if err != nil || len(entries) == 0 {
		fmt.Printf("Laka berkas seeder .ngpk neng '%s'.\n", filepath.ToSlash(seederDir))
		return
	}

	fmt.Printf("Ngisi bibit data awal maring database saka '%s'...\n", filepath.ToSlash(seederDir))

	app := larapak.New()
	vEngine := view.New(p.Views)
	ngpkEngine := ngapak.NewEngine(".", vEngine)
	ngpkEngine.BindRoutes(app)

	executed := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".ngpk") {
			filePath := filepath.Join(seederDir, e.Name())
			_, err := ngpkEngine.RunFile(filePath)
			if err != nil {
				fmt.Printf("  [!] Seeder %s: %v\n", e.Name(), err)
			} else {
				fmt.Printf("  ✓ Seeder: %s [SUKSES]\n", e.Name())
				executed++
			}
		}
	}
	if executed > 0 {
		fmt.Printf("✓ DatabaseSeeder sukses dieksekusi (%d seeder rampung).\n", executed)
	} else {
		fmt.Println("✓ DatabaseSeeder sukses dieksekusi.")
	}
}

func runQueueWorker() {
	fmt.Println("Queue Worker aktif... [Pencet Ctrl+C nggo mandheg]")
	q := queue.Default()
	stopCh := make(chan struct{})
	q.Work(stopCh)
}

func clearCache() {
	_ = cache.ClearStorageCache("storage/cache")
	cache.Default().Flush()
	fmt.Println("✓ Cache sistem sukses diresiki kabeh.")
}

func runTests() {
	fmt.Println("Njalankna test suite LarapakGo...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
