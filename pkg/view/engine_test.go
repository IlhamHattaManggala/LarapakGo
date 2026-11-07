package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IlhamHattaManggala/larapak-go/pkg/rikatlive"
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

type sampleLiveComponent struct{}

func (s *sampleLiveComponent) Render() string {
	return `<span class="live-status">Aktif Realtime</span>`
}

func TestEngineRikatLiveDirective(t *testing.T) {
	rikatlive.Register("SampleLive", func() rikatlive.Component {
		return &sampleLiveComponent{}
	})

	engine := New("")
	tmpl := `<div class="box">
    @live("SampleLive")
</div>`

	out, err := engine.RenderString(tmpl, nil)
	if err != nil {
		t.Fatalf("failed to render live directive: %v", err)
	}

	if !strings.Contains(out, `r-component="SampleLive"`) {
		t.Fatalf("expected r-component in output:\n%s", out)
	}
	if !strings.Contains(out, `<span class="live-status">Aktif Realtime</span>`) {
		t.Fatalf("expected component inner HTML in output:\n%s", out)
	}
}

func TestEngineFlashAndSessionDirectives(t *testing.T) {
	engine := New("")
	tmpl := `<div class="container">
    @flash("pesan")
    <div class="alert alert-success">{{ message }}</div>
    @endflash

    @session("username")
    <span class="user">{{ value }}</span>
    @endsession
</div>`

	data := map[string]any{
		"flash": map[string]string{
			"pesan": "Operasi kasil!",
		},
		"session": map[string]any{
			"username": "budi_ngapak",
		},
	}

	out, err := engine.RenderString(tmpl, data)
	if err != nil {
		t.Fatalf("failed to render flash/session directives: %v", err)
	}

	if !strings.Contains(out, `<div class="alert alert-success">Operasi kasil!</div>`) {
		t.Fatalf("expected flash alert in output:\n%s", out)
	}
	if !strings.Contains(out, `<span class="user">budi_ngapak</span>`) {
		t.Fatalf("expected session span in output:\n%s", out)
	}
}

func TestEngineErrorDirective(t *testing.T) {
	engine := New("")
	tmpl := `<div class="form-group">
    @error("email")
    <span class="text-red-500">{{ message }}</span>
    @enderror

    @error("password")
    <span class="text-red-500">{{ pesan }}</span>
    @enderror

    @error("nama")
    <span class="text-red-500">{{ message }}</span>
    @enderror
</div>`

	data := map[string]any{
		"errors": map[string]string{
			"email":    "Email kudu diisi!",
			"password": "Password kurang dawa!",
		},
	}

	out, err := engine.RenderString(tmpl, data)
	if err != nil {
		t.Fatalf("failed to render @error directive: %v", err)
	}

	if !strings.Contains(out, `<span class="text-red-500">Email kudu diisi!</span>`) {
		t.Fatalf("expected email error in output:\n%s", out)
	}
	if !strings.Contains(out, `<span class="text-red-500">Password kurang dawa!</span>`) {
		t.Fatalf("expected password error in output:\n%s", out)
	}
	if strings.Contains(out, "nama") {
		t.Fatalf("expected no nama error rendered, but got:\n%s", out)
	}
}
