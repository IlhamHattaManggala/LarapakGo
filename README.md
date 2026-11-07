<p align="center">
  <img src="public/favicon.png" width="140" alt="LarapakGo Logo" />
</p>

# LarapakGo

[![Release Version](https://img.shields.io/badge/version-v1.0.0-orange.svg)](https://github.com/IlhamHattaManggala/LarapakGo/releases)
[![Go Version](https://img.shields.io/badge/go-1.21%20%7C%201.22%20%7C%201.23%20%7C%201.24%20%7C%201.25-00ADD8.svg?logo=go)](https://golang.org)
[![Dependencies](https://img.shields.io/badge/dependencies-0%20external-brightgreen.svg)](https://golang.org/pkg/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20Linux%20%7C%20macOS-blue.svg)](https://github.com/IlhamHattaManggala/LarapakGo/releases)

**Larapak MVC Web Framework versi Golang (Go 1.21+)**

LarapakGo adalah adaptasi performa tinggi dari ekosistem web framework Larapak yang ditulis murni dalam bahasa pemrograman **Go (Golang)**. Mengadopsi arsitektur lengkap ala Laravel, dialek lokal Ngapak Tegalan yang ramah dan lugas, serta performa tinggi khas Go dengan **100% Go Standard Library** tanpa ketergantungan paket eksternal pihak ketiga yang membengkak.

---

## Navigasi Dokumentasi Resmi

Untuk panduan mendalam dan tata kelola proyek, silakan rujuk dokumen resmi berikut:

- [Kebijakan Keamanan (SECURITY.md)](SECURITY.md) — Panduan arsitektur keamanan, proteksi CSRF, XSS, hashing PBKDF2, dan pelaporan kerentanan.
- [Panduan Deployment Produksi (DEPLOYMENT.md)](DEPLOYMENT.md) — Langkah deployment ke VPS Linux (Systemd & Nginx), Docker Compose, dan Cloud PaaS.
- [Panduan Kontribusi (CONTRIBUTING.md)](CONTRIBUTING.md) — Alur kerja kontribusi, strategi percabangan Git, standar kode, dan unit testing.
- [Kode Etik Komunitas (CODE_OF_CONDUCT.md)](CODE_OF_CONDUCT.md) — Standar perilaku dan komitmen lingkungan komunitas yang inklusif.
- [Catatan Perubahan (CHANGELOG.md)](CHANGELOG.md) — Riwayat rilis fitur, perbaikan bug, dan evolusi proyek sesuai Semantic Versioning.
- [Lisensi Perangkat Lunak (LICENSE)](LICENSE) — Ketentuan lisensi terbuka MIT.

---

## Cara Instalasi Rikat

Anda bisa langsung memasang LarapakGo & NgapakIn ke sistem operasi Anda:

### 1. Windows (Installer Resmi .exe)
Unduh berkas instalasi **`LarapakGo-Setup-v1.0.0.exe`** dari rilis resmi, lalu buka dan ikuti wizard instalasi (*Next > Next > Finish*).

Installer ini secara otomatis:
- Memasang binary `larapak.exe` dan alias `ngapakin.exe`.
- Mendaftarkan path ke variabel lingkungan `PATH` Windows secara otomatis (tanpa perlu pengaturan manual).
- Menyediakan uninstaller resmi yang terintegrasi di Windows Settings (*Installed Apps*).

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
- **Form Inputs & Redirects**: `c.FormValue("nama")`, `c.PostForm("email")`, `c.AllInputs()`, dan `c.Redirect(code, targetURL)`.
- **Penanganan Unggah Berkas (File Uploads)**: Fungsi `c.FormFile("berkas")`, `c.SaveUploadedFile(fileHeader, tujuan)`, dan parsing multipart form yang aman.
- **Middleware Terintegrasi & Structured Logging**: Logger nanodetik, Panic Recovery otomatis, CORS handler, serta penulisan log terstruktur dan stack trace ke `storage/logs/larapak.log`.

### 2. Mesin Template Rikat Go (`pkg/view`)
- **Pewarisan Layout**: Mendukung `@layout`, `@section`, dan `@yield`.
- **Proteksi XSS Otomatis**: Interpolasi aman `{{ var }}` dengan auto-escaping, dan `{!! raw !!}` untuk HTML mentah.
- **Direktif CSRF**: `@csrf` otomatis menghasilkan input token hidden.
- **Pesan Kilat & Sesi**: Direktif `@flash("pesan") ... @endflash` (dengan variabel bawaan `{{ message }}` / `{{ pesan }}`) dan `@session("user") ... @endsession` (`{{ value }}` / `{{ nilai }}`).
- **Smart In-Memory Template Cache**: Optimasi render secepat kilat pada environment produksi (`APP_ENV=production`) dengan cache thread-safe (`sync.RWMutex`), serta auto-invalidation instan pada mode pengembangan.

### 3. Keamanan & Sandi (`pkg/security`)
- **Auto-CSRF Protection**: Token 32-byte kriptografis aman, validasi *constant-time* anti-timing attack, dan pengecualian rute aman.
- **Password Hasher PBKDF2-SHA256**: Enkripsi password standar RFC 2898 tanpa butuh CGo/library eksternal (`HashPassword`, `VerifyPassword`, `NeedsRehash`).
- **Jembatan Sandi di .ngpk**: Fungsi global `sandi.gawe(pass)` / `hash.make(pass)` dan `sandi.cocok(pass, hash)` / `hash.check(pass, hash)` untuk sistem auth login/register native di NgapakIn.

### 4. Manajemen Sesi (`pkg/session`)
- **Signed Session Cookies**: Cookie sesi terenkripsi HMAC-SHA256 untuk mencegah *cookie tampering*.
- **Pesan Kilat (*Flash Messages*)**: Penyimpanan notifikasi satu kali baca (`sesi.kilat("pesan", "...")` / `session.flash(...)`) untuk respon redirect pengguna.
- **Akses Sesi di .ngpk**: Objek `sesi.simpen(...)`, `sesi.jupuk(...)`, `sesi.busak(...)`, `sesi.kilat(...)`.

### 5. Multi-Database Manager, SQL Dialect Transformer & Migrasi (`pkg/database`)
- **Multi-Connection Manager**: Mendukung banyak koneksi database sekaligus dalam satu aplikasi (`GetConnection("mysql")`, `GetConnection("pgsql")`, dll.).
- **SQL Dialect Transformer Otomatis**: Mendukung sintaks dialek khusus SQLite, MySQL/MariaDB, PostgreSQL, dan SQL Server (MSSQL).
- **Fluent Query Builder & Executor**: Sintaks deklaratif chaining:
  - Dialek Ngapak: `db.tabel("users").pilih("id", "jeneng").neng("aktif", 1).urut("id", "DESC").wates(10).jupuk()`
  - Dialek Internasional: `db.table("users").select(...).where(...).orderBy(...).limit(...).get()`
  - Eksekusi DML & Agregasi: `siji()` / `first()`, `cacah()` / `count()`, `tambah()` / `insert()`, `owahi()` / `update()`, `busak()` / `delete()`.
- **Pagination Bawaan**: Kalkulasi halaman otomatis via `db.tabel("produk").kaca(1, 10)` atau `db.table(...).paginate(page, perPage)` menghasilkan total baris, jumlah halaman, dan metadata navigasi secara rapi.
- **Schema & Migrations**: Blueprint pembangun skema tabel database (`ID()`, `String()`, `Integer()`, `Text()`, `Boolean()`, `Timestamps()`) yang otomatis beradaptasi dengan tipe data dialek target.

### 6. Mesin Validasi Formulir (`pkg/validation`) & Direktif `@error`
- Aturan validasi lengkap: `required`, `email`, `min:N`, `max:N`, `numeric`, `alpha`, `in:a,b,c`, `confirmed`.
- **Integrasi Penuh di Controller `.ngpk`**:
  ```python
  v = req.validasi({
      "jeneng": "required|min:3",
      "email": "required|email"
  })
  nek v.fails() ya
      balekna res.karo_input().karo_kliru(v.errors()).balek()
  rampung
  ```
- **Direktif `@error` di Template `.nview`**:
  ```html
  @error("jeneng")
      <span class="text-red-500 text-sm">{{ message }}</span>
  @enderror
  ```
- **Flashing Input Formulir Lawas (`old` / `lawas`)**: Nilai input sebelumnya otomatis tersimpan di sesi dan disajikan ke template via `{{ old.jeneng }}` / `{{ lawas.jeneng }}` atau di controller via `req.old("field")` / `req.lawas("field")`. Password otomatis disaring (*safe filtering*).

### 7. Native Redis Engine, Caching & Distributed Queues (`pkg/redis`, `pkg/cache`, `pkg/queue`)
- **Native Redis Client (100% Go Standard Library)**: Implementasi penuh protokol RESP (*REdis Serialization Protocol*) murni berbasis socket `net.Dial` dengan connection pool (`sync.Pool`), *auth*, multi-database index, tanpa butuh library pihak ketiga di `go.mod`.
- **Fleksibel Caching (`pkg/cache`)**: Dukungan driver terpadu `CACHE_DRIVER=memory|redis`. Fitur `Get()`, `Set(key, val, ttl)`, `Delete()`, dan `Flush()`.
- **Distributed Job Queues (`pkg/queue`)**: Pemrosesan antrean latar belakang `QUEUE_DRIVER=memory|redis` dengan metode non-busy-waiting `BRPOP` yang hemat CPU untuk arsitektur multi-server/multi-worker.

### 8. Route & API Studio Bawaan & Database Inspector (`pkg/studio`)
- **Route & API Studio Interaktif**: Halaman pengujian API interaktif di `/_studio` dengan palet tema modern gelap dan **100% SVG icons** (tanpa emoticon).
- **Database Inspector Visual**: Tab inspektur database bawaan untuk meninjau daftar tabel, struktur kolom, dan isi baris rekaman database (SQLite/MySQL/Postgres) langsung dari peramban web pada mode pengembangan lokal.
- Live testing request HTTP & pratinjau respons.
- Ekspor otomatis ke spesifikasi standar **OpenAPI 3.0.3** JSON di `/_studio/openapi.json`.

### 9. CLI, Auto Hot-Reload & Browser Live-Refresh (`cmd/larapak` & `pkg/paths`)
- **Auto Hot-Reload Bawaan (Default)**: Perintah `larapak serve [port]` (atau `larapak ladeni`) kini **secara otomatis memantau seluruh perubahan file** skrip `.ngpk`, template `.nview`, konfigurasi `.env`, CSS, dan JS. Setiap kali berkas disimpan, server me-reload router dan template engine dalam memori (< 10ms) **tanpa memutus port socket TCP** dan tanpa perlu argumen `--watch` tambahan!
- **Browser Live-Reload Otomatis (SSE)**: Dilengkapi koneksi Server-Sent Events (SSE) bawaan di mode dev; setiap kali berkas disimpan, tab browser Anda **langsung me-refresh halamannya secara otomatis** tanpa perlu menekan tombol F5!
- **Halaman Error Kustom (404 & 500)**: Otomatis mendeteksi dan me-render berkas template kustom `sawangan/errors/404.nview` dan `500.nview` jika dibuat oleh pengguna.
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
- Dual-support 24 perintah bahasa Inggris & dialek Ngapak:
  - `larapak new <nama>` / `larapak anyar <nama> [--bahasa=...]`
  - `larapak serve [port]` / `larapak ladeni [port]` / `urubna` / `mlaku` (dengan Auto Hot-Reload & Browser Refresh aktif otomatis)
  - `larapak route:list` / `larapak dalan:daptar` (Otomatis membaca rute aktif dari `web.ngpk`)
  - `larapak run <skrip.ngpk>` / `larapak mlakukna <skrip.ngpk>` / `larapak eksekusi <skrip.ngpk>`
  - `larapak make:auth` / `gawe:otentikasi` / `gawe:auth` (Scaffolding lengkap Sistem Login, Register, Auth Middleware & Views)
  - `larapak make:resource <Nama>` / `gawe:sumber` / `gawe:resource` (Scaffolding lengkap Model, Migrasi, CRUD Controller & Views)
  - `larapak make:controller <Nama>` / `gawe:pangatur` / `gawe:controller`
  - `larapak make:model <Nama>` / `gawe:wujud` / `gawe:model`
  - `larapak make:live <Nama>` / `gawe:urip` / `gawe:live` (Komponen Realtime RikatLive)
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

### 10. Mesin Interpreter Native NgapakIn (`.ngpk`) & Sawangan (`.nview`) (`pkg/ngapak`)
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
  - `res.redirect("/tujuan")` / `res.alurna("/tujuan")` untuk redirect HTTP.
  - `req.input("field")` / `req.isi("field")` dan `req.all()` / `req.kabeh()` untuk membaca data form & query.
  - `req.params["id"]`, `req.query("cari")`, `req.method`.
  - **Unggah Berkas**: `file = req.file("foto")` atau `req.berkas("foto")` lalu `file.save("storage/uploads/" + file.name)`.
  - **Akses Database Fluent & Pagination**: `data = db.tabel("produk").pilih("id", "jeneng", "reva").neng("status", "aktif").kaca(1, 10)`.
  - **Keamanan & Hashing**: `sandi.gawe("rahasia")` dan `sandi.cocok("rahasia", hash)` untuk modul otentikasi native.
  - **Sesi & Flash Message**: `sesi.simpen("user", 1)`, `sesi.jupuk("user")`, dan `sesi.kilat("pesan", "Kasil!")`.
  - **Klien HTTP Standar Native (`http.get` & `http.post`)**: Konsumsi API pihak ketiga langsung dari `.ngpk`:
    ```python
    res = http.get("https://api.github.com/zen")
    nek res.ok ya
        tulis("Status: " + res.status + " Body: " + res.body)
    rampung
    ```
  - **Rate Limiting & Anti-Brute Force (`throttle` / `wates`)**:
    ```python
    # Batasi 5 request per 60 detik (HTTP 429 Too Many Requests jika berlebih)
    rute.post("/masuk", "AuthController@login").wates(5, 60)
    ```
  - **Mailer Native SMTP Go (`email.kirim` / `mail.send`)**:
    ```python
    email.kirim({
        "marang": "pelanggan@example.com",
        "subyek": "Sugeng Rawuh",
        "sawangan": "emails/welcome",
        "data": {"jeneng": "Budi"}
    })
    ```
  - **Disk Storage & Upload Helper (`lumbung` / `storage`)**:
    - Simpan file upload otomatis dengan nama unik: `foto = req.berkas("avatar"); path = foto.simpen_neng("avatars")`
    - Operasi berkas: `storage.simpen(path, isi)`, `storage.jupuk(path)`, `storage.ana(path)`, `storage.busak(path)`, `storage.dalan(path)`.
    - Helper aset di `.nview`: `<img src="{{ asset('avatars/avatar.png') }}" />`.
  - **Helper Variabel Lingkungan (`env` / `lingkungan`)**: `env("APP_NAME", "LarapakGo")` atau `lingkungan("DB_DATABASE", "storage/db.sqlite")` untuk membaca berkas `.env`.
  - Otomatis tersinkronisasi ke antarmuka Route & API Studio di `/_studio`.

### 11. RikatLive — Reaktivitas Komponen Full-Stack Realtime (`pkg/rikatlive`)
- **100% Zero File `.go` di Proyek Pengguna**: Developer tidak perlu menyentuh atau menulis kode Go sama sekali di dalam proyek Larapak. Logika komponen ditulis murni dengan skrip NgapakIn (`.ngpk`) di folder `jeroan/Urip/` (atau `app/Live/`) dan tampilan template di `sawangan/live/` (`.nview`).
- **CLI Generator Cepat**: Jalankan `larapak gawe:urip Counter` atau `larapak make:live Counter` untuk men-generate file logika `.ngpk` dan template `.nview` secara otomatis.
- **Under-the-Hood Engine**: Mesin RikatLive di dalam binary Go menangani eksekusi AST NgapakIn, two-way state binding, snapshot kriptografis HMAC-SHA256 anti-tampering, dan routing `POST /_rikat/live` secara transparan.
- **Direktif Template Sederhana**: Cukup pasang `@rikatlive("NamaKomponen")` atau alias praktis `@live("NamaKomponen")` di berkas template `.nview`.
- **Rikat.js DOM Morphing Runtime**:
  - `r-click="method(arg)"`: Trigger aksi server secara asinkron tanpa reload halaman.
  - `r-model="field"`: Sinkronisasi dua arah (*two-way binding*) data input dengan server secara otomatis.
  - `r-submit="method"`: Intersepsi form submission dengan AJAX realtime.
  - `r-loading`: Menampilkan status/indikator loading secara otomatis saat request sedang diproses.
  - `r-poll="2s:method"`: Pembaruan data berkala (*background polling*) tanpa reload.
  - Algoritma DOM Morphing cerdas mempertahankan fokus kursor pengguna, status seleksi, dan posisi scroll secara mulus.

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

## Struktur Repositori

```
LarapakGo/
├── go.mod                     # Modul Go
├── README.md                  # Dokumentasi Utama
├── SECURITY.md                # Kebijakan Keamanan & Laporan Kerentanan
├── DEPLOYMENT.md              # Panduan Deployment Produksi (VPS, Docker, Cloud)
├── CONTRIBUTING.md            # Panduan Kontribusi & Standar Kode
├── CODE_OF_CONDUCT.md         # Kode Etik Komunitas
├── CHANGELOG.md               # Catatan Perubahan & Riwayat Rilis SemVer
├── LICENSE                    # Lisensi Open-Source MIT
├── .gitignore                 # Filter git
├── install.sh                 # Skrip instalasi satu baris untuk Linux & macOS (Bash)
├── installer/                 # Skrip pembuatan paket installer Windows (Inno Setup)
│   ├── setup.iss
│   └── app.ico
├── larapak.exe                # Binary mandiri All-in-One (LarapakGo + Mesin NgapakIn)
├── dist/                      # Rilis binary mandiri lintas platform & ekstensi
│   ├── LarapakGo-Setup-v1.0.0.exe
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
│   └── larapak/               # Binary CLI All-in-One Framework & Mesin NgapakIn
│       └── main.go
└── pkg/
    ├── larapak/               # Application, Context, Router, Middleware
    ├── ngapak/                # Interpreter Native NgapakIn (.ngpk) & Runtime Bridge
    ├── view/                  # Rikat/Sawangan Go Template Engine (.nview, @layout, @live)
    ├── database/              # Fluent Query Builder, Multi-DB Manager & Dialects (SQLite, MySQL, Postgres, MSSQL)
    ├── rikatlive/             # RikatLive Reactive Component Engine, HMAC State Protection & Dispatcher
    ├── redis/                 # Native Redis Client (RESP protocol 100% Go StdLib)
    ├── cache/                 # Cache Engine (Memory & Redis auto-driver)
    ├── queue/                 # Distributed Queue & BRPOP Worker Engine
    ├── security/              # CSRF Protection & PBKDF2 Password Hashing
    ├── session/               # Encrypted Sessions & Flash Messages
    ├── validation/            # Input & Form Validation Engine
    ├── paths/                 # Scaffolding & Multi-Language Path Resolver
    └── studio/                # Route & API Studio UI & OpenAPI 3.0 Generator
```

---

## Lisensi
Lisensi MIT — Dibuat dengan bangga oleh [Ilham Hatta Manggala](https://ilhamhattamanggala.my.id).
