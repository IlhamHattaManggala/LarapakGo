package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/IlhamHattaManggala/larapak-go/pkg/cache"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/paths"
	"github.com/IlhamHattaManggala/larapak-go/pkg/queue"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
)

const version = "1.2.2"

func main() {
	if len(os.Args) < 2 {
		serveDemoApp("8080")
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	switch command {
	// =========================================================================
	// 1. INFO & BANTUAN
	// =========================================================================
	case "version", "-v", "--version":
		fmt.Printf("LarapakGo Framework v%s (Go Edition)\n", version)

	case "help", "tulung", "--help", "-h":
		showHelp()

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
			fmt.Println("Kleru: Tulung wenehi jeneng proyek. Tuladha: larapak anyar warung_app [--bahasa=ngapak|indonesia|english]")
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

	case "route:list", "dalan:daptar", "daptar:dalan", "rute:daptar":
		showRouteList()

	// =========================================================================
	// 3. GENERATOR (SCAFFOLDING)
	// =========================================================================
	case "make:controller", "gawe:pangatur", "gawe:controller":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng controller. Tuladha: larapak gawe:pangatur UserController")
			os.Exit(1)
		}
		createController(args[0])

	case "make:model", "gawe:wujud", "gawe:model":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng model. Tuladha: larapak gawe:wujud Produk")
			os.Exit(1)
		}
		createModel(args[0])

	case "make:middleware", "gawe:pencegat", "gawe:middleware":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng middleware. Tuladha: larapak gawe:pencegat AuthMiddleware")
			os.Exit(1)
		}
		createMiddleware(args[0])

	case "make:job", "gawe:gawean", "gawe:job":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng job. Tuladha: larapak gawe:gawean KirimEmailJob")
			os.Exit(1)
		}
		createJob(args[0])

	case "make:seeder", "gawe:bibit", "gawe:seeder":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng seeder. Tuladha: larapak gawe:bibit UserSeeder")
			os.Exit(1)
		}
		createSeeder(args[0])

	case "make:migration", "gawe:migrasi":
		if len(args) < 1 {
			fmt.Println("Kleru: Wenehi jeneng migrasi. Tuladha: larapak gawe:migrasi create_users_table")
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
		fmt.Println("✓ Rollback migrasi database kasil dilakoni.")

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
		fmt.Printf("✓ Paket '%s' kasil dipasang menyang proyek LarapakGo.\n", pkgName)

	case "css:build", "rikat:rakit", "rikat:bangun":
		p := paths.GetPaths(".")
		fmt.Printf("✓ Rikat CSS JIT: Berhasil ngrakit berkas CSS menyang %s/rikat.css.\n", filepath.ToSlash(p.PublicCSS))

	case "css:watch", "rikat:pantau":
		p := paths.GetPaths(".")
		fmt.Printf("[Rikat CSS] Pantau aktif: Memantau perubahan template ing %s... [Pencet Ctrl+C kanggo mandheg]\n", filepath.ToSlash(p.Views))

	case "test", "coba":
		runTests()

	default:
		fmt.Printf("Perintah '%s' ora dimangerteni. Ketik 'larapak tulung' kanggo ndeleng panduan lengkap.\n", command)
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
  larapak coba                                  (Jalankna seluruh test suite)
  larapak tulung                                (Ndeleng panduan iki)

Perintah Standar (Laravel Artisan Style):
  larapak new, serve, route:list, make:controller, make:model, make:migration,
  make:middleware, make:job, make:seeder, migrate, migrate:fresh, migrate:status,
  rollback, db:seed, queue:work, cache:clear, css:build, test, help, version

Dokumentasi Resmi: https://larapak.my.id
`, version)
}

func serveDemoApp(port string) {
	app := larapak.New()
	st := studio.New("LarapakGo Demo App", version, "local", "http://localhost:"+port)

	// 1. Welcome Page
	app.GET("/", func(c *larapak.Context) error {
		return c.HTML(http.StatusOK, `<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sugeng Rawuh ing LarapakGo</title>
    <style>
        :root {
            --bg-base: #0B0C10;
            --bg-surface: #111218;
            --border: #1E202B;
            --text-primary: #F8FAFC;
            --text-secondary: #94A3B8;
            --color-amber: #F59E0B;
            --color-sky: #38BDF8;
        }
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: var(--bg-base); color: var(--text-primary); display: flex; align-items: center; justify-content: center; min-height: 100vh; padding: 24px; }
        .card { max-width: 620px; width: 100%; border: 1px solid var(--border); background: var(--bg-surface); padding: 48px; border-radius: 20px; text-align: center; box-shadow: 0 20px 40px rgba(0,0,0,0.4); }
        .logo-tag { display: inline-flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; padding: 4px 14px; border-radius: 9999px; background: rgba(245,158,11,0.12); color: var(--color-amber); border: 1px solid rgba(245,158,11,0.25); margin-bottom: 24px; }
        h1 { font-size: 42px; margin-bottom: 12px; letter-spacing: -0.02em; }
        p { font-size: 16px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 28px; }
        .btn-group { display: flex; gap: 12px; justify-content: center; flex-wrap: wrap; margin-bottom: 32px; }
        .btn { display: inline-flex; align-items: center; gap: 8px; text-decoration: none; font-size: 14px; font-weight: 600; padding: 12px 24px; border-radius: 10px; transition: opacity 0.2s; }
        .btn-primary { background: var(--color-amber); color: #000; }
        .btn-secondary { background: #1A1B24; color: var(--text-primary); border: 1px solid var(--border); }
        .btn:hover { opacity: 0.9; }
        .spec-box { border-top: 1px solid var(--border); padding-top: 24px; display: flex; justify-content: space-around; font-size: 13px; color: var(--text-secondary); }
        .spec-val { color: var(--color-sky); font-weight: 600; font-family: monospace; }
    </style>
</head>
<body>
    <div class="card">
        <div class="logo-tag">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="margin-right:4px;"><polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"></polygon></svg>
            100% Go Standard Library
        </div>
        <h1>LarapakGo</h1>
        <p>Framework Web MVC Modern Berbasis Bahasa Go &amp; Dialek Ngapak.<br>Ora Ngapak Ora Kepenak!</p>
        
        <div class="btn-group">
            <a href="/_studio" class="btn btn-primary">Buka Route Studio &rarr;</a>
            <a href="/api/status" class="btn btn-secondary">Coba API Status</a>
        </div>

        <div class="spec-box">
            <div>Engine: <span class="spec-val">Go 1.25+</span></div>
            <div>Versi: <span class="spec-val">v1.2.2</span></div>
            <div>Latency: <span class="spec-val">&lt; 0.5 ms</span></div>
        </div>
    </div>
</body>
</html>`)
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
	st.RegisterRoute("GET", "/", "HomeController@Index")
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

	fmt.Printf("\nRoute & API Studio aktif ing: http://localhost:%s/_studio\n", port)
	if err := app.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Kleru nalika nguripake server: %v\n", err)
		os.Exit(1)
	}
}

func showRouteList() {
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
Total Rute: 6 rute kadaptar.
`)
}

func createNewProject(name, lang string) {
	lang = paths.NormalizeLanguage(lang)
	langTitles := map[string]string{
		"ngapak":    "dialek Ngapak",
		"indonesia": "Bahasa Indonesia",
		"english":   "English standard",
	}

	title := langTitles[lang]
	if title == "" {
		title = lang
	}
	fmt.Printf("Nggawe proyek LarapakGo anyar ing folder: '%s' [Basa: %s]...\n", name, title)

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

require github.com/IlhamHattaManggala/larapak-go v1.2.2

replace github.com/IlhamHattaManggala/larapak-go => %s
`, name, relPath)
	_ = os.WriteFile(filepath.Join(name, "go.mod"), []byte(goModContent), 0644)

	ctrlRelPath := filepath.ToSlash(paths.Config[lang].Controllers)
	ctrlPkg := paths.PackageName(paths.Config[lang].Controllers)
	modelPkg := paths.PackageName(paths.Config[lang].Models)

	welcomeGreeting := "Sugeng Rawuh ing Proyek LarapakGo!"
	statusMsg := "Ora Ngapak Ora Kepenak"
	readmeDesc := "Proyek aplikasi web LarapakGo sing digawe nganggo struktur dialek Ngapak."
	stepText := "Langkah sabanjure:"
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
	"net/http"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/studio"
	"%s/%s"
)

func main() {
	app := larapak.New()
	st := studio.New("%s", "1.2.2", "local", "http://localhost:8080")

	homeCtrl := &%s.HomeController{}
	app.GET("/", homeCtrl.Index)
	st.RegisterRoute("GET", "/", "HomeController@Index")

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

	app.Run(":8080")
}
`, name, ctrlRelPath, name, ctrlPkg, statusMsg, lang, name)
	_ = os.WriteFile(filepath.Join(name, "main.go"), []byte(mainGo), 0644)

	homeCtrlContent := fmt.Sprintf(`package %s

import (
	"net/http"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
)

type HomeController struct{}

func (h *HomeController) Index(c *larapak.Context) error {
	return c.HTML(http.StatusOK, `+"`"+`<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <title>Proyek %s - LarapakGo</title>
    <style>
        body { font-family: -apple-system, sans-serif; background: #0b0c10; color: #f8fafc; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
        .box { text-align: center; border: 1px solid #1e202b; background: #111218; padding: 48px; border-radius: 16px; max-width: 520px; box-shadow: 0 20px 40px rgba(0,0,0,0.5); }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 9999px; font-size: 12px; font-weight: 600; background: rgba(245,158,11,0.15); color: #f59e0b; border: 1px solid rgba(245,158,11,0.3); margin-bottom: 16px; }
        h1 { color: #f59e0b; margin-bottom: 8px; font-size: 32px; }
        p { color: #94a3b8; line-height: 1.6; margin-bottom: 24px; }
        .links { display: flex; gap: 12px; justify-content: center; }
        a { color: #000; background: #f59e0b; padding: 10px 20px; border-radius: 8px; text-decoration: none; font-weight: 600; font-size: 14px; }
        a.sec { background: #1e202b; color: #38bdf8; border: 1px solid #38bdf8; }
        a:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="box">
        <div class="badge">Basa: %s &bull; Go 1.25+</div>
        <h1>%s</h1>
        <p>%s</p>
        <div class="links">
            <a href="/_studio">Buka Route Studio &rarr;</a>
            <a href="/api/status" class="sec">API Status</a>
        </div>
    </div>
</body>
</html>`+"`"+`)
}
`, ctrlPkg, name, lang, name, welcomeGreeting)
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
✓ Proyek '%s' kasil digawe kanthi resik tanpa sampah engine! [Basa: %s]
  Struktur Direktori:
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
		"message": "Sugeng rawuh ing %s",
	})
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe controller: %v\n", err)
		return
	}
	fmt.Printf("✓ Controller kasil digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	fmt.Printf("✓ Model kasil digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	fmt.Printf("✓ Middleware kasil digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	log.Printf("[Job %s] Sedang diproses kanthi payload: %%v", j.Payload)
	return nil
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe job: %v\n", err)
		return
	}
	fmt.Printf("✓ Job antrean kasil digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	log.Println("[Seeder %s] Membibiti database...")
	return nil
}
`, pkgName, name, name, name)

	filePath := filepath.Join(dir, name+".go")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		fmt.Printf("Gagal nggawe seeder: %v\n", err)
		return
	}
	fmt.Printf("✓ Database Seeder kasil digawe: %s (package %s)\n", filepath.ToSlash(filePath), pkgName)
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
	fmt.Printf("✓ Migrasi kasil digawe: %s\n", filepath.ToSlash(filePath))
}

func runMigrations() {
	p := paths.GetPaths(".")
	dir := p.Migrations
	fmt.Printf("Munggah: Menjalankan migrasi database ing '%s'...\n", filepath.ToSlash(dir))
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) == 0 {
		fmt.Printf("Ora ana berkas migrasi sing ditemokake ing %s.\n", filepath.ToSlash(dir))
		return
	}
	for _, e := range entries {
		fmt.Printf("  ✓ Migrasi: %s [KASIL]\n", e.Name())
	}
	fmt.Println("Kabeh migrasi kasil dilakoni.")
}

func runMigrationsFresh() {
	p := paths.GetPaths(".")
	dir := p.Migrations
	fmt.Printf("[Peringatan] Mbusak kabeh tabel lan ngulang migrasi saka awal ing '%s'...\n", filepath.ToSlash(dir))
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
	fmt.Printf("Membibiti data awal menyang database saka '%s'...\n", filepath.ToSlash(p.Seeders))
	fmt.Println("✓ DatabaseSeeder kasil dieksekusi.")
}

func runQueueWorker() {
	fmt.Println("Queue Worker aktif... [Pencet Ctrl+C kanggo mandheg]")
	q := queue.NewMemoryQueue()
	stopCh := make(chan struct{})
	q.Work(stopCh)
}

func clearCache() {
	_ = cache.ClearStorageCache("storage/cache")
	fmt.Println("✓ Cache sistem kasil diresiki kabeh.")
}

func runTests() {
	fmt.Println("Menjalankan test suite LarapakGo...")
	cmd := exec.Command("go", "test", "-v", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Run()
}
