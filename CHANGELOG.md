# Catatan Perubahan (Changelog)

Semua perubahan besar pada proyek **LarapakGo** dan mesin bahasa **NgapakIn (Go Edition)** dicatat secara kronologis dalam dokumen ini.

Format pencatatan mengacu pada panduan [Keep a Changelog](https://keepachangelog.com/id/1.0.0/) dan menganut prinsip [Semantic Versioning (SemVer)](https://semver.org/).

---

## [1.0.0] - 2026-11-07

### Rilis Perdana Resmi (Initial Production Release)

Versi 1.0.0 menandai peluncuran perdana arsitektur web framework Larapak yang ditulis murni dalam bahasa pemrograman **Go (Golang)** dengan prinsip **100% Go Standard Library (Zero External Dependencies)**.

### Ditambahkan (Added)
- **All-in-One Multi-Call Binary (`larapak` & `ngapakin`)**:
  - Binary tunggal mandiri tanpa dependensi runtime pihak ketiga.
  - Berfungsi ganda: sebagai CLI web framework (`larapak`) maupun mesin eksekusi skrip / REPL bahasa (`ngapakin`).
  - Mendukung eksekusi skrip langsung via `larapak run skrip.ngpk` atau `ngapakin skrip.ngpk`.
  - REPL terminal interaktif berbasis buffer standar via `larapak repl` atau `ngapakin`.

- **Mesin Interpreter Native NgapakIn (`pkg/ngapak`)**:
  - Implementasi mandiri Lexer, Parser, AST, dan Tree-Walk Evaluator 100% Go Standard Library tanpa ketergantungan Python atau CGo.
  - Mendukung sintaksis penuh dialek Banyumasan/Ngapak:
    - Kontrol alur: `nek`, `ya`, `liyane`, `rampung`, `baleni`, `mandeg`, `terusna`.
    - Fungsi dan pemanggilan: `gawe`, `balekna`.
    - Operator logika: `lan`, `utawa`, `ora`, `neng`.
    - Tipe data: Angka (int/float), Teks, List/Array, Map/Dictionary, Boolean (`bener`/`salah`), dan Null (`kosong`).
  - Integrasi controller native: memanggil handler controller berekstensi `.ngpk` (`HomeController@index`) langsung melalui HTTP request context.

- **Mesin Template Sawangan Native (`pkg/view`)**:
  - Pewarisan tata letak dengan direktif `@layout`, `@section`, `@endsection`, dan `@yield`.
  - Sanitasi otomatis XSS untuk interpolasi variabel `{{ var }}` dan raw printing `{!! raw !!}`.
  - Direktif keamanan `@csrf` yang otomatis menyuntikkan token tersembunyi ke dalam formulir.
  - Evaluasi blok kondisional dan loop iterasi langsung di sisi server.

- **HTTP Core Engine & Parameterized Router (`pkg/larapak`)**:
  - Trie Parameterized Router berkecepatan tinggi dengan dukungan parameter dinamis (`:id`, `*filepath`).
  - Pengelompokan rute (`Group`) dan perantaian middleware (*middleware chaining*).
  - Context API ergonomis: `c.JSON()`, `c.HTML()`, `c.String()`, `c.BindJSON()`, `c.Param()`, `c.Query()`, `c.Header()`.
  - Middleware inti bawaan: Request Logger presisi nanodetik, Panic Recovery otomatis, dan CORS handler.

- **Route & API Studio GUI Terintegrasi (`pkg/studio`)**:
  - Halaman pengujian API interaktif di `/_studio` dengan palet warna modern gelap dan ikon SVG teknis murni (bebas emoticon).
  - Pengujian live HTTP request (URL params, query params, headers, JSON body).
  - Ekspor otomatis ke spesifikasi standar **OpenAPI 3.0.3** JSON di `/_studio/openapi.json`.

- **Modul Keamanan & Kriptografi Standar (`pkg/security`)**:
  - Proteksi otomatis CSRF dengan token kriptografis 32-byte acak dan validasi *constant-time* via `crypto/subtle`.
  - Password Hasher aman berbasis **PBKDF2-SHA256** (RFC 2898) dengan salt 16-byte dan 100.000 iterasi.
  - Helper verifikasi `VerifyPassword` dan deteksi rotasi keamanan `NeedsRehash`.

- **Manajemen Sesi Terenkripsi (`pkg/session`)**:
  - Signed Session Cookies berbasis **HMAC-SHA256** untuk menjamin integritas sesi terhadap manipulasi klien.
  - Fitur pesan kilat (*flash messages*) satu kali baca untuk respon redirect formulir.

- **Fluent SQL Query Builder & Schema Migration (`pkg/database`)**:
  - Fluent Query Builder berbasis prepared statement aman anti-SQL Injection.
  - Blueprint perancangan skema tabel database (`ID`, `String`, `Integer`, `Text`, `Boolean`, `Timestamps`).
  - CLI migrasi database: `migrate`, `migrate:fresh`, `migrate:status`, `rollback`, dan `db:seed`.

- **Mesin Validasi Formulir Input (`pkg/validation`) & Direktif `@error`**:
  - Aturan validasi lengkap: `required`, `email`, `min`, `max`, `numeric`, `alpha`, `in`, `confirmed`.
  - Jembatan validasi langsung di `.ngpk` via `req.validate(...)` / `req.validasi(...)`.
  - Direktif template `@error("field") ... @enderror` dengan variabel pesan otomatis `{{ message }}` / `{{ pesan }}`.
  - Sesi penyimpanan input lawas (*Old Input Flashing*) via `res.with_input().with_errors(v.errors()).back()` / `res.karo_input().karo_kliru(...).balek()`.
  - Akses input lawas di template `{{ old.field }}` / `{{ lawas.field }}` dan di controller `req.old("field")` / `req.lawas("field")`.

- **Klien HTTP Standar Native & Helper Lingkungan di `.ngpk`**:
  - Klien HTTP berbasis Go Standard Library (`http.get(url, [headers])`, `http.post(url, [data], [headers])`) mengembalikan status, body, dan parsed JSON.
  - Helper variabel lingkungan `env(key, [default])` dan alias `lingkungan(...)` langsung dari skrip NgapakIn.

- **Full Resource CRUD Generator (`make:resource` / `gawe:sumber`)**:
  - Perintah satu langkah `larapak make:resource <Nama>` atau `larapak gawe:sumber <Nama>`.
  - Otomatis men-generate Model (`.ngpk`), Migrasi tabel (`.sql`), CRUD Controller (`.ngpk` mencakup index, create, store, show, edit, update, delete), dan Template Views lengkap (`index.nview`, `create.nview`, `edit.nview`).

- **Scaffolding Otentikasi Lengkap (`make:auth` / `gawe:otentikasi`)**:
  - Perintah satu langkah `larapak make:auth` atau `larapak gawe:otentikasi`.
  - Otomatis men-generate Model `User.ngpk`, Migrasi `users`, Middleware `AuthMiddleware.ngpk`, Controller `AuthController.ngpk` (Login, Register, Logout, Dashboard), dan Template Views lengkap (`login.nview`, `register.nview`, `dashboard.nview`).

- **Rate Limiting & Anti-Brute Force (`throttle` / `wates`)**:
  - Proteksi rute dari serangan spam dan brute force dengan sliding window token bucket per IP.
  - Dukungan perataan rute `rute.wates(max, detik)` atau chaining per rute `rute.post(...).wates(max, detik)`.
  - Mengembalikan respons HTTP 429 Too Many Requests dan header `Retry-After`.

- **Database Inspector Visual di Route & API Studio (`/_studio`)**:
  - Tab navigasi inspektur database untuk meninjau daftar tabel, struktur kolom, dan isi rekaman baris data langsung dari browser pada lingkungan pengembangan lokal dengan 100% SVG icons.

- **Mailer Native SMTP Go (`pkg/mail`)**:
  - Pengiriman email transaksional langsung dari skrip `.ngpk` via `email.kirim(...)` / `mail.send(...)` menggunakan paket standar `net/smtp`.
  - Integrasi template email dengan `sawangan` dan fallback pencatatan log otomatis di mode pengembangan lokal.

- **Disk Storage & Asset Helper (`pkg/storage`)**:
  - Penyimpanan berkas upload otomatis dengan nama acak unik: `foto.simpen_neng("folder")` / `foto.save_to("folder")`.
  - Operasi berkas storage: `storage.simpen`, `storage.jupuk`, `storage.ana`, `storage.busak`, `storage.dalan`.
  - Helper template `{{ asset("path") }}` / `{{ aset("path") }}` dan evaluasi fallback interpolasi `{{ old.field || item.field }}`.

- **Asosiasi Relasi Database Model (`duwe_akeh`, `has_many`, `melu`, `belongs_to`, `duwe_siji`, `has_one`)**:
  - Relasi `has_many` (`duwe_akeh`) untuk relasi satu-ke-banyak antar model dan tabel.
  - Relasi `belongs_to` (`melu`) untuk relasi child ke induk dengan deteksi foreign key otomatis.
  - Relasi `has_one` (`duwe_siji`) untuk relasi satu-ke-satu.
  - Integrasi fluent chaining langsung pada record: `user.duwe_akeh("artikels").neng("status", "publikasi").jupuk()`.
  - Dukungan pemanggilan global via `db.duwe_akeh(user, "artikels")` dan `db.melu(artikel, "users")`.

- **Mesin Faker & Generator Model Factory (`pkg/factory`, `make:factory` / `gawe:pabrik`)**:
  - Generator data palsu 100% Go Standard Library berbasis `crypto/rand` tanpa dependensi eksternal.
  - Objek global `pabrik` dan `factory`: `jeneng()` / `name()`, `email()`, `hp()` / `phone()`, `alamat()` / `address()`, `kutha()` / `city()`, `teks()` / `text()`, `angka()` / `number()`, `rega()` / `price()`, `pilihan()` / `choice()`, `tanggal()` / `date()`, `uuid()`.
  - Batch data inserter otomatis via `pabrik.gawe("tabel", jumlah, callback)`.
  - CLI `larapak make:factory <Nama>` atau `larapak gawe:pabrik <Nama>`.
  - Eksekusi database seeder native via `larapak db:seed` / `larapak bibiti` yang membaca dan mengeksekusi berkas `.ngpk` di direktori seeders.

- **Middleware Response Caching (`rute.cache` / `rute.simpen_cache`)**:
  - Caching in-memory berkecepatan tinggi untuk rute GET dengan durasi TTL yang dapat dikonfigurasi.
  - Dukungan perataan rute `rute.cache(detik)` dan chaining per rute `rute.get(...).cache(detik)`.
  - Header respons diagnostik otomatis `X-Larapak-Cache: HIT` dan `X-Larapak-Cache: MISS`.

- **JSON API Resource Transformer (`make:resource-api` / `gawe:sumber-api`, `res.resource`)**:
  - Standardisasi respons API JSON sesuai format JSON:API modern dengan wrapper `data` dan `meta`.
  - Perintah generator CLI `larapak make:resource-api <Nama>` / `larapak gawe:sumber-api <Nama>`.
  - Helper controller `res.resource(data, "NamaResource")` / `res.sumber(...)` atau menggunakan closure `gawe(item) ... rampung`.
  - Mendukung transformasi rekaman tunggal (single record), daftar rekaman (list), maupun hasil paginasi (`.kaca()` / `.paginate()`) dengan metadata otomatis.

- **Dual-Language CLI & Multi-Language Directory Scaffolding**:
  - Mendukung 22 pasangan perintah dalam bahasa Inggris dan dialek Ngapak Tegalan (`anyar`, `ladeni`, `dalan:daptar`, `gawe:pangatur`, `munggah`, `mudhun`, `resik:cache`, dll).
  - Dukungan pembuatan struktur direktori multi-bahasa (`--bahasa=ngapak|indonesia|english`).
  - Pembuatan proyek baru yang bersih 100% murni berkas `.ngpk` dan `.nview` tanpa berkas engine `.go` atau `go.mod`.

- **Frontend In-House Rikat CSS & Rikat UI**:
  - Bundling aset statis bawaan `rikat.css` (utilitas layout responsif) dan `rikat.js` (komponen reaktif).

- **Dukungan Editor Visual Studio Code**:
  - Ekstensi resmi `larapak-vscode-1.0.0.vsix` dengan pewarnaan sintaks TextMate `.ngpk` dan `.nview`.
  - Snippets produktivitas pengembangan web.
  - Tema ikon berkas kustom dengan logo resmi Larapak.

- **Paket Distribusi & Pemasang Resmi**:
  - Installer Windows resmi berbasis Inno Setup (`dist/LarapakGo-Setup-v1.0.0.exe`) dengan auto-PATH registration dan uninstaller.
  - Skrip pemasang otomatis Unix (`install.sh`) untuk Linux & macOS.
  - Binary portabel mandiri: `larapak-v1.0.0-windows-amd64.exe`, `larapak-v1.0.0-linux-amd64`, dan `larapak-v1.0.0-darwin-arm64`.

