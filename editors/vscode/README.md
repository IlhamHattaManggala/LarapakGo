# Larapak & NgapakIn Language Support for VS Code

Ekstensi resmi Visual Studio Code untuk ekosistem **LarapakGo** dan bahasa skrip **NgapakIn**.

## Fitur Utama

- **Logo Berkas Asli (File Icons)**:
  - Berkas `.ngpk` (skrip NgapakIn) dan `.nview` (template Sawangan) kini otomatis memiliki logo resmi Larapak (`favicon.png`) di VS Code Explorer tree.
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

### 1. Pasang Ekstensi VSIX
Di terminal jalankan:
```bash
code --install-extension larapak-vscode-1.2.2.vsix
```
*Atau di VS Code: Buka tab **Extensions (Ctrl+Shift+X)** -> Klik ikon titik tiga `...` di kanan atas -> Pilih **Install from VSIX...** -> Pilih berkas `larapak-vscode-1.2.2.vsix`.*

### 2. Menampilkan Logo Berkas di Explorer

#### Opsi A: Menggunakan Tema Bawaan Ekstensi (Larapak File Icons)
1. Buka Command Palette di VS Code (`Ctrl+Shift+P`).
2. Ketik dan pilih **Preferences: File Icon Theme**.
3. Pilih **Larapak File Icons**.
Berkas `.ngpk` dan `.nview` akan langsung menampilkan logo asli Larapak!

#### Opsi B: Jika Menggunakan Ekstensi "vscode-icons"
Jika Anda saat ini menggunakan ekstensi `vscode-icons`, tambahkan asosiasi berikut ke `settings.json` VS Code Anda:
```json
"vsicons.customIconFolderPath": "C:/Users/<Username>/.vscode/vsicons_custom_icons",
"vsicons.associations.files": [
    { "icon": "ngpk", "extensions": ["ngpk"], "format": "svg" },
    { "icon": "nview", "extensions": ["nview"], "format": "svg" }
]
```
Lalu tekan `Ctrl+Shift+P` -> jalankan **vscode-icons: Apply Icons Customization**.

#### Opsi C: Jika Menggunakan Ekstensi "Material Icon Theme"
Tambahkan baris ini ke `settings.json` VS Code Anda:
```json
"material-icon-theme.files.associations": {
    "*.ngpk": "go",
    "*.nview": "blade"
}
```

## Lisensi
MIT License - Hak Cipta (c) 2026 Ilham Hatta Manggala.
