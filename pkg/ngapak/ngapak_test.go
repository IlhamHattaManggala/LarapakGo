package ngapak

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IlhamHattaManggala/larapak-go/pkg/lang"
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

func TestEnvHelperInNgapak(t *testing.T) {
	os.Setenv("LARAPAK_TEST_KEY", "banyumas_mendoan")
	defer os.Unsetenv("LARAPAK_TEST_KEY")

	vEngine := view.New(t.TempDir())
	ngpkEngine := NewEngine(t.TempDir(), vEngine)
	app := larapak.New()
	ngpkEngine.BindRoutes(app)

	script := `
val1 = env("LARAPAK_TEST_KEY", "standar")
val2 = env("KEY_ORA_ANA", "standar_fallback")
val3 = lingkungan("LARAPAK_TEST_KEY")
`
	_, err := ngpkEngine.RunSource(script, "env_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource env test failed: %v", err)
	}

	if v, _ := ngpkEngine.Evaluator.Globals.Get("val1"); v != "banyumas_mendoan" {
		t.Errorf("expected banyumas_mendoan, got: %v", v)
	}
	if v, _ := ngpkEngine.Evaluator.Globals.Get("val2"); v != "standar_fallback" {
		t.Errorf("expected standar_fallback, got: %v", v)
	}
	if v, _ := ngpkEngine.Evaluator.Globals.Get("val3"); v != "banyumas_mendoan" {
		t.Errorf("expected banyumas_mendoan, got: %v", v)
	}
}

func TestNativeHttpClientInNgapak(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"sukses","pesan":"halo saka test server"}`))
			return
		}
		if r.Method == "POST" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"digawe":true}`))
			return
		}
	}))
	defer ts.Close()

	vEngine := view.New(t.TempDir())
	ngpkEngine := NewEngine(t.TempDir(), vEngine)
	app := larapak.New()
	ngpkEngine.BindRoutes(app)

	script := `
respGet = http.get("` + ts.URL + `")
statusGet = respGet.status
okGet = respGet.ok

respPost = http.post("` + ts.URL + `", {"item": "mendoan"})
statusPost = respPost.status
okPost = respPost.ok
`
	_, err := ngpkEngine.RunSource(script, "http_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource http test failed: %v", err)
	}

	if s, _ := ngpkEngine.Evaluator.Globals.Get("statusGet"); s != 200 {
		t.Errorf("expected statusGet 200, got %v", s)
	}
	if ok, _ := ngpkEngine.Evaluator.Globals.Get("okGet"); ok != true {
		t.Errorf("expected okGet true, got %v", ok)
	}
	if s, _ := ngpkEngine.Evaluator.Globals.Get("statusPost"); s != 201 {
		t.Errorf("expected statusPost 201, got %v", s)
	}
	if ok, _ := ngpkEngine.Evaluator.Globals.Get("okPost"); ok != true {
		t.Errorf("expected okPost true, got %v", ok)
	}
}

func TestValidationAndOldInputInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
rute.post("/daftar", gawe(req, res)
    v = req.validasi({
        "jeneng": "required|min:3",
        "email": "required|email"
    })
    nek v.fails() ya
        balekna res.karo_input().karo_kliru(v.errors()).balek("/formulir")
    rampung
    balekna res.json({"status": "sukses"})
rampung)
`
	_, err := ngpkEngine.RunSource(script, "validasi_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource validasi test failed: %v", err)
	}

	// Submit invalid POST (jeneng is empty)
	form := strings.NewReader("jeneng=&email=dudu-email")
	req := httptest.NewRequest("POST", "/daftar", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", "/formulir")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected 302 Found, got: %d", w.Code)
	}
	if w.Header().Get("Location") != "/formulir" {
		t.Errorf("expected Location /formulir, got: %s", w.Header().Get("Location"))
	}
}

func TestMailAndStorageInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
mailRes = email.kirim({
    "marang": "user@example.com",
    "subyek": "Sugeng Rawuh",
    "isi": "Isi layang email Larapak"
})

storage.simpen("cathetan.txt", "Teks Rahasia")
adaFile = storage.ana("cathetan.txt")
isiFile = storage.jupuk("cathetan.txt")
urlFile = storage.dalan("cathetan.txt")
`
	_, err := ngpkEngine.RunSource(script, "mail_storage_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource mail & storage test failed: %v", err)
	}

	if ada, _ := ngpkEngine.Evaluator.Globals.Get("adaFile"); ada != true {
		t.Errorf("expected adaFile true, got %v", ada)
	}
	if isi, _ := ngpkEngine.Evaluator.Globals.Get("isiFile"); isi != "Teks Rahasia" {
		t.Errorf("expected isiFile 'Teks Rahasia', got %v", isi)
	}
	if url, _ := ngpkEngine.Evaluator.Globals.Get("urlFile"); url != "/storage/cathetan.txt" {
		t.Errorf("expected urlFile '/storage/cathetan.txt', got %v", url)
	}
}

func TestRouteThrottleInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
rute.get("/api/terbatas", gawe(req, res)
    balekna res.json({"pesan": "aman"})
rampung).wates(1, 10)
`
	_, err := ngpkEngine.RunSource(script, "throttle_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource throttle test failed: %v", err)
	}

	// 1st request -> 200 OK
	req1 := httptest.NewRequest("GET", "/api/terbatas", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200 on 1st request, got %d", w1.Code)
	}

	// 2nd request -> 429 Too Many Requests
	req2 := httptest.NewRequest("GET", "/api/terbatas", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 2nd request, got %d", w2.Code)
	}
}

func TestModelRelationshipsInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
user = {
    "id": 7,
    "jeneng": "Slamet"
}
hasManySql = db.duwe_akeh(user, "artikels").toSql()

artikel = {
    "id": 99,
    "user_id": 7,
    "judul": "Sinau Ngapak"
}
belongsToSql = db.melu(artikel, "users").toSql()
`
	_, err := ngpkEngine.RunSource(script, "rel_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource failed: %v", err)
	}

	hSql, _ := ngpkEngine.Evaluator.Globals.Get("hasManySql")
	if !strings.Contains(fmt.Sprintf("%v", hSql), "WHERE user_id = ?") {
		t.Fatalf("expected user_id in hasManySql, got: %v", hSql)
	}

	bSql, _ := ngpkEngine.Evaluator.Globals.Get("belongsToSql")
	if !strings.Contains(fmt.Sprintf("%v", bSql), "WHERE id = ?") {
		t.Fatalf("expected id = ? in belongsToSql, got: %v", bSql)
	}
}

func TestFactoryInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
nama = pabrik.jeneng()
surel = pabrik.email("Agus Santoso")
telepon = pabrik.hp()
kutha = pabrik.kutha()
angka = pabrik.angka(10, 50)
rega = pabrik.rega(10000, 50000)
`
	_, err := ngpkEngine.RunSource(script, "factory_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource factory failed: %v", err)
	}

	n, _ := ngpkEngine.Evaluator.Globals.Get("nama")
	if n == "" || !strings.Contains(fmt.Sprintf("%v", n), " ") {
		t.Fatalf("invalid nama from pabrik: %v", n)
	}

	s, _ := ngpkEngine.Evaluator.Globals.Get("surel")
	if !strings.Contains(fmt.Sprintf("%v", s), "agus.santoso@") {
		t.Fatalf("invalid surel from pabrik: %v", s)
	}

	tel, _ := ngpkEngine.Evaluator.Globals.Get("telepon")
	if !strings.HasPrefix(fmt.Sprintf("%v", tel), "08") {
		t.Fatalf("invalid hp from pabrik: %v", tel)
	}
}

func TestRouteCacheInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
rute.get("/api/produk-cepet", gawe(req, res)
    balekna res.json({"status": "sukses", "produk": "Mendoan"})
rampung).cache(5)
`
	_, err := ngpkEngine.RunSource(script, "cache_route_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource route cache failed: %v", err)
	}

	// 1st request -> MISS
	req1 := httptest.NewRequest("GET", "/api/produk-cepet", nil)
	w1 := httptest.NewRecorder()
	app.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w1.Code)
	}
	if w1.Header().Get("X-Larapak-Cache") != "MISS" {
		t.Fatalf("expected MISS on 1st request, got: %s", w1.Header().Get("X-Larapak-Cache"))
	}

	// 2nd request -> HIT
	req2 := httptest.NewRequest("GET", "/api/produk-cepet", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200 on 2nd, got %d", w2.Code)
	}
	if w2.Header().Get("X-Larapak-Cache") != "HIT" {
		t.Fatalf("expected HIT on 2nd request, got: %s", w2.Header().Get("X-Larapak-Cache"))
	}
}

func TestApiResourceTransformerInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
rute.get("/api/user-profil", gawe(req, res)
    user = {
        "id": 1,
        "jeneng": "Dewi Lestari",
        "email": "dewi@example.com",
        "password_hash": "rahasia_aja_disebar"
    }
    balekna res.resource(user, gawe(u)
        balekna {
            "id": u.id,
            "jeneng": u.jeneng,
            "email": u.email
        }
    rampung, {"pesan": "Sukses"})
rampung)
`
	_, err := ngpkEngine.RunSource(script, "resource_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource resource test failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/user-profil", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, "password_hash") {
		t.Fatalf("expected password_hash to be omitted by transformer: %s", body)
	}
	if !strings.Contains(body, `"data"`) || !strings.Contains(body, `"Dewi Lestari"`) {
		t.Fatalf("expected data with Dewi Lestari: %s", body)
	}
	if !strings.Contains(body, `"meta"`) || !strings.Contains(body, `"Sukses"`) {
		t.Fatalf("expected meta with Sukses: %s", body)
	}
}

func TestBasaAndLangInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	lang.GetTranslator().AddMessages("id", map[string]string{
		"salam": "Halo :jeneng, sugeng rawuh!",
	})
	lang.SetLocale("id")

	script := `
pesan = basa.jupuk("salam", {"jeneng": "Kurniawan"})
basa_saiki = basa.saiki()
`
	_, err := ngpkEngine.RunSource(script, "basa_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource basa failed: %v", err)
	}

	p, _ := ngpkEngine.Evaluator.Globals.Get("pesan")
	if p != "Halo Kurniawan, sugeng rawuh!" {
		t.Fatalf("unexpected pesan: %v", p)
	}

	b, _ := ngpkEngine.Evaluator.Globals.Get("basa_saiki")
	if b != "id" {
		t.Fatalf("unexpected basa_saiki: %v", b)
	}
}

func TestKedadeyanAndEventInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
hasil_notifikasi = ""
kedadeyan.rungokna("pesanan.digawe", gawe(ev, payload)
    hasil_notifikasi = "Pesanan mlebu: " + payload.kode
rampung)

kedadeyan.kirim("pesanan.digawe", {"kode": "INV-999"})
`
	_, err := ngpkEngine.RunSource(script, "event_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource event failed: %v", err)
	}

	h, _ := ngpkEngine.Evaluator.Globals.Get("hasil_notifikasi")
	if h != "Pesanan mlebu: INV-999" {
		t.Fatalf("unexpected hasil_notifikasi: %v", h)
	}
}

func TestSoftDeletesInNgapak(t *testing.T) {
	tmpDir := t.TempDir()
	app := larapak.New()
	vEngine := view.New(tmpDir)
	ngpkEngine := NewEngine(tmpDir, vEngine)
	ngpkEngine.BindRoutes(app)

	script := `
sql1 = basisdata.tabel("produks").mbusak_alus().neng("kategori", "elektronik").toSql()
sql2 = basisdata.tabel("produks").mbusak_alus().karo_dibusek().toSql()
sql3 = basisdata.tabel("produks").mbusak_alus().mung_dibusek().toSql()
`
	_, err := ngpkEngine.RunSource(script, "soft_delete_test.ngpk")
	if err != nil {
		t.Fatalf("RunSource soft delete failed: %v", err)
	}

	s1, _ := ngpkEngine.Evaluator.Globals.Get("sql1")
	if !strings.Contains(fmt.Sprintf("%v", s1), "deleted_at IS NULL") {
		t.Fatalf("expected deleted_at IS NULL in sql1, got: %v", s1)
	}

	s2, _ := ngpkEngine.Evaluator.Globals.Get("sql2")
	if strings.Contains(fmt.Sprintf("%v", s2), "deleted_at IS NULL") {
		t.Fatalf("expected karo_dibusek to not contain deleted_at IS NULL, got: %v", s2)
	}

	s3, _ := ngpkEngine.Evaluator.Globals.Get("sql3")
	if !strings.Contains(fmt.Sprintf("%v", s3), "deleted_at IS NOT NULL") {
		t.Fatalf("expected mung_dibusek to contain deleted_at IS NOT NULL, got: %v", s3)
	}
}





