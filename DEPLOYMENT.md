# Panduan Deployment LarapakGo ke Lingkungan Produksi

Dokumen ini menjelaskan langkah-langkah komprehensif untuk melakukan deployment aplikasi web berbasis **LarapakGo** dan mesin bahasa **NgapakIn** ke server produksi (VPS Linux, Docker Container, maupun platform Cloud PaaS).

---

## Karakteristik Binary Produksi LarapakGo

Berbeda dengan framework berbasis bahasa interpreted (seperti Python atau Node.js) yang membutuhkan instalasi runtime, virtualenv, dan process manager eksternal, LarapakGo dikompilasi menjadi **single standalone executable binary**:

- **Zero External Runtime**: Server produksi tidak membutuhkan instalasi Go compiler, Python, atau paket eksternal apapun.
- **Efisiensi Sumber Daya**: Konsumsi memori RAM sangat rendah (~10 sampai 25 MB pada beban normal).
- **Latency Sangat Rendah**: Latency respon HTTP sub-milidetik (< 0.5 ms) didukung Trie router dan standard library net/http bawaan Go.
- **Portabel**: Tersedia binary siap pakai untuk Linux (`x86_64` & `arm64`), macOS (`Apple Silicon & Intel`), dan Windows.

---

## Ringkasan Opsi Deployment

| Metode | Deskripsi | Rekomendasi Target |
|---|---|---|
| **Systemd Service + Nginx** | Binary standalone berjalan sebagai daemon latar belakang di balik Nginx reverse proxy | VPS Linux (Ubuntu, Debian, AlmaLinux, RHEL) |
| **Docker & Docker Compose** | Terisolasi dalam container OCI ringan berbasis Docker Compose | Server berbasis Container / Kubernetes |
| **Cloud PaaS (Railway / Render / Fly.io)** | Deployment otomatis berbasis Git dengan deteksi port dinamis | Cloud Serverless / PaaS |

---

## Metode 1: VPS Linux Menggunakan Systemd & Nginx (Standar Industri)

### 1. Menyiapkan Direktori & Binary Aplikasi di Server

Di server VPS Anda (misalnya Ubuntu 22.04 / 24.04 LTS):

```bash
# Buat pengguna sistem khusus aplikasi
sudo useradd -r -s /bin/false larapak

# Siapkan direktori aplikasi
sudo mkdir -p /var/www/my-app
cd /var/www/my-app

# Unduh binary rilis Linux amd64
sudo curl -fsSL -o /usr/local/bin/larapak \
  https://github.com/IlhamHattaManggala/LarapakGo/releases/download/v1.0.0/larapak-v1.0.0-linux-amd64

# Beri izin eksekusi
sudo chmod +x /usr/local/bin/larapak
```

Salin berkas proyek Anda (`routes/`, `sawangan/` atau `views/`, `app/`, `public/`, `.env`) ke `/var/www/my-app`, lalu atur kepemilikan izin berkas:

```bash
sudo chown -R larapak:larapak /var/www/my-app
sudo chmod -R 755 /var/www/my-app
sudo chmod -R 775 /var/www/my-app/storage
```

### 2. Konfigurasi Lingkungan (`.env`)

Pastikan berkas `/var/www/my-app/.env` disesuaikan untuk mode produksi:

```ini
APP_NAME=MyProductionApp
APP_ENV=production
APP_DEBUG=false
APP_PORT=8080
APP_KEY=secret_kriptografis_acak_32_karakter_atau_lebih_2026
APP_LANG=english

DB_CONNECTION=sqlite
DB_DATABASE=storage/database.sqlite
```

> **Catatan Keamanan**: Selalu ubah `APP_DEBUG=false` di lingkungan produksi untuk mencegah kebocoran informasi jejak tumpukan (*stack trace*) kepada pengguna luar.

### 3. Membuat Systemd Service Daemon

Buat berkas unit systemd:

```bash
sudo nano /etc/systemd/system/larapak.service
```

Masukkan konfigurasi berikut:

```ini
[Unit]
Description=LarapakGo Production Web Service
After=network.target

[Service]
Type=simple
User=larapak
Group=larapak
WorkingDirectory=/var/www/my-app
ExecStart=/usr/local/bin/larapak serve 8080
Restart=always
RestartSec=5s
EnvironmentFile=/var/www/my-app/.env

# Pembatasan Keamanan Linux
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=full

[Install]
WantedBy=multi-user.target
```

Muat ulang dan aktifkan service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable larapak
sudo systemctl start larapak

# Periksa status
sudo systemctl status larapak
```

Untuk memantau log aplikasi secara langsung:
```bash
journalctl -u larapak -f
```

---

## Metode 2: Nginx sebagai Reverse Proxy & Static Asset Caching

Nginx bertindak sebagai gerbang terdepan (*reverse proxy*), mengelola penghentian enkripsi SSL/TLS, serta melayani aset berkas statis (`/css/`, `/js/`, `/images/`) langsung dari disk tanpa membebani aplikasi.

### 1. Konfigurasi Nginx Server Block

Pasang Nginx jika belum terpasang:
```bash
sudo apt update && sudo apt install nginx -y
```

Buat konfigurasi situs:
```bash
sudo nano /etc/nginx/sites-available/larapak
```

Isi konfigurasi berikut:

```nginx
server {
    listen 80;
    server_name domainanda.com www.domainanda.com;

    # Batas ukuran payload unggahan
    client_max_body_size 25M;

    # Melayani aset statis publik secara langsung dengan caching agresif
    location /css/ {
        alias /var/www/my-app/public/css/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }

    location /js/ {
        alias /var/www/my-app/public/js/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }

    location /images/ {
        alias /var/www/my-app/public/images/;
        expires 30d;
        add_header Cache-Control "public, no-transform";
    }

    location = /favicon.ico {
        alias /var/www/my-app/public/favicon.ico;
        expires 30d;
    }

    # Teruskan request aplikasi dinamis ke LarapakGo
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Pengaturan buffer
        proxy_buffering on;
        proxy_buffer_size 128k;
        proxy_buffers 4 256k;
    }
}
```

Aktifkan konfigurasi dan periksa sintaks:
```bash
sudo ln -s /etc/nginx/sites-available/larapak /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 2. Memasang Sertifikat SSL Gratis (Let's Encrypt / Certbot)

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d domainanda.com -d www.domainanda.com
```

Certbot akan memperbarui sertifikat HTTPS secara otomatis saat masa berlakunya mendekati 30 hari.

---

## Metode 3: Deployment Menggunakan Docker & Docker Compose

Setiap proyek yang dibuat dengan `larapak anyar` / `larapak new` sudah menyertakan berkas `docker-compose.yml` bawaan:

```yaml
services:
  app:
    image: ilhamhattamanggala/larapak:latest
    container_name: my-app
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - .:/app
    working_dir: /app
    command: ["larapak", "serve"]
    environment:
      - APP_ENV=production
      - APP_DEBUG=false
      - APP_PORT=8080
      - APP_LANG=english
```

### Menjalankan Container Produksi:
```bash
# Jalankan container di latar belakang
docker compose up -d

# Periksa status
docker compose ps

# Memantau log real-time
docker compose logs -f
```

---

## Metode 4: Deployment ke Cloud PaaS (Railway / Render / Fly.io)

LarapakGo mendukung variabel lingkungan standar cloud seperti `PORT` dan `APP_PORT`.

### Konfigurasi Railway / Render:
1. Hubungkan repositori Git Anda di dashboard penyedia Cloud.
2. Tambahkan variabel lingkungan:
   - `APP_ENV`: `production`
   - `APP_DEBUG`: `false`
   - `APP_KEY`: `[kunci-rahasia-acak]`
3. Perintah Start (*Start Command*):
   ```bash
   larapak serve
   ```
   *(LarapakGo otomatis mendeteksi variabel lingkungan `$PORT` yang disediakan oleh platform Cloud).*

---

## Checklist Pengerasan Keamanan Produksi (Production Hardening)

Sebelum mempublikasikan aplikasi ke publik, verifikasi checklist berikut:

- [ ] `APP_DEBUG=false` sudah disetel di berkas `.env`.
- [ ] `APP_KEY` telah diubah dari nilai default menjadi string acak kriptografis unik.
- [ ] Hak akses berkas `storage/` dan `database.sqlite` hanya dapat dibaca/tulis oleh user pengelola daemon (`chmod 775`).
- [ ] Firewall server (UFW) hanya membuka port yang diperlukan:
  ```bash
  sudo ufw allow OpenSSH
  sudo ufw allow 'Nginx Full'
  sudo ufw enable
  ```
- [ ] SSL/TLS sudah aktif (akses HTTP otomatis diarahkan ke HTTPS).
- [ ] Token `@csrf` dipasang pada seluruh formulir interaktif di template Sawangan (`.nview`).

