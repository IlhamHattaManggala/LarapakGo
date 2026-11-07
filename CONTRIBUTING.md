# Panduan Kontribusi LarapakGo & NgapakIn

Terima kasih atas minat Anda untuk berkontribusi pada pengembangan **LarapakGo** dan bahasa **NgapakIn**! Proyek ini terbuka bagi siapa saja yang ingin membantu meningkatkan performa, memperbaiki bug, menambah fitur baru, maupun menyempurnakan dokumentasi.

---

## Prinsip Utama Proyek

Sebelum mulai menulis kode, harap pahami prinsip arsitektur yang menjadi landasan proyek ini:

1. **100% Go Standard Library (Zero Third-Party Dependencies)**:
   - LarapakGo dibangun murni menggunakan pustaka standar Go (`net/http`, `crypto`, `database/sql`, `html/template`, `text/scanner`, dll).
   - Kami **tidak menerima** penambahan ketergantungan paket eksternal pihak ketiga di `go.mod` kecuali ada persetujuan eksplisit dari pemelihara utama demi menjaga kemandirian dan keamanan rantai pasok binary.
2. **Kualitas Tinggi & Kode Bersih**:
   - Tulis kode yang mudah dibaca, modular, dan dapat diuji (*testable*).
   - Utamakan kesederhanaan (*Keep It Simple, Stupid - KISS*) dan hindari duplikasi (*Don't Repeat Yourself - DRY*).
3. **Anti-AI Slop**:
   - Dilarang menambahkan dekorasi emoticon/emoji acak pada antarmuka CLI, log terminal, atau template sistem. Gunakan output teknis yang lugas, terstruktur, dan profesional.

---

## Alur Kerja Kontribusi (Workflow)

### 1. Membuat Issue Diskusi
Sebelum melakukan perubahan kode berskala menengah atau besar:
- Buka issue baru di repositori GitHub untuk mendiskusikan usulan fitur atau melaporkan bug.
- Sertakan konteks masalah, perilaku yang diharapkan, serta usulan solusi rancangan.

### 2. Fork dan Clone Repositori
Fork repositori ini ke akun GitHub Anda, lalu clone ke komputer lokal:

```bash
git clone https://github.com/IlhamHattaManggala/LarapakGo.git
cd LarapakGo
```

### 3. Struktur Percabangan (Branching Strategy)
- **`main`**: Branch stabil untuk rilis publik dan kompilasi installer resmi.
- **`development`**: Branch integrasi aktif untuk fitur dan perbaikan yang sedang dikembangkan.

Silakan buat branch fitur atau perbaikan baru dari branch `development`:

```bash
git checkout development
git checkout -b feature/nama-fitur-anda
# Atau untuk perbaikan bug:
git checkout -b fix/nama-perbaikan
```

### 4. Menjalankan Pengujian Lokal (Unit Testing)
Seluruh paket di LarapakGo memiliki unit test mandiri. Sebelum mengirimkan perubahan, pastikan semua tes lulus tanpa kesalahan:

```bash
# Menjalankan seluruh test suite
go test -v ./...

# Menjalankan pengujian spesifik interpreter NgapakIn
go test -v ./pkg/ngapak -run TestVM

# Menjalankan pengujian router dan context HTTP
go test -v ./pkg/larapak
```

Pastikan pula kode lolos pemeriksaan format dan analisis statis standar Go:

```bash
# Format kode sesuai standar idiomatis Go
go fmt ./...

# Analisis potensi bug dan kesalahan umum
go vet ./...
```

### 5. Komitmen dan Pengiriman Pull Request (PR)
- Gunakan format pesan commit konvensional (*Conventional Commits*), contoh:
  - `feat: Tambah helper rute dinamis di Trie router`
  - `fix: Perbaiki penanganan syntax error pada parser NgapakIn`
  - `docs: Lengkapi contoh penggunaan PBKDF2 di modul security`
- Push branch Anda ke repositori fork:
  ```bash
  git push origin feature/nama-fitur-anda
  ```
- Buka **Pull Request (PR)** baru dengan target branch **`development`** di repositori utama.
- Berikan deskripsi yang jelas mengenai perubahan apa saja yang dibuat dan issue terkait yang diselesaikan.

---

## Standar Kode & Konvensi

1. **Penamaan (Naming Conventions)**:
   - Gunakan nama yang deskriptif dan mencerminkan tujuannya secara eksplisit (hindari singkatan yang membingungkan seperti `tmp`, `val1`, `abc`).
   - Ekspor fungsi/metode/struct dengan huruf kapital (*PascalCase*), sedangkan identitas privat menggunakan *camelCase*.
2. **Penanganan Error (Error Handling)**:
   - Periksa dan tangani error secara eksplisit (`if err != nil`).
   - Hindari mengabaikan error secara diam-diam (*silent failure*).
   - Berikan konteks error yang jelas saat membungkus error (`fmt.Errorf("gagal memproses berkas %s: %w", path, err)`).
3. **Dokumentasi & Komentar**:
   - Berikan komentar dokumen (*docstring*) pada setiap fungsi, metode, tipe, dan konstanta publik.
   - Komentar harus menjelaskan **alasan (WHY)** suatu logika diterapkan, bukan sekadar mengulang kode (WHAT).
