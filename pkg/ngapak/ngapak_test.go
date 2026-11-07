package ngapak

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
	"github.com/IlhamHattaManggala/larapak-go/pkg/view"
)

func TestLexer(t *testing.T) {
	src := `
# Iki komentar
jeneng = "Larapak"
umur = 25
bener_ora = bener
`
	lexer := NewLexer(src, "test.ngpk")
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}

	if len(tokens) == 0 {
		t.Fatal("expected tokens, got 0")
	}

	hasIdentifier := false
	for _, tok := range tokens {
		if tok.Type == T_IDENTIFIER && tok.Value == "jeneng" {
			hasIdentifier = true
		}
	}
	if !hasIdentifier {
		t.Error("expected identifier 'jeneng'")
	}
}

func TestEvaluatorBasicArithmetic(t *testing.T) {
	ev := NewEvaluator()
	var out bytes.Buffer
	ev.Output = &out

	code := `
a = 10
b = 20
c = a + b * 2
tulis c
`
	lexer := NewLexer(code, "test.ngpk")
	toks, err := lexer.Tokenize()
	if err != nil {
		t.Fatal(err)
	}

	p := NewParser(toks, "test.ngpk")
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ev.ExecuteProgram(prog)
	if err != nil {
		t.Fatal(err)
	}

	res := strings.TrimSpace(out.String())
	if res != "50" {
		t.Errorf("expected 50, got %q", res)
	}
}

func TestEvaluatorConditionalsAndLoops(t *testing.T) {
	ev := NewEvaluator()
	var out bytes.Buffer
	ev.Output = &out

	code := `
total = 0
baleni i saka 1 nganti 5
    nek i == 3 ya
        # skip 3
    liyane
        total = total + i
    rampung
rampung
tulis total
`
	lexer := NewLexer(code, "test.ngpk")
	toks, _ := lexer.Tokenize()
	p := NewParser(toks, "test.ngpk")
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ev.ExecuteProgram(prog)
	if err != nil {
		t.Fatal(err)
	}

	// 1 + 2 + 4 + 5 = 12
	res := strings.TrimSpace(out.String())
	if res != "12" {
		t.Errorf("expected 12, got %q", res)
	}
}

func TestEvaluatorFunctionsAndDict(t *testing.T) {
	ev := NewEvaluator()
	var out bytes.Buffer
	ev.Output = &out

	code := `
gawe tambah(x, y)
    balekna x + y
rampung

data = ["nama": "Banyumas", "poin": tambah(10, 5)]
tulis data["nama"] + " " + teks(data["poin"])
`
	lexer := NewLexer(code, "test.ngpk")
	toks, _ := lexer.Tokenize()
	p := NewParser(toks, "test.ngpk")
	prog, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ev.ExecuteProgram(prog)
	if err != nil {
		t.Fatal(err)
	}

	res := strings.TrimSpace(out.String())
	if res != "Banyumas 15" {
		t.Errorf("expected 'Banyumas 15', got %q", res)
	}
}

func TestBridgeRouteWithNgapakAndNview(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create a sample .nview template
	viewsDir := filepath.Join(tmpDir, "sawangan")
	_ = os.MkdirAll(viewsDir, 0755)
	tplContent := `<h1>Selamat Teka {{ nama }}</h1>`
	_ = os.WriteFile(filepath.Join(viewsDir, "salam.nview"), []byte(tplContent), 0644)

	// 2. Setup Larapak App & Ngapak Engine
	app := larapak.New()
	vEngine := view.New(viewsDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	// 3. Run a .ngpk script that registers routes
	routeScript := `
rute.get("/salam", gawe(req, res)
    balekna res.view("salam", ["nama": "Slamet"])
rampung)

rute.get("/api/info", gawe(req, res)
    balekna res.json(["status": "bener", "versi": "1.0.0"])
rampung)
`
	_, err := ngpkEngine.RunSource(routeScript, "web.ngpk")
	if err != nil {
		t.Fatalf("RunSource failed: %v", err)
	}

	// 4. Test GET /salam (HTML template from .nview)
	req := httptest.NewRequest("GET", "/salam", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Selamat Teka Slamet") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}

	// 5. Test GET /api/info (JSON response)
	reqAPI := httptest.NewRequest("GET", "/api/info", nil)
	wAPI := httptest.NewRecorder()
	app.ServeHTTP(wAPI, reqAPI)

	if wAPI.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", wAPI.Code)
	}
	if !strings.Contains(wAPI.Body.String(), `"status": "bener"`) {
		t.Errorf("unexpected json body: %s", wAPI.Body.String())
	}
}

func TestBridgeRouteWithControllerNgpk(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. Create a sample .nview template
	viewsDir := filepath.Join(tmpDir, "resources", "views")
	_ = os.MkdirAll(viewsDir, 0755)
	tplContent := `<h1>Sugeng Rawuh {{ title }}</h1>`
	_ = os.WriteFile(filepath.Join(viewsDir, "index.nview"), []byte(tplContent), 0644)

	// 2. Create HomeController.ngpk in app/Controllers
	ctrlDir := filepath.Join(tmpDir, "app", "Controllers")
	_ = os.MkdirAll(ctrlDir, 0755)
	ctrlContent := `# HomeController
gawe index(req, res)
    balekna res.view("index", ["title": "Larapak NgapakIn"])
rampung
`
	_ = os.WriteFile(filepath.Join(ctrlDir, "HomeController.ngpk"), []byte(ctrlContent), 0644)

	// 3. Setup Larapak App & Ngapak Engine
	app := larapak.New()
	vEngine := view.New(viewsDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	// 4. Run route that references "HomeController@index"
	routeScript := `
rute.get("/", "HomeController@index")
`
	_, err := ngpkEngine.RunSource(routeScript, "web.ngpk")
	if err != nil {
		t.Fatalf("RunSource failed: %v", err)
	}

	// 5. Test GET /
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Sugeng Rawuh Larapak NgapakIn") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestDbBridgeInNgapak(t *testing.T) {
	app := larapak.New()
	vEngine := view.New(t.TempDir())
	ngpkEngine := NewEngine(t.TempDir(), vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
sql = db.tabel("users").pilih("id", "jeneng").neng("status", "=", "aktif").wates(5).toSql()
`
	_, err := ngpkEngine.RunSource(script, "db_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource db test failed: %v", err)
	}

	val, ok := ngpkEngine.Evaluator.Globals.Get("sql")
	if !ok {
		t.Fatal("expected sql variable to be set in globals")
	}
	sqlStr, ok := val.(string)
	if !ok || !strings.Contains(sqlStr, "SELECT id, jeneng FROM users WHERE status = ? LIMIT 5") {
		t.Errorf("unexpected compiled SQL from NgapakIn: %s", sqlStr)
	}
}

func TestSandiBridgeInNgapak(t *testing.T) {
	app := larapak.New()
	vEngine := view.New(t.TempDir())
	ngpkEngine := NewEngine(t.TempDir(), vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
hashSandi = sandi.gawe("rahasia123")
isCocok = sandi.cocok("rahasia123", hashSandi)
isSalah = sandi.cocok("salah123", hashSandi)
`
	_, err := ngpkEngine.RunSource(script, "sandi_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource sandi test failed: %v", err)
	}

	valCocok, ok := ngpkEngine.Evaluator.Globals.Get("isCocok")
	if !ok || valCocok != true {
		t.Errorf("expected isCocok to be true, got: %v", valCocok)
	}

	valSalah, ok := ngpkEngine.Evaluator.Globals.Get("isSalah")
	if !ok || valSalah != false {
		t.Errorf("expected isSalah to be false, got: %v", valSalah)
	}
}

func TestFormInputAndRedirectInNgapak(t *testing.T) {
	app := larapak.New()
	vEngine := view.New(t.TempDir())
	ngpkEngine := NewEngine(t.TempDir(), vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
rute.post("/proses-login", gawe(req, res)
    email = req.input("email")
    nek email == "admin@ngapak.org" ya
        balekna res.redirect("/dashboard")
    liyane
        balekna res.json({"pesan": "email ora cocok"})
    rampung
rampung)
`
	_, err := ngpkEngine.RunSource(script, "login_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource login test failed: %v", err)
	}

	// 1. Submit valid form POST
	form := strings.NewReader("email=admin@ngapak.org&sandi=123")
	req := httptest.NewRequest("POST", "/proses-login", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 Found, got: %d", w.Code)
	}
	if w.Header().Get("Location") != "/dashboard" {
		t.Errorf("expected Location /dashboard, got: %s", w.Header().Get("Location"))
	}
}


