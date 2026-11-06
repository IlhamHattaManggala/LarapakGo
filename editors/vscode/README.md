# Larapak & NgapakIn Language Support for VS Code

Ekstensi resmi Visual Studio Code untuk ekosistem **LarapakGo** dan bahasa skrip **NgapakIn**.

## Fitur Utama

- **NgapakIn Script (`.ngpk`)**:
  - Pewarnaan sintaks (*Syntax Highlighting*) untuk semua keyword (`nek`, `ya`, `liyane`, `rampung`, `baleni`, `saka`, `nganti`, `gawe`, `balekna`, `kelas`, `extends`, `mandeg`, `terusna`).
  - Penyorotan konstanta (`bener`, `salah`, `kosong`) dan fungsi bawaan (`tulis`, `panjang`).
  - Dukungan perutean HTTP web (`rute.get`, `rute.post`, `req`, `res`).
- **Sawangan Template (`.nview`)**:
  - Pewarnaan direktif template: `@layout(...)`, `@section(...)`, `@endsection`, `@yield(...)`, `@csrf`, `@include(...)`.
  - Penyorotan interpolasi variabel aman `{{ var }}` dan HTML mentah `{!! rawHtml !!}`.
  - Integrasi lengkap dengan elemen tag HTML.
- **Koleksi Snippets Produktif**:
  - `rute:get`, `rute:post`, `gawe`, `nek`, `nek:liyane`, `baleni:range`, `baleni:neng`
  - `layout`, `section`, `yield`, `csrf`, `var`, `raw`

## Cara Pasang di VS Code

### Cara 1: Menggunakan Berkas VSIX
1. Unduh atau gunakan berkas `larapak-vscode-1.2.2.vsix`.
2. Di terminal jalankan:
   ```bash
   code --install-extension larapak-vscode-1.2.2.vsix
   ```
   Atau di VS Code: Buka tab **Extensions (Ctrl+Shift+X)** -> Klik ikon titik tiga `...` di kanan atas -> Pilih **Install from VSIX...** -> Pilih berkas `larapak-vscode-1.2.2.vsix`.

### Cara 2: Tautan Manual ke Direktori Ekstensi
Salin folder `editors/vscode` ke direktori ekstensi VS Code Anda:
- **Windows**: `%USERPROFILE%\.vscode\extensions\larapak-vscode`
- **Linux / macOS**: `~/.vscode/extensions/larapak-vscode`

## Lisensi
MIT License - Hak Cipta (c) 2026 Ilham Hatta Manggala.
