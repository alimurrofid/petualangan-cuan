# 🚀 Panduan Lengkap Setup Server & Deployment

## **Petualangan Cuan**

*(VPS + Docker + GitHub Actions + Nginx + HTTPS)*

Dokumen ini menjelaskan **end-to-end deployment** aplikasi **Petualangan Cuan** ke VPS menggunakan:

* Docker & Docker Compose
* GitHub Actions (CI/CD)
* GitHub Container Registry (GHCR)
* Nginx sebagai Reverse Proxy + HTTPS

---

## 📑 Daftar Isi

1. Setup Akses SSH (Laptop → Server)
2. Konfigurasi GitHub Secrets
3. Persiapan Awal Server (VPS)
4. Setup Project & Environment Variables
5. Konfigurasi Nginx (HTTPS & Reverse Proxy)
6. Deployment Otomatis (CI/CD)
7. Verifikasi & Troubleshooting

---

## 1️⃣ Setup Akses SSH (Laptop → Server)

Agar laptop dan GitHub Actions dapat mengakses server **tanpa password**, gunakan SSH Key.

---

### 1.1 Generate SSH Key di Laptop

```bash
ssh-keygen -t ed25519 -C "email_github_anda"
```

Tekan **Enter terus** sampai selesai.

File akan tersimpan di:

* **Private key**: `~/.ssh/id_ed25519`
* **Public key**: `~/.ssh/id_ed25519.pub`

---

### 1.2 Daftarkan Public Key ke Server

Copy public key:

```bash
cat ~/.ssh/id_ed25519.pub
```

Login ke server, lalu:

```bash
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys
```

Paste public key, simpan, lalu set permission:

```bash
chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

---

### 1.3 Ambil Private Key untuk GitHub Actions

```bash
cat ~/.ssh/id_ed25519
```

⚠️ **PERINGATAN**

* **JANGAN PERNAH** commit private key ke repository
* Private key hanya disimpan di **GitHub Secrets**

---

## 2️⃣ Konfigurasi GitHub Secrets

Masuk ke repository GitHub:

**Settings → Secrets and variables → Actions → New repository secret**

Tambahkan secrets berikut:

| Nama Secret       | Deskripsi                           |
| ----------------- | ----------------------------------- |
| `SERVER_IP`       | IP Public VPS                       |
| `SERVER_USER`     | User SSH VPS (`ubuntu` / `root`)    |
| `SSH_PRIVATE_KEY` | Isi file `id_ed25519` (private key) |

Secrets ini digunakan oleh workflow:

```
.github/workflows/deploy.yml
```

---

## 3️⃣ Persiapan Awal Server (VPS)

Login ke server:

```bash
ssh user@IP_SERVER
```

---

### 3.1 Install Docker & Docker Compose

```bash
sudo apt update && sudo apt upgrade -y

sudo apt install -y \
  apt-transport-https \
  ca-certificates \
  curl \
  software-properties-common

curl -fsSL https://download.docker.com/linux/ubuntu/gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

echo \
"deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] \
https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" \
| sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

Cek instalasi:

```bash
docker --version
docker compose version
```

---

### 3.2 Login ke GitHub Container Registry (GHCR)

Agar server bisa menarik image dari GitHub.

1. Buat **Personal Access Token (Classic)**
   Scope minimum:

   * `read:packages`
   * `repo`
   * `workflow`

2. Login di server:

```bash
echo "GITHUB_PAT_ANDA" | docker login ghcr.io -u USERNAME_GITHUB --password-stdin
```

---

## 4️⃣ Setup Project & Environment Variables

### 4.1 Buat Folder Project

```bash
mkdir -p ~/petualangan-cuan
cd ~/petualangan-cuan
```

---

### 4.2 Buat File `.env` (Produksi)

```bash
nano .env
```

Contoh isi:

```env
# Server
PORT=8080

# Database (PostgreSQL)
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_aman
DB_NAME=petualangan_cuan

# Security
JWT_SECRET=string_random_panjang
JWT_EXPIRY_HOURS=72

# Google OAuth (WAJIB PRODUKSI)
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxxxx
GOOGLE_REDIRECT_URL=https://petualangancuan.rofid.me/api/auth/google/callback

# Frontend
FRONTEND_URL=https://petualangancuan.rofid.me
```

⚠️ **Catatan Penting**

* Jangan gunakan `localhost` di production
* OAuth redirect **HARUS domain publik**
* Jika masih redirect ke localhost → `.env` **belum ter-load**

---

### 4.3 Buat `docker-compose.yml`

```bash
nano docker-compose.yml
```

📌 **WAJIB**

* Isi **HARUS SAMA** dengan versi di repository
* Jika file di repo berubah → **update juga di server**

---

### 4.4 Buat Folder `uploads` (Penyimpanan File)

Folder ini digunakan untuk menyimpan file upload (gambar, dokumen, dll) dan **harus persisten** saat redeploy.

#### 1️⃣ Buat Folder `uploads`

```bash
mkdir -p uploads
```

📌 Folder **wajib berada di root project**:

`~/petualangan-cuan/uploads`

---

#### 2️⃣ Cek UID & GID Container Backend

Sebelum mengatur permission, pastikan user yang digunakan di dalam container backend.

```bash
docker exec petualangan_cuan_backend id
```

Output yang **diharapkan**:

```text
uid=1000(appuser) gid=1000(appgroup)
```

Jika UID/GID **berbeda**:

* Sesuaikan `chown` dengan UID/GID yang muncul
* Atau pastikan Dockerfile backend menggunakan user `1000:1000`

---

#### 3️⃣ Atur Ownership & Permission Folder

Jika UID/GID container adalah **1000:1000**, jalankan:

```bash
sudo chown -R 1000:1000 uploads
sudo chmod -R 755 uploads
```

---

📌 **Catatan Penting**

* Folder `uploads` **harus di-mount sebagai volume** di `docker-compose.yml`
* **Jangan menghapus folder ini** saat redeploy
* Salah permission akan menyebabkan error:

  * upload gagal
  * file tidak tersimpan
  * permission denied
  
---

## 5️⃣ Konfigurasi Nginx (HTTPS & Reverse Proxy)

### 5.1 Buat Config Nginx

```bash
sudo nano /etc/nginx/sites-available/petualangancuan
```

```nginx
# ===============================
# HTTP ->  HTTPS
# ===============================
server {
    listen 80;
    server_name petualangancuan.rofid.me;

    return 301 https://petualangancuan.rofid.me$request_uri;
}

# ===============================
# HTTPS
# ===============================
server {
    listen 443 ssl;
    server_name petualangancuan.rofid.me;

    ssl_certificate     /etc/ssl/cloudflare/origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/origin.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # ===============================
    # FRONTEND (Vue)
    # ===============================
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # ===============================
    # BACKEND (API)
    # ===============================
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # ===============================
    # STATIC UPLOADS (Backend)
    # ===============================
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

}
```

---

### 5.2 Aktifkan & Restart Nginx

```bash
sudo ln -s /etc/nginx/sites-available/petualangancuan /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

---

## 6️⃣ Deployment Otomatis (CI/CD)

Setiap push ke branch `main` akan otomatis:

1. Build image frontend & backend
2. Push ke GHCR
3. SSH ke VPS
4. Jalankan:

   ```bash
   docker compose pull
   docker compose up -d
   ```

Trigger manual:

```bash
git push origin main
```

---

## 7️⃣ Verifikasi & Troubleshooting

### Cek Container

```bash
cd ~/petualangan-cuan
docker compose ps
```

### Lihat Log Backend

```bash
docker compose logs -f backend
```

### Jika ENV Tidak Ter-update

```bash
docker compose down
docker compose up -d
```

### Jika OAuth Masih Redirect ke Localhost

* Cek `.env`
* Pastikan `GOOGLE_REDIRECT_URL` benar
* Restart container backend

---

## ✅ Selesai

Aplikasi dapat diakses di:

👉 **[https://petualangancuan.rofid.me](https://petualangancuan.rofid.me)**