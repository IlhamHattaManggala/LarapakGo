package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ngapak", "ngapak"},
		{"NGPK", "ngapak"},
		{"banyumasan", "ngapak"},
		{"indonesia", "indonesia"},
		{"ID", "indonesia"},
		{"english", "english"},
		{"inggris", "english"},
		{"unknown", "english"},
	}

	for _, tc := range tests {
		got := NormalizeLanguage(tc.input)
		if got != tc.expected {
			t.Errorf("NormalizeLanguage(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}

func TestDetectProjectLanguage(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Default to english if empty
	if DetectProjectLanguage(tmpDir) != "english" {
		t.Fatal("expected default english")
	}

	// 2. Detect via jeroan folder -> ngapak
	_ = os.Mkdir(filepath.Join(tmpDir, "jeroan"), 0755)
	if DetectProjectLanguage(tmpDir) != "ngapak" {
		t.Fatal("expected ngapak when jeroan exists")
	}

	// 3. Detect via .env APP_LANG
	tmpDir2 := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir2, ".env"), []byte("APP_LANG=indonesia\n"), 0644)
	if DetectProjectLanguage(tmpDir2) != "indonesia" {
		t.Fatal("expected indonesia from .env")
	}
}

func TestPackageNameAndInDir(t *testing.T) {
	p := Config["ngapak"]
	if PackageName(p.Controllers) != "pangatur" {
		t.Errorf("PackageName(Controllers) = %q; want pangatur", PackageName(p.Controllers))
	}
	if PackageName(p.Models) != "wujud" {
		t.Errorf("PackageName(Models) = %q; want wujud", PackageName(p.Models))
	}

	inDir := p.InDir("my_app")
	expected := filepath.Join("my_app", "jeroan", "Pangatur")
	if inDir.Controllers != expected {
		t.Errorf("inDir.Controllers = %q; want %q", inDir.Controllers, expected)
	}
}
