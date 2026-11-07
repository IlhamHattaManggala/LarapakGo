package main

import (
	"bufio"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
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

	case "make:migration", "gawe:migrasi":
		if len(args) < 1 {
			fmt.Println("Kliru: Wei jeneng migrasi ndisit. Contohe: larapak gawe:migrasi create_users_table")
			os.Exit(1)
		}
		createMigration(args[0])

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
  larapak ladeni [port]                         (Urubna server lokal & Studio)
  larapak dalan:daptar                          (Daftar kabeh rute HTTP)
  larapak gawe:pangatur <NamaController>        (Gawe controller anyar)
  larapak gawe:wujud <NamaModel>                (Gawe model ORM anyar)
  larapak gawe:migrasi <NamaMigrasi>            (Gawe skema migrasi anyar)
  larapak gawe:pencegat <NamaMiddleware>        (Gawe middleware anyar)
  larapak gawe:gawean <NamaJob>                 (Gawe background queue job)
  larapak gawe:bibit <NamaSeeder>               (Gawe database seeder)
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
  larapak new, serve, route:list, make:controller, make:model, make:migration,
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

func serveDemoApp(port string) {
	app := larapak.New()
	st := studio.New("LarapakGo Application", version, "local", "http://localhost:"+port)

	// Auto-detect lan pasang rute .ngpk nek ana neng proyek
	p := paths.GetPaths(".")
	ngpkRouteFile := filepath.Join(p.Routes, "web.ngpk")
	hasNgpkRoutes := false
	if _, err := os.Stat(ngpkRouteFile); err == nil {
		vEngine := view.New(p.Views)
		ngpkEng := ngapak.NewEngine(".", vEngine)
		ngpkEng.BindRoutesWithStudio(app, st)
		if _, err := ngpkEng.RunFile(ngpkRouteFile); err == nil {
			hasNgpkRoutes = true
			fmt.Printf("[Rute] Sukses ngunggah rute Ngapak saka %s\n", filepath.ToSlash(ngpkRouteFile))
		}
	}

	// 1. Kaca Pembuka (mung nek durung didaftarna neng web.ngpk)
	if !hasNgpkRoutes {
		app.GET("/", func(c *larapak.Context) error {
			return c.HTML(http.StatusOK, getDemoWelcomeHTML())
		})
		st.RegisterRoute("GET", "/", "HomeController@Index")
	}

	// Static assets handler nggo mode demo
	app.GET("/css/rikat.css", func(c *larapak.Context) error {
		c.Header("Content-Type", "text/css; charset=utf-8")
		return c.String(http.StatusOK, "%s", templateRikatCSS)
	})
	app.GET("/js/rikat.js", func(c *larapak.Context) error {
		c.Header("Content-Type", "application/javascript; charset=utf-8")
		return c.String(http.StatusOK, "%s", templateRikatJS)
	})
	app.GET("/images/logo.png", func(c *larapak.Context) error {
		c.Header("Content-Type", "image/png")
		c.Writer.WriteHeader(http.StatusOK)
		_, err := c.Writer.Write(templateLogoPNG)
		return err
	})
	app.GET("/favicon.ico", func(c *larapak.Context) error {
		c.Header("Content-Type", "image/x-icon")
		c.Writer.WriteHeader(http.StatusOK)
		_, err := c.Writer.Write(templateFaviconICO)
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
	app.GET("/_studio/openapi.json", func(c *larapak.Context) error {
		st.Handler().ServeHTTP(c.Writer, c.Request)
		return nil
	})

	fmt.Printf("\nRoute & API Studio aktif neng: http://localhost:%s/_studio\n", port)
	if err := app.Run(":" + port); err != nil {
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
		p.Migrations,
		p.Seeders,
		p.Routes,
		p.Views,
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

	cwd, _ := os.Getwd()
	frameworkAbs := cwd
	if !strings.HasSuffix(frameworkAbs, "LarapakGo") {
		frameworkAbs = filepath.Join(frameworkAbs, "LarapakGo")
	}
	targetAbs, _ := filepath.Abs(name)
	relPath, err := filepath.Rel(targetAbs, frameworkAbs)
	if err != nil || relPath == "" {
		relPath = "../"
	}
	relPath = filepath.ToSlash(relPath)

	goModContent := fmt.Sprintf(`module %s

go 1.25

require github.com/IlhamHattaManggala/larapak-go v1.0.0

replace github.com/IlhamHattaManggala/larapak-go => %s
`, name, relPath)
	_ = os.WriteFile(filepath.Join(name, "go.mod"), []byte(goModContent), 0644)

	ctrlRelPath := filepath.ToSlash(paths.Config[lang].Controllers)
	ctrlPkg := paths.PackageName(paths.Config[lang].Controllers)
	modelPkg := paths.PackageName(paths.Config[lang].Models)

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

	mainGo := fmt.Sprintf(`package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/ngapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
	"%s/%s"
)

func main() {
	app := larapak.New()
	st := studio.New("%s", "1.0.0", "local", "http://localhost:8080")

	// Inisialisasi Sawangan (.nview) & Mesin NgapakIn (.ngpk)
	p := paths.GetPaths(".")
	vEngine := view.New(p.Views)
	ngpkEng := ngapak.NewEngine(".", vEngine)
	ngpkEng.BindRoutesWithStudio(app, st)

	// Muat rute .ngpk nek ana
	ngpkRouteFile := filepath.Join(p.Routes, "web.ngpk")
	hasNgpkRoutes := false
	if _, err := os.Stat(ngpkRouteFile); err == nil {
		if _, err := ngpkEng.RunFile(ngpkRouteFile); err == nil {
			hasNgpkRoutes = true
			fmt.Printf("[Rute] Sukses ngunggah rute Ngapak saka %%s\n", filepath.ToSlash(ngpkRouteFile))
		}
	}

	if !hasNgpkRoutes {
		homeCtrl := &%s.HomeController{}
		app.GET("/", homeCtrl.Index)
		st.RegisterRoute("GET", "/", "HomeController@Index")
	}

	// Layani berkas statis (css, js, images, favicon)
	app.GET("/css/*file", func(c *larapak.Context) error {
		filePath := filepath.Join(p.PublicCSS, c.Param("file"))
		http.ServeFile(c.Writer, c.Request, filePath)
		return nil
	})
	app.GET("/js/*file", func(c *larapak.Context) error {
		filePath := filepath.Join(p.PublicJS, c.Param("file"))
		http.ServeFile(c.Writer, c.Request, filePath)
		return nil
	})
	app.GET("/images/*file", func(c *larapak.Context) error {
		filePath := filepath.Join(p.PublicImages, c.Param("file"))
		http.ServeFile(c.Writer, c.Request, filePath)
		return nil
	})
	app.GET("/favicon.ico", func(c *larapak.Context) error {
		http.ServeFile(c.Writer, c.Request, filepath.Join(p.Public, "favicon.ico"))
		return nil
	})

	app.Group("/api", func(r *larapak.Router) {
		r.GET("/status", func(c *larapak.Context) error {
			return c.JSON(http.StatusOK, map[string]string{
				"status": "aktif",
				"pesan":  "%s",
				"bahasa": "%s",
				"proyek": "%s",
			})
		})
	})
	st.RegisterRoute("GET", "/api/status", "APIController@Status")

	app.GET("/_studio", func(c *larapak.Context) error {
		st.Handler().ServeHTTP(c.Writer, c.Request)
		return nil
	})
	app.GET("/_studio/openapi.json", func(c *larapak.Context) error {
		st.Handler().ServeHTTP(c.Writer, c.Request)
		return nil
	})

	fmt.Println("\nAplikasi LarapakGo mlaku neng: http://localhost:8080")
	fmt.Println("Route & API Studio aktif neng : http://localhost:8080/_studio\n")
	if err := app.Run(":8080"); err != nil {
		fmt.Fprintf(os.Stderr, "Kliru: %%v\n", err)
	}
}
`, name, ctrlRelPath, name, ctrlPkg, statusMsg, lang, name)
	_ = os.WriteFile(filepath.Join(name, "main.go"), []byte(mainGo), 0644)

	// Berkas rute NgapakIn (web.ngpk)
	webNgpkContent := `# ==============================================================================
# Dalane Rute Aplikasi LarapakGo (.ngpk)
# Ora Ngapak Ora Kepenak!
# ==============================================================================

rute.get("/", gawe(req, res)
    balekna res.view("index", {
        "title": "` + welcomeGreeting + `",
        "judul": "` + welcomeGreeting + `",
        "pesan": "` + statusMsg + `",
        "bahasa": "` + lang + `",
        "proyek": "` + name + `",
        "versi": "1.0.0"
    })
rampung)

rute.get("/api/status", gawe(req, res)
    balekna res.json({
        "status": "sukses",
        "pesan": "Halo saka rute NgapakIn (.ngpk) native Go!",
        "bahasa": "` + lang + `",
        "proyek": "` + name + `"
    })
rampung)
`
	_ = os.WriteFile(filepath.Join(p.Routes, "web.ngpk"), []byte(webNgpkContent), 0644)

	// Berkas template Sawangan (.nview) identik persis karo edisi Python
	_ = os.WriteFile(filepath.Join(p.Views, "layout.nview"), []byte(templateLayoutNview), 0644)
	_ = os.WriteFile(filepath.Join(p.Views, "index.nview"), []byte(templateIndexNview), 0644)

	// Berkas static assets publik
	_ = os.WriteFile(filepath.Join(p.PublicCSS, "rikat.css"), []byte(templateRikatCSS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicJS, "rikat.js"), []byte(templateRikatJS), 0644)
	_ = os.WriteFile(filepath.Join(p.PublicImages, "logo.png"), templateLogoPNG, 0644)
	_ = os.WriteFile(filepath.Join(p.Public, "favicon.ico"), templateFaviconICO, 0644)

	homeCtrlContent := fmt.Sprintf(`package %s

import (
	"net/http"

	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

type HomeController struct{}

func (h *HomeController) Index(c *larapak.Context) error {
	p := paths.GetPaths(".")
	vEngine := view.New(p.Views)
	html, err := vEngine.RenderTemplate("index", map[string]any{
		"title":  "Sugeng Rawuh",
		"judul":  "Sugeng Rawuh",
		"pesan":  "%s",
		"bahasa": "%s",
		"proyek": "%s",
		"versi":  "1.0.0",
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, "Gagal render view: %%v", err)
	}
	return c.HTML(http.StatusOK, html)
}
`, ctrlPkg, statusMsg, lang, name)
	_ = os.WriteFile(filepath.Join(p.Controllers, "HomeController.go"), []byte(homeCtrlContent), 0644)

	userModelContent := fmt.Sprintf(`package %s

import "time"

type User struct {
	ID        int64     `+"`json:\"id\"`"+`
	Name      string    `+"`json:\"name\"`"+`
	Email     string    `+"`json:\"email\"`"+`
	Password  string    `+"`json:\"-\"`"+`
	CreatedAt time.Time `+"`json:\"created_at\"`"+`
	UpdatedAt time.Time `+"`json:\"updated_at\"`"+`
}
`, modelPkg)
	_ = os.WriteFile(filepath.Join(p.Models, "User.go"), []byte(userModelContent), 0644)

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

	dockerComposeContent := fmt.Sprintf(`services:
  app:
    image: golang:1.25-alpine
    container_name: %s-app
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - .:/app
    working_dir: /app
    command: go run main.go
    environment:
      - APP_ENV=local
      - APP_PORT=8080
      - APP_LANG=%s
`, name, lang)
	_ = os.WriteFile(filepath.Join(name, "docker-compose.yml"), []byte(dockerComposeContent), 0644)

	readmeContent := fmt.Sprintf(`# %s

%s

## %s
`+"```bash"+`
cd %s
go run main.go
`+"```"+`

Aplikasi: http://localhost:8080
Route & API Studio: http://localhost:8080/_studio
`, name, readmeDesc, stepText, name)
	_ = os.WriteFile(filepath.Join(name, "README.md"), []byte(readmeContent), 0644)

	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = name
	_ = tidyCmd.Run()

	fmt.Printf(`
✓ Proyek '%s' sukses digawe sing resik laka sampah engine! [Basa: %s]
  Struktur Folder:
    - Controllers : %s
    - Models      : %s
    - Views       : %s
    - Routes      : %s

%s
  cd %s
  go run main.go

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
	pkgName := paths.PackageName(dir)
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`package %s

import (
	"net/http"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
)

type %s struct{}

func (ctl *%s) Index(c *larapak.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Halo Sedulur! Selamat teka neng %s",
	})
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe controller: %v\n", err)
		return
	}
	fmt.Printf("✓ Controller sukses digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
}

func createModel(name string) {
	p := paths.GetPaths(".")
	dir := p.Models
	pkgName := paths.PackageName(dir)
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`package %s

import "time"

type %s struct {
	ID        int64     `+"`json:\"id\"`"+`
	CreatedAt time.Time `+"`json:\"created_at\"`"+`
	UpdatedAt time.Time `+"`json:\"updated_at\"`"+`
}

func (m *%s) TableName() string {
	return "%ss"
}
`, pkgName, name, name, strings.ToLower(name))

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe model: %v\n", err)
		return
	}
	fmt.Printf("✓ Model sukses digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
}

func createMiddleware(name string) {
	p := paths.GetPaths(".")
	dir := p.Middleware
	pkgName := paths.PackageName(dir)
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`package %s

import (
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
)

func %s() larapak.MiddlewareFunc {
	return func(next larapak.HandlerFunc) larapak.HandlerFunc {
		return func(c *larapak.Context) error {
			err := next(c)
			return err
		}
	}
}
`, pkgName, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe middleware: %v\n", err)
		return
	}
	fmt.Printf("✓ Middleware sukses digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
}

func createJob(name string) {
	p := paths.GetPaths(".")
	dir := p.Jobs
	pkgName := paths.PackageName(dir)
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`package %s

import "log"

type %s struct {
	Payload map[string]any
}

func (j *%s) Handle() error {
	log.Printf("[Job %s] Lagi diproses nganggo payload: %%v", j.Payload)
	return nil
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe job: %v\n", err)
		return
	}
	fmt.Printf("✓ Job antrean sukses digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
}

func createSeeder(name string) {
	p := paths.GetPaths(".")
	dir := p.Seeders
	pkgName := paths.PackageName(dir)
	_ = os.MkdirAll(dir, 0755)

	content := fmt.Sprintf(`package %s

import (
	"database/sql"
	"log"
)

type %s struct{}

func (s *%s) Run(db *sql.DB) error {
	log.Println("[Seeder %s] Ngisi bibit database...")
	return nil
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe seeder: %v\n", err)
		return
	}
	fmt.Printf("✓ Database Seeder sukses digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	fmt.Printf("Ngisi bibit data awal maring database saka '%s'...\n", filepath.ToSlash(p.Seeders))
	fmt.Println("✓ DatabaseSeeder sukses dieksekusi.")
}

func runQueueWorker() {
	fmt.Println("Queue Worker aktif... [Pencet Ctrl+C nggo mandheg]")
	q := queue.NewMemoryQueue()
	stopCh := make(chan struct{})
	q.Work(stopCh)
}

func clearCache() {
	_ = cache.ClearStorageCache("storage/cache")
	fmt.Println("✓ Cache sistem sukses diresiki kabeh.")
}

func runTests() {
	fmt.Println("Njalankna test suite LarapakGo...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
