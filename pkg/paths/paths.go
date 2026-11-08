package paths

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// ProjectPaths defines directory mappings for a specific language scaffolding.
type ProjectPaths struct {
	App          string
	Controllers  string
	Models       string
	Middleware   string
	Jobs         string
	Live         string
	Resources    string
	Database     string
	Migrations   string
	Seeders      string
	Factories    string
	Routes       string
	Views        string
	Public       string
	PublicCSS    string
	PublicJS     string
	PublicImages string
}

// Config maps language identifiers ("ngapak", "indonesia", "english") to ProjectPaths.
var Config = map[string]ProjectPaths{
	"ngapak": {
		App:          "jeroan",
		Controllers:  filepath.Join("jeroan", "Pangatur"),
		Models:       filepath.Join("jeroan", "Wujud"),
		Middleware:   filepath.Join("jeroan", "Pencegat"),
		Jobs:         filepath.Join("jeroan", "Gawean"),
		Live:         filepath.Join("jeroan", "Urip"),
		Resources:    filepath.Join("jeroan", "Sumber"),
		Database:     "lumbung",
		Migrations:   filepath.Join("lumbung", "owahan"),
		Seeders:      filepath.Join("lumbung", "bibit"),
		Factories:    filepath.Join("lumbung", "pabrik"),
		Routes:       "dalan",
		Views:        "sawangan",
		Public:       "njobo",
		PublicCSS:    filepath.Join("njobo", "css"),
		PublicJS:     filepath.Join("njobo", "js"),
		PublicImages: filepath.Join("njobo", "images"),
	},
	"indonesia": {
		App:          "aplikasi",
		Controllers:  filepath.Join("aplikasi", "Pengendali"),
		Models:       filepath.Join("aplikasi", "Model"),
		Middleware:   filepath.Join("aplikasi", "Penyaring"),
		Jobs:         filepath.Join("aplikasi", "Tugas"),
		Live:         filepath.Join("aplikasi", "Live"),
		Resources:    filepath.Join("aplikasi", "Sumber"),
		Database:     "basisdata",
		Migrations:   filepath.Join("basisdata", "migrasi"),
		Seeders:      filepath.Join("basisdata", "seeder"),
		Factories:    filepath.Join("basisdata", "pabrik"),
		Routes:       "rute",
		Views:        "tampilan",
		Public:       "publik",
		PublicCSS:    filepath.Join("publik", "css"),
		PublicJS:     filepath.Join("publik", "js"),
		PublicImages: filepath.Join("publik", "gambar"),
	},
	"english": {
		App:          "app",
		Controllers:  filepath.Join("app", "Controllers"),
		Models:       filepath.Join("app", "Models"),
		Middleware:   filepath.Join("app", "Middleware"),
		Jobs:         filepath.Join("app", "Jobs"),
		Live:         filepath.Join("app", "Live"),
		Resources:    filepath.Join("app", "Resources"),
		Database:     "database",
		Migrations:   filepath.Join("database", "migrations"),
		Seeders:      filepath.Join("database", "seeders"),
		Factories:    filepath.Join("database", "factories"),
		Routes:       "routes",
		Views:        filepath.Join("resources", "views"),
		Public:       "public",
		PublicCSS:    filepath.Join("public", "css"),
		PublicJS:     filepath.Join("public", "js"),
		PublicImages: filepath.Join("public", "images"),
	},
}

// NormalizeLanguage normalizes input strings to "ngapak", "indonesia", or "english".
func NormalizeLanguage(lang string) string {
	l := strings.ToLower(strings.TrimSpace(lang))
	switch l {
	case "ngapak", "ngpk", "banyumasan":
		return "ngapak"
	case "indonesia", "idn", "id":
		return "indonesia"
	case "english", "en", "inggris":
		return "english"
	default:
		return "english"
	}
}

// DetectProjectLanguage auto-detects the project's language from .env or directory structure.
func DetectProjectLanguage(baseDir string) string {
	if baseDir == "" {
		baseDir = "."
	}

	// 1. Check .env for APP_LANG
	envFile := filepath.Join(baseDir, ".env")
	if file, err := os.Open(envFile); err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "APP_LANG=") {
				val := strings.TrimPrefix(line, "APP_LANG=")
				return NormalizeLanguage(val)
			}
		}
	}

	// 2. Check directory existence indicators
	if exists(filepath.Join(baseDir, "jeroan")) || exists(filepath.Join(baseDir, "sawangan")) {
		return "ngapak"
	}
	if exists(filepath.Join(baseDir, "aplikasi")) || exists(filepath.Join(baseDir, "tampilan")) {
		return "indonesia"
	}
	return "english"
}

// PackageName returns the Go package name from a folder path (e.g. "jeroan/Pangatur" -> "pangatur").
func PackageName(dirPath string) string {
	base := filepath.Base(dirPath)
	return strings.ToLower(base)
}

// InDir returns a new ProjectPaths with all relative paths prefixed by baseDir.
func (p ProjectPaths) InDir(baseDir string) ProjectPaths {
	if baseDir == "" || baseDir == "." {
		return p
	}
	return ProjectPaths{
		App:          filepath.Join(baseDir, p.App),
		Controllers:  filepath.Join(baseDir, p.Controllers),
		Models:       filepath.Join(baseDir, p.Models),
		Middleware:   filepath.Join(baseDir, p.Middleware),
		Jobs:         filepath.Join(baseDir, p.Jobs),
		Live:         filepath.Join(baseDir, p.Live),
		Resources:    filepath.Join(baseDir, p.Resources),
		Database:     filepath.Join(baseDir, p.Database),
		Migrations:   filepath.Join(baseDir, p.Migrations),
		Seeders:      filepath.Join(baseDir, p.Seeders),
		Factories:    filepath.Join(baseDir, p.Factories),
		Routes:       filepath.Join(baseDir, p.Routes),
		Views:        filepath.Join(baseDir, p.Views),
		Public:       filepath.Join(baseDir, p.Public),
		PublicCSS:    filepath.Join(baseDir, p.PublicCSS),
		PublicJS:     filepath.Join(baseDir, p.PublicJS),
		PublicImages: filepath.Join(baseDir, p.PublicImages),
	}
}

// GetPaths returns the resolved ProjectPaths for a base directory.
func GetPaths(baseDir string) ProjectPaths {
	lang := DetectProjectLanguage(baseDir)
	return Config[lang].InDir(baseDir)
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
