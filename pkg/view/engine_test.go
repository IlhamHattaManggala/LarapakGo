package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineRenderString(t *testing.T) {
	engine := New("")

	tmpl := `<div class="card">
    <h1>Selamat Teka {{ name }}</h1>
    <p>{!! rawHtml !!}</p>
    <form method="POST">
        @csrf
    </form>
</div>`

	data := map[string]any{
		"name":    "<b>Ilham</b>",                      // Should be escaped
		"rawHtml": "<span class='badge'>Ngapak</span>", // Should remain raw
		"_csrf":   "token_12345",
	}

	out, err := engine.RenderString(tmpl, data)
	if err != nil {
		t.Fatalf("failed to render string: %v", err)
	}

	if !strings.Contains(out, "&lt;b&gt;Ilham&lt;/b&gt;") {
		t.Fatalf("expected escaped HTML, got:\n%s", out)
	}
	if !strings.Contains(out, "<span class='badge'>Ngapak</span>") {
		t.Fatalf("expected raw HTML to be preserved, got:\n%s", out)
	}
	if !strings.Contains(out, `<input type="hidden" name="_csrf" value="token_12345">`) {
		t.Fatalf("expected csrf input, got:\n%s", out)
	}
}

func TestEngineLayoutInheritance(t *testing.T) {
	tmpDir := t.TempDir()

	layoutContent := `<!DOCTYPE html>
<html>
<head><title>@yield "title"</title></head>
<body>
    <main>@yield "content"</main>
</body>
</html>`
	_ = os.WriteFile(filepath.Join(tmpDir, "app.nview"), []byte(layoutContent), 0644)

	childContent := `@layout "app"
@section "title"
Kaca Utama LarapakGo
@endsection
@section "content"
<h1>Halo Donya!</h1>
@endsection`
	_ = os.WriteFile(filepath.Join(tmpDir, "home.nview"), []byte(childContent), 0644)

	engine := New(tmpDir)
	out, err := engine.RenderTemplate("home", nil)
	if err != nil {
		t.Fatalf("failed to render layout template: %v", err)
	}

	if !strings.Contains(out, "<title>Kaca Utama LarapakGo</title>") {
		t.Fatalf("title section not yielded properly, got:\n%s", out)
	}
	if !strings.Contains(out, "<main><h1>Halo Donya!</h1></main>") {
		t.Fatalf("content section not yielded properly, got:\n%s", out)
	}
}
