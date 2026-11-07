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

- **Mesin Validasi Formulir Input (`pkg/validation`)**:
  - Aturan validasi lengkap: `required`, `email`, `min`, `max`, `numeric`, `alpha`, `in`, `confirmed`.

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

