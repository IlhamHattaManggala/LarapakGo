# Kebijakan Keamanan (Security Policy)

Keamanan aplikasi, data, dan infrastruktur pengguna merupakan prioritas utama dalam perancangan serta pemeliharaan **LarapakGo** dan mesin bahasa **NgapakIn**. Dokumen ini menjelaskan kebijakan keamanan, fitur perlindungan bawaan, serta prosedur pelaporan kerentanan secara bertanggung jawab (*responsible disclosure*).

---

## Versi yang Didukung

Pembaruan keamanan dan perbaikan *critical vulnerabilities* diprioritaskan pada versi aktif berikut:

| Versi | Status | Catatan |
|---|---|---|
| **1.0.x** | **Didukung Penuh** | Versi stabil saat ini (Main & Release) |
| < 1.0.0 | Tidak Didukung | Harap segera perbarui ke versi 1.0.0+ |

---

## Arsitektur Keamanan Bawaan (Built-in Security)

LarapakGo dirancang dari awal dengan prinsip *Security by Design* menggunakan **100% Go Standard Library** tanpa dependensi pihak ketiga (*Zero Third-Party Dependencies*):

1. **Bebas Kerentanan Rantai Pasok (*Zero Supply Chain Attacks*)**:
   - Karena tidak menggunakan pustaka eksternal pihak ketiga (seperti paket npm atau modul third-party yang rentan terhadap pembajakan dependensi), vektor serangan injeksi dependensi terminimalisir secara menyeluruh.

2. **Perlindungan Pemalsuan Permintaan Lintas Situs (CSRF Protection)**:
   - Modul `pkg/security` menyediakan token CSRF 32-byte berbasis kriptografi acak `crypto/rand`.
   - Validasi token dilakukan menggunakan perbandingan waktu-konstan (*constant-time comparison* via `subtle.ConstantTimeCompare`) untuk mencegah serangan *timing attack*.
   - Integrasi otomatis pada formulir Sawangan lewat direktif `@csrf` dan verifikasi header `X-CSRF-TOKEN` untuk request AJAX/API.

3. **Pencegahan Cross-Site Scripting (XSS Prevention)**:
   - Mesin template Sawangan (`pkg/view`) secara otomatis melakukan sanitasi dan *HTML-escaping* pada seluruh interpolasi variabel `{{ variabel }}`.
   - Konten HTML mentah hanya dapat ditampilkan melalui sintaks eksplisit `{!! raw_html !!}` yang mengharuskan pengembang memvalidasi input terlebih dahulu.

4. **Keamanan Sandi Standar Industri (Password Hashing)**:
   - Menggunakan algoritma **PBKDF2-SHA256** (standar RFC 2898) dengan salt unik kriptografis 16-byte dan iterasi adaptif tinggi (100.000 iterasi).
   - Dilengkapi fungsi `VerifyPassword` berbasis *constant-time* dan helper `NeedsRehash` untuk rotasi keamanan berkala.

5. **Manajemen Sesi Aman (*Signed Session Cookies*)**:
   - Modul `pkg/session` menandatangani cookie sesi menggunakan hash autentikasi pesan **HMAC-SHA256** dengan `APP_KEY` rahasia.
   - Mencegah manipulasi cookie (*cookie tampering*) di sisi klien.
   - Atribut cookie bawaan menggunakan konfigurasi aman (`HttpOnly`, `SameSite=Lax`, dan `Secure` pada mode HTTPS).

6. **Pencegahan SQL Injection**:
   - Query Builder pada `pkg/database` memanfaatkan *parameterized queries* (prepared statements) standar `database/sql`, memastikan nilai masukan pengguna selalu dipisahkan secara aman dari struktur perintah SQL.

7. **Keamanan Memori & Konkurensi**:
   - Dibangun di atas runtime Go yang memiliki *garbage collection* dan perlindungan memori otomatis, memitigasi risiko *buffer overflow* dan serangan manipulasi pointer level rendah.

---

## Melaporkan Kerentanan Keamanan

Jika Anda menemukan potensi kerentanan keamanan di LarapakGo, kami sangat menghargai kerja sama Anda untuk **TIDAK** mempublikasikannya melalui GitHub Issues publik atau media sosial sebelum perbaikan dirilis.

### Langkah Pelaporan:

1. Kirimkan laporan terperinci melalui email ke:
   - **Email Pengelola**: `ilhamhattamanggala31@gmail.com`
   - **Situs Resmi / Kontak**: [ilhamhattamanggala.my.id](https://ilhamhattamanggala.my.id)
2. Cantumkan informasi berikut di dalam laporan:
   - Deskripsi lengkap mengenai sifat kerentanan.
   - Komponen atau paket yang terdampak (misalnya: `pkg/security`, `pkg/larapak`, `cmd/larapak`).
   - Langkah-langkah reproduksi langkah-demi-langkah atau *Proof of Concept* (PoC).
   - Dampak potensial terhadap kerahasiaan, integritas, atau ketersediaan sistem.
   - Saran perbaikan atau mitigasi sementara jika Anda memilikinya.

### Komitmen Waktu Respons Kami:
- **Konfirmasi Penerimaan**: Maksimal dalam 1x24 jam hingga 48 jam kerja.
- **Tinjauan & Validasi**: Tim kami akan memverifikasi temuan dan menentukan tingkat keparahan (CVSS).
- **Rilis Tambalan (*Patch*)**: Perbaikan akan disiapkan di branch privat dan dirilis sesegera mungkin bersamaan dengan pengumuman rilis dan atribusi kepada penemu kerentanan.

