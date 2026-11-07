# LarapakGo

**Larapak MVC Web Framework versi Golang (Go 1.25+)**

LarapakGo adalah adaptasi performa tinggi dari ekosistem web framework Larapak yang ditulis murni dalam bahasa pemrograman **Go (Golang)**. Mengadopsi arsitektur lengkap ala Laravel, dialek lokal Ngapak Tegalan yang ramah dan lugas, serta performa tinggi khas Go dengan **100% Go Standard Library** tanpa ketergantungan paket eksternal pihak ketiga yang membengkak.

---

## Cara Instalasi Rikat (One-Line Installer)

Anda bisa langsung memasang binary mandiri `larapak` ke sistem operasi Anda hanya dengan satu baris perintah tanpa perlu mengompilasi manual:

### 1. Windows (PowerShell)
Buka terminal PowerShell lalu jalankan:
```powershell
irm https://raw.githubusercontent.com/IlhamHattaManggala/LarapakGo/main/install.ps1 | iex
```
*Installer ini otomatis mengunduh `larapak.exe` ke `$HOME\.larapak\bin` dan mendaftarkannya ke variabel Environment PATH User Anda.*

### 2. Linux & macOS (Bash / Zsh)
Buka terminal lalu jalankan:
```bash
curl -fsSL https://raw.githubusercontent.com/IlhamHattaManggala/LarapakGo/main/install.sh | bash
```
*Installer ini mendeteksi arsitektur sistem (`x86_64` / `arm64`), mengunduh binary ke `$HOME/.larapak/bin`, mengatur izin eksekusi (`chmod +x`), dan menyelaraskan `$PATH` profil shell Anda (`~/.bashrc` / `~/.zshrc`).*

---

## Dukungan Visual Studio Code (Syntax Highlighting)

Agar penulisan kode skrip `.ngpk` (NgapakIn) dan template `.nview` (Sawangan) nyaman dengan pewarnaan sintaks dan autocomplete snippets, pasang ekstensi resmi **`larapak-vscode`**:

### Cara Pasang Ekstensi:
1. Unduh berkas `larapak-vscode-1.0.0.vsix` dari folder `dist/` repositori ini.
2. Di terminal, jalankan perintah:
   ```bash
   code --install-extension dist/larapak-vscode-1.0.0.vsix
   ```
   *Atau dari VS Code: Buka tab Extensions (`Ctrl+Shift+X`) -> Klik ikon menu `...` di kanan atas -> Pilih **Install from VSIX...** -> Pilih berkas `larapak-vscode-1.0.0.vsix`.*

### Fitur Ekstensi:
- **Logo Berkas Asli (File Icons)**: Berkas `.ngpk` dan `.nview` otomatis memiliki logo resmi Larapak (`favicon.png`) di explorer tree VS Code (tersedia tema ikon *Larapak File Icons*, dan mendukung integrasi *vscode-icons* serta *Material Icon Theme*).
- **Pewarnaan Sintaks `.ngpk`**: Menyorot seluruh kata kunci (`nek`, `ya`, `liyane`, `rampung`, `baleni`, `saka`, `nganti`, `gawe`, `balekna`, `kelas`, `extends`, `mandeg`, `terusna`), operator logika (`lan`, `utawa`, `ora`, `neng`), konstanta (`bener`, `salah`, `kosong`), serta objek bawaan web (`rute`, `req`, `res`, `tulis`).
- **Pewarnaan Sintaks `.nview`**: Menyorot direktif template (`@layout`, `@section`, `@endsection`, `@yield`, `@csrf`, `@include`) dan interpolasi variabel (`{{ var }}`, `{!! raw !!}`).
- **Snippets Produktif**:
  - Di `.ngpk`: `rute:get`, `rute:post`, `gawe`, `nek`, `nek:liyane`, `baleni:range`, `baleni:neng`.
  - Di `.nview`: `layout`, `section`, `yield`, `csrf`, `var`, `raw`.

---

## Modul & Fitur Lengkap

### 1. HTTP Core & Router (`pkg/larapak`)
- **Trie / Parameterized Router**: Mendukung rute dinamis (`:id`, `*filepath`), route grouping (`app.Group()`), dan middleware chaining.
- **Context API Ekspresif**: Helper praktis `c.JSON()`, `c.HTML()`, `c.String()`, `c.BindJSON()`, `c.Param()`, `c.Query()`, `c.Header()`.
- **Middleware Terintegrasi**: Logger nanodetik, Panic Recovery otomatis, dan CORS handler.

### 2. Mesin Template Rikat Go (`pkg/view`)
- **Pewarisan Layout**: Mendukung `@layout`, `@section`, dan `@yield`.
- **Proteksi XSS Otomatis**: Interpolasi aman `{{ var }}` dengan auto-escaping, dan `{!! raw !!}` untuk HTML mentah.
- **Direktif CSRF**: `@csrf` otomatis menghasilkan input token hidden.

### 3. Keamanan & Sandi (`pkg/security`)
- **Auto-CSRF Protection**: Token 32-byte kriptografis aman, validasi *constant-time* anti-timing attack, dan pengecualian rute aman.
- **Password Hasher PBKDF2-SHA256**: Enkripsi password standar RFC 2898 tanpa butuh CGo/library eksternal (`HashPassword`, `VerifyPassword`, `NeedsRehash`).

### 4. Manajemen Sesi (`pkg/session`)
- **Signed Session Cookies**: Cookie sesi terenkripsi HMAC-SHA256 untuk mencegah *cookie tampering*.
- **Pesan Kilat (*Flash Messages*)**: Penyimpanan notifikasi satu kali baca untuk respon redirect pengguna.

### 5. Fluent Query Builder & Migrasi Database (`pkg/database`)
- **Query Builder**: Sintaks deklaratif `Table("users").Select(...).Where(...).OrderBy(...).Limit(...).ToSelectSQL()`.
- **Schema & Migrations**: Blueprint pembangun skema tabel database (`ID()`, `String()`, `Integer()`, `Text()`, `Boolean()`, `Timestamps()`).

### 6. Mesin Validasi Formulir (`pkg/validation`)
- Aturan validasi lengkap: `required`, `email`, `min:N`, `max:N`, `numeric`, `alpha`, `in:a,b,c`, `confirmed`.

### 7. Route & API Studio Bawaan (`pkg/studio`)
- Antarmuka GUI interaktif Route & API Studio di `/_studio` dengan palet tema modern gelap dan **100% SVG icons** (tanpa emoticon).
- Live testing request HTTP & pratinjau respons.
- Ekspor otomatis ke spesifikasi standar **OpenAPI 3.0.3** JSON di `/_studio/openapi.json`.

### 8. CLI & Multi-Language Directory Scaffolding (`cmd/larapak` & `pkg/paths`)
- Dukungan pembuatan proyek dengan tag bahasa `--bahasa=ngapak|indonesia|english` (sama persis dengan versi Python):
  - **Ngapak**: `larapak anyar toko_online --bahasa=ngapak`
    - Struktur: `jeroan/Pangatur`, `jeroan/Wujud`, `jeroan/Pencegat`, `jeroan/Gawean`, `lumbung/owahan`, `lumbung/bibit`, `sawangan/`, `dalan/`, `njobo/`.
    - Otomatis menghasilkan `dalan/web.ngpk` dan `sawangan/index.nview`.
  - **Indonesia**: `larapak anyar toko_online --bahasa=indonesia`
    - Struktur: `aplikasi/Pengendali`, `aplikasi/Model`, `aplikasi/Penyaring`, `aplikasi/Tugas`, `basisdata/migrasi`, `basisdata/seeder`, `tampilan/`, `rute/`, `publik/`.
    - Otomatis menghasilkan `rute/web.ngpk` dan `tampilan/index.nview`.
  - **English**: `larapak new my_app --bahasa=english`
    - Struktur: `app/Controllers`, `app/Models`, `app/Middleware`, `app/Jobs`, `database/migrations`, `database/seeders`, `resources/views/`, `routes/`, `public/`.
    - Otomatis menghasilkan `routes/web.ngpk` dan `resources/views/index.nview`.
- Dual-support 22 perintah bahasa Inggris & dialek Ngapak:
  - `larapak new <nama>` / `larapak anyar <nama> [--bahasa=...]`
  - `larapak serve [port]` / `larapak ladeni [port]` / `urubna` / `mlaku`
  - `larapak route:list` / `larapak dalan:daptar` (Otomatis membaca rute aktif dari `web.ngpk`)
  - `larapak run <skrip.ngpk>` / `larapak mlakukna <skrip.ngpk>` / `larapak eksekusi <skrip.ngpk>`
  - `larapak make:controller <Nama>` / `gawe:pangatur` / `gawe:controller`
  - `larapak make:model <Nama>` / `gawe:wujud` / `gawe:model`
  - `larapak make:middleware <Nama>` / `gawe:pencegat` / `gawe:middleware`
  - `larapak make:migration <Nama>` / `gawe:migrasi`
  - `larapak make:job <Nama>` / `gawe:gawean`
  - `larapak make:seeder <Nama>` / `gawe:bibit`
  - `larapak migrate` / `larapak migrasi` / `munggah`
  - `larapak migrate:fresh` / `larapak migrasi:anyar` / `migrasi:resik`
  - `larapak migrate:status` / `larapak migrasi:status`
  - `larapak rollback` / `larapak mudhun`
  - `larapak db:seed` / `larapak bibiti`
  - `larapak queue:work` / `larapak gawean:tandang`
  - `larapak cache:clear` / `larapak resik:cache`
  - `larapak css:build` / `larapak rikat:rakit`
  - `larapak test` / `larapak coba`
  - `larapak help` / `larapak tulung`

### 9. Mesin Interpreter Native NgapakIn (`.ngpk`) & Sawangan (`.nview`) (`pkg/ngapak`)
- **Native Go Interpreter (Tanpa Python & Tanpa CGo)**:
  - Dilengkapi Lexer, Parser, AST, dan Tree-Walk Evaluator mandiri berbasis 100% Go Standard Library.
  - Latency eksekusi sub-milidetik (< 0.5 ms) dengan konsumsi RAM sangat hemat (~10-20 MB).
- **Tata Bahasa NgapakIn Asli**:
  - `tulis(...)`, `nek ... ya ... liyane ... rampung`
  - `baleni ... saka ... nganti ... rampung`, `baleni ... neng ... rampung`
  - `gawe jeneng(param) ... rampung`, `balekna ...`
  - Nilai boolean & null: `bener`, `salah`, `kosong`
  - Koleksi: list `[1, 2, 3]` dan dictionary `{"kunci": "nilai"}`
- **Integrasi Web Route & HTTP Bridge**:
  - `rute.get("/", gawe(req, res) ... rampung)`
  - `rute.post("/api/simpen", "ProdukController@simpen")`
  - `res.view("index", {"judul": "..."})` untuk me-render template `.nview` secara instan.
  - `res.json({"status": "sukses"})` untuk API respons.
  - `req.params["id"]`, `req.query("cari")`, `req.method`.
  - Otomatis tersinkronisasi ke antarmuka Route & API Studio di `/_studio`.

---

## Memulai Cepat

```go
package main

import (
	"net/http"
	"github.com/IlhamHattaManggala/larapak-go/pkg/larapak"
)

func main() {
	app := larapak.New()

	app.GET("/", func(c *larapak.Context) error {
		return c.String(http.StatusOK, "Selamat Teka neng LarapakGo!")
	})

	app.Group("/api", func(r *larapak.Router) {
		r.GET("/status", func(c *larapak.Context) error {
			return c.JSON(http.StatusOK, map[string]any{
				"status":  "aktif",
				"pesan":   "Ora Ngapak Ora Kepenak",
				"version": "1.0.0",
			})
		})
	})

	app.Run(":8080")
}
```

---

## Struktur Arsitektur

```
LarapakGo/
├── go.mod                     # Modul Go
├── README.md                  # Dokumentasi
├── LICENSE                    # Lisensi MIT
├── .gitignore                 # Filter git
├── install.ps1                # Skrip instalasi satu baris untuk Windows (PowerShell)
├── install.sh                 # Skrip instalasi satu baris untuk Linux & macOS (Bash)
├── larapak.exe                # Binary mandiri root Windows
├── dist/                      # Rilis binary mandiri lintas platform & ekstensi
│   ├── larapak-v1.0.0-windows-amd64.exe
│   ├── larapak-v1.0.0-linux-amd64
│   ├── larapak-v1.0.0-darwin-arm64
│   └── larapak-vscode-1.0.0.vsix
├── editors/
│   └── vscode/                # Paket ekstensi Visual Studio Code untuk .ngpk & .nview
│       ├── package.json
│       ├── language-configuration-ngpk.json
│       ├── language-configuration-nview.json
│       ├── syntaxes/
│       │   ├── ngapak.tmLanguage.json
│       │   └── nview.tmLanguage.json
│       └── snippets/
│           ├── ngapak.json
│           └── nview.json
├── cmd/
│   └── larapak/               # Binary CLI & Project Scaffolder
│       └── main.go
└── pkg/
    ├── larapak/               # Application, Context, Router, Middleware
    ├── ngapak/                # Interpreter Native NgapakIn (.ngpk) & Runtime Bridge
    ├── view/                  # Rikat/Sawangan Go Template Engine (.nview, @layout)
    ├── database/              # Fluent Query Builder & Migration Blueprint
    ├── security/              # CSRF Protection & PBKDF2 Password Hashing
    ├── session/               # Encrypted Sessions & Flash Messages
    ├── validation/            # Input & Form Validation Engine
    ├── paths/                 # Scaffolding & Multi-Language Path Resolver
    └── studio/                # Route & API Studio UI & OpenAPI 3.0 Generator
```

---

## Lisensi
Lisensi MIT — Dibuat dengan bangga oleh [Ilham Hatta Manggala](https://ilhamhattamanggala.my.id).
