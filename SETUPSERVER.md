# 🚀 Panduan Lengkap Setup Server & Deployment

## **Petualangan Cuan**

*(VPS + Docker + GitHub Actions + Nginx + HTTPS)*

Dokumen ini menjelaskan **end-to-end deployment** aplikasi **Petualangan Cuan** ke VPS menggunakan:

* Docker & Docker Compose
* GitHub Actions (CI/CD)
* GitHub Container Registry (GHCR)
* Nginx sebagai Reverse Proxy + HTTPS
* Multi-Environment Isolation (**Production** & **Staging**)

---

## 📑 Daftar Isi

1. Setup Akses SSH (Laptop → Server)
2. Konfigurasi GitHub Secrets
3. Persiapan Awal Server (VPS)
4. Setup Project, Shared Assets, & Environment Variables
5. Konfigurasi Nginx (HTTPS & Reverse Proxy)
6. Deployment (Manual Awal & Otomatis CI/CD)
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

```text
.github/workflows/deploy.yml
.github/workflows/deploy-staging.yml
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
sudo apt update
sudo apt install -y \
    ca-certificates \
    curl \
    gnupg \
    git \
    unzip \
    zip \
    htop \
    ncdu \
    build-essential \
    nginx

sudo install -m 0755 -d /etc/apt/keyrings

sudo curl -fsSL \
    https://download.docker.com/linux/ubuntu/gpg \
    -o /etc/apt/keyrings/docker.asc

sudo chmod a+r /etc/apt/keyrings/docker.asc

echo \
  "Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc" | \
sudo tee /etc/apt/sources.list.d/docker.sources > /dev/null

sudo apt update

sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```

Cek instalasi:

```bash
docker --version
docker compose version
```

---

### 3.2 Login ke GitHub Container Registry (GHCR)

Agar server bisa menarik (*pull*) image dari GitHub:

1. Buat **Personal Access Token (Classic)**:
   - Buka: [https://github.com/settings/tokens](https://github.com/settings/tokens) (atau via **GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)**)
   - Klik **Generate new token (classic)**
   - Pilih scope minimum:
     * `read:packages`
     * `repo`
     * `workflow`

2. Login di server VPS:

```bash
echo "GITHUB_PAT_ANDA" | docker login ghcr.io -u USERNAME_GITHUB --password-stdin
```

---

## 4️⃣ Setup Project, Shared Assets, & Environment Variables

### 4.1 Struktur Direktori Multi-Environment di Server

Aplikasi dibagi menjadi dua environment terisolasi (**Production** dan **Staging**) dengan konfigurasi monitoring masing-masing serta direktori **shared** untuk model AI:

```text
/home/<SERVER_USER>/petualangan-cuan/
├── shared/
│   └── ai-models/
│       ├── google_gemma-3-4b-it-Q4_K_M.gguf
│       ├── mmproj-google_gemma-3-4b-it-f16.gguf
│       └── whisper/
├── staging/
│   ├── .env.staging
│   ├── docker-compose.staging.yml
│   ├── monitoring/
│   │   ├── promtail-config.yml
│   │   └── grafana/
│   │       └── provisioning/
│   ├── uploads_staging/
│   └── wa-gateway-data_staging/
└── production/
    ├── .env.prod
    ├── docker-compose.prod.yml
    ├── monitoring/
    │   ├── promtail-config.yml
    │   └── grafana/
    │       └── provisioning/
    ├── uploads_prod/
    └── wa-gateway-data_prod/
```

Buat struktur folder awal di VPS:

```bash
mkdir -p ~/petualangan-cuan/shared/ai-models/whisper
mkdir -p ~/petualangan-cuan/production/uploads_prod
mkdir -p ~/petualangan-cuan/production/wa-gateway-data_prod
mkdir -p ~/petualangan-cuan/staging/uploads_staging
mkdir -p ~/petualangan-cuan/staging/wa-gateway-data_staging
```

---

### 4.2 Atur Ownership & Permission Folder `uploads_*`

Folder ini digunakan untuk menyimpan file upload (gambar, dokumen, struk transaksi, dll) dan **harus persisten**.
Container backend berjalan dengan user `appuser` (UID `100`, GID `101`). Atur izin folder secara langsung:

```bash
# Production uploads permission
sudo chown -R 100:101 ~/petualangan-cuan/production/uploads_prod
sudo chmod -R 755 ~/petualangan-cuan/production/uploads_prod

# Staging uploads permission
sudo chown -R 100:101 ~/petualangan-cuan/staging/uploads_staging
sudo chmod -R 755 ~/petualangan-cuan/staging/uploads_staging
```

📌 **Catatan Penting**

* Folder `uploads_*` **harus di-mount sebagai volume** di docker-compose masing-masing
* **Jangan menghapus folder ini** saat redeploy
* Salah permission akan menyebabkan error `upload gagal` / `permission denied`

---

### 4.3 Matriks Port & Nama Container (Production vs Staging)

| Service | Container Production | Port Prod | Container Staging | Port Staging | Keterangan |
|---|---|---|---|---|---|
| **Frontend** | `prod_cuan_frontend` | `3000` | `stg_cuan_frontend` | `3300` | UI SPA |
| **Backend API** | `prod_cuan_backend` | `8080` | `stg_cuan_backend` | `8888` | REST API |
| **PostgreSQL** | `prod_cuan_db` | `5432` | `stg_cuan_db` | `5555` | Database |
| **WA Gateway** | `prod_cuan_wa_gateway` | `3003` | `stg_cuan_wa_gateway` | `3330` | WhatsApp bridge |
| **Local LLM** | `prod_cuan_llm` | `8081` | `stg_cuan_llm` | `8881` | GGUF Inference |
| **Whisper** | `prod_cuan_whisper` | `8000` | `stg_cuan_whisper` | `8800` | Speech to Text |
| **Loki** | `prod_cuan_loki` | `3100` | `stg_cuan_loki` | `3110` | Log Collector |
| **Promtail** | `prod_cuan_promtail` | - | `stg_cuan_promtail` | - | Log Scraper |
| **Grafana** | `prod_cuan_grafana` | `3002` | `stg_cuan_grafana` | `3302` | Monitoring UI |
| **Uptime Kuma** | `prod_cuan_uptime` | `3001` | `stg_cuan_uptime` | `3301` | Health Check |

---

### 4.4 Setup File Environment Variables (`.env.prod` & `.env.staging`)

> [!IMPORTANT]
> - Template variabel dapat dilihat pada [.env.example](file:///.env.example).
> - **JANGAN PERNAH** meng-commit file `.env.prod` atau `.env.staging` ke git repository.

#### 🅰️ Buat File `.env.prod` Production (`~/petualangan-cuan/production/.env.prod`)

```bash
nano ~/petualangan-cuan/production/.env.prod
```

Contoh isi:

```env
# Server
PORT=8080
TZ=Asia/Jakarta

# Database (PostgreSQL)
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_aman_production
DB_NAME=petualangan_cuan_prod

# Security (Generate: openssl rand -base64 32)
JWT_SECRET=string_random_panjang_production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=72h

# Google OAuth (WAJIB PRODUKSI)
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxxxx
GOOGLE_REDIRECT_URL=https://petualangancuan.rofid.me/api/auth/google/callback

# Frontend
FRONTEND_URL=https://petualangancuan.rofid.me

# AI Configuration
AI_PROVIDER=local
LOCAL_LLM_URL=http://llm-server:8080
LOCAL_WHISPER_URL=http://whisper-server:8000
EXTERNAL_AI_URL=
EXTERNAL_AI_API_KEY=
EXTERNAL_AI_MODEL=

# WhatsApp Gateway (prod_cuan_wa_gateway)
WA_GATEWAY_USERNAME=admin
WA_GATEWAY_PASSWORD=password_wa_kuat
WA_WEBHOOK_URL=http://prod_cuan_backend:8080/api/webhook/whatsapp
WHATSAPP_WEBHOOK_SECRET=secret_webhook_acak_32_karakter
WA_GATEWAY_URL=http://prod_cuan_wa_gateway:3000

# Observability
GRAFANA_USER=admin
GRAFANA_PASSWORD=admincuanprod

# Docker Host Port
FRONTEND_PORT=3000
BACKEND_PORT=8080
LLM_PORT=8081
WHISPER_PORT=8000
WA_GATEWAY_PORT=3003
LOKI_PORT=3100
GRAFANA_PORT=3002
UPTIME_KUMA_PORT=3001
```

⚠️ **Catatan Penting**

* Jangan gunakan `localhost` di production
* OAuth redirect **HARUS domain publik**
* Jika masih redirect ke localhost → `.env.prod` **belum ter-load**

---

#### 🅱️ Buat File `.env.staging` Staging (`~/petualangan-cuan/staging/.env.staging`)

```bash
nano ~/petualangan-cuan/staging/.env.staging
```

Contoh isi:

```env
# Server
PORT=8080
TZ=Asia/Jakarta

# Database (PostgreSQL)
DB_HOST=postgres
DB_PORT=5555
DB_USER=postgres
DB_PASSWORD=password_aman_staging
DB_NAME=petualangan_cuan_staging

# Security (Generate: openssl rand -base64 32)
JWT_SECRET=string_random_panjang_staging
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=72h

# Google OAuth (Staging)
GOOGLE_CLIENT_ID=xxxxx.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=xxxxx
GOOGLE_REDIRECT_URL=https://staging.petualangancuan.rofid.me/api/auth/google/callback

# Frontend
FRONTEND_URL=https://staging.petualangancuan.rofid.me

# AI Configuration
AI_PROVIDER=local
LOCAL_LLM_URL=http://llm-server:8080
LOCAL_WHISPER_URL=http://whisper-server:8000
EXTERNAL_AI_URL=
EXTERNAL_AI_API_KEY=
EXTERNAL_AI_MODEL=

# WhatsApp Gateway (stg_cuan_wa_gateway)
WA_GATEWAY_USERNAME=admin
WA_GATEWAY_PASSWORD=password_wa_staging
WA_WEBHOOK_URL=http://stg_cuan_backend:8080/api/webhook/whatsapp
WHATSAPP_WEBHOOK_SECRET=secret_webhook_acak_staging
WA_GATEWAY_URL=http://stg_cuan_wa_gateway:3000

# Observability
GRAFANA_USER=admin
GRAFANA_PASSWORD=admincuanstaging

# Docker Host Port (Staging Offset)
FRONTEND_PORT=3300
BACKEND_PORT=8888
LLM_PORT=8881
WHISPER_PORT=8800
WA_GATEWAY_PORT=3330
LOKI_PORT=3110
GRAFANA_PORT=3302
UPTIME_KUMA_PORT=3301
```

---

### 4.5 Salin File `docker-compose`, Assets `monitoring`, & Upload Model AI

#### 1️⃣ Salin file `docker-compose` dari laptop ke VPS
Jalankan dari terminal laptop di direktori root project:

```bash
scp docker-compose.prod.yml user@IP_SERVER:~/petualangan-cuan/production/
scp docker-compose.staging.yml user@IP_SERVER:~/petualangan-cuan/staging/
```

📌 **WAJIB**

* Isi **HARUS SAMA** dengan versi di repository
* Jika file di repo berubah → **update juga di server**

---

#### 2️⃣ Salin Folder `monitoring/` ke Staging & Production di VPS

Konfigurasi Promtail dan Grafana Provisioning disalin ke masing-masing folder environment.

> [!CAUTION]
> **Jika sebelumnya sempat menjalankan `docker compose up` dan muncul error mount folder:**
> Hapus folder palsu yang sempat dibuat otomatis oleh Docker di VPS terlebih dahulu:
> ```bash
> sudo rm -rf ~/petualangan-cuan/staging/monitoring
> sudo rm -rf ~/petualangan-cuan/production/monitoring
> ```

Salin folder `monitoring` dari laptop ke VPS:

```bash
scp -r monitoring user@IP_SERVER:~/petualangan-cuan/staging/
scp -r monitoring user@IP_SERVER:~/petualangan-cuan/production/
```

---

#### 3️⃣ Upload / Download File Model AI ke Shared Directory

**Opsi A — Upload dari laptop (jika sudah didownload lokal):**

```bash
scp ai-models/google_gemma-3-4b-it-Q4_K_M.gguf user@IP_SERVER:~/petualangan-cuan/shared/ai-models/
scp ai-models/mmproj-google_gemma-3-4b-it-f16.gguf user@IP_SERVER:~/petualangan-cuan/shared/ai-models/
```

**Opsi B — Download langsung di server VPS:**

```bash
cd ~/petualangan-cuan/shared/ai-models
wget -c -O google_gemma-3-4b-it-Q4_K_M.gguf \
"https://huggingface.co/bartowski/google_gemma-3-4b-it-GGUF/resolve/main/google_gemma-3-4b-it-Q4_K_M.gguf?download=true"
wget -c -O mmproj-google_gemma-3-4b-it-f16.gguf \
"https://huggingface.co/bartowski/google_gemma-3-4b-it-GGUF/resolve/main/mmproj-google_gemma-3-4b-it-f16.gguf?download=true"
```

---

## 5️⃣ Konfigurasi Nginx (HTTPS & Reverse Proxy)

Konfigurasi Nginx dipisahkan menjadi 2 file virtual host terpisah di `/etc/nginx/sites-available/`:

---

### 5.1 Buat Config Nginx Production (`petualangancuan_prod`)

```bash
sudo nano /etc/nginx/sites-available/petualangancuan_prod
```

```nginx
# ===============================
# HTTP -> HTTPS
# ===============================
server {
    listen 80;
    server_name petualangancuan.rofid.me;

    return 301 https://petualangancuan.rofid.me$request_uri;
}

# ===============================
# HTTPS (Production)
# ===============================
server {
    listen 443 ssl;
    server_name petualangancuan.rofid.me;

    client_max_body_size 25M;

    ssl_certificate     /etc/ssl/cloudflare/origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/origin.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # FRONTEND (Vue - Production Port 3000)
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # BACKEND (API - Production Port 8080)
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # STATIC UPLOADS (Backend Production)
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

### 5.2 Buat Config Nginx Staging (`petualangancuan_staging`)

```bash
sudo nano /etc/nginx/sites-available/petualangancuan_staging
```

```nginx
# ===============================
# HTTP -> HTTPS
# ===============================
server {
    listen 80;
    server_name staging.petualangancuan.rofid.me;

    return 301 https://staging.petualangancuan.rofid.me$request_uri;
}

# ===============================
# HTTPS (Staging)
# ===============================
server {
    listen 443 ssl;
    server_name staging.petualangancuan.rofid.me;

    client_max_body_size 25M;

    ssl_certificate     /etc/ssl/cloudflare/origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/origin.key;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # FRONTEND (Vue - Staging Port 3300)
    location / {
        proxy_pass http://127.0.0.1:3300;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # BACKEND (API - Staging Port 8888)
    location /api/ {
        proxy_pass http://127.0.0.1:8888;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # STATIC UPLOADS (Backend Staging)
    location /uploads/ {
        proxy_pass http://127.0.0.1:8888/uploads/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

### 5.3 Aktifkan Virtual Host & Restart Nginx

```bash
# Aktifkan konfigurasi Production & Staging
sudo ln -sf /etc/nginx/sites-available/petualangancuan_prod /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/petualangancuan_staging /etc/nginx/sites-enabled/

# Test syntax konfigurasi
sudo nginx -t

# Restart Nginx
sudo systemctl restart nginx
```

---

## 6️⃣ Deployment (Manual Awal & Otomatis CI/CD)

### 6.1 Bootstrapping Pertama Kali (Manual di Server)

Setelah file `.env.*`, folder `monitoring/`, dan `docker-compose.*.yml` berada di tempatnya, jalankan container pertama kali:

#### Menjalankan Staging Pertama Kali:
```bash
cd ~/petualangan-cuan/staging
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging pull
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging up -d
```

#### Menjalankan Production Pertama Kali:
```bash
cd ~/petualangan-cuan/production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod pull
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod up -d
```

---

### 6.2 Deployment Otomatis (CI/CD GitHub Actions)

Setelah setup server selesai, deployment selanjutnya akan berjalan otomatis melalui push git:

#### 🧪 Branch Staging (`staging`)
Setiap push ke branch `staging` akan otomatis memicu `.github/workflows/deploy-staging.yml`:
1. Build & Push image ke GHCR dengan tag `:staging`
2. SSH ke VPS dan melakukan rolling update di `~/petualangan-cuan/staging`

```bash
git push origin staging
```

#### 🚀 Branch Production (`main`)
Setiap push ke branch `main` akan otomatis memicu `.github/workflows/deploy.yml`:
1. Build & Push image ke GHCR dengan tag `:latest`
2. SSH ke VPS dan melakukan rolling update di `~/petualangan-cuan/production`

```bash
git push origin main
```

---

## 7️⃣ Verifikasi & Troubleshooting

### 7.1 Cek Status Container

Pastikan seluruh container berstatus `Up` / `running`:

```bash
# Production
cd ~/petualangan-cuan/production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod ps

# Staging
cd ~/petualangan-cuan/staging
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging ps
```

---

### 7.2 Verifikasi UID/GID Container Backend (Setelah Container Berjalan)

Untuk memverifikasi bahwa container backend menggunakan user `appuser` (UID `100`, GID `101`):

```bash
# Production
docker exec -it prod_cuan_backend id appuser

# Staging
docker exec -it stg_cuan_backend id appuser
```

Output yang **diharapkan**:
```text
uid=100(appuser) gid=101(appgroup)
```

---

### 7.3 Lihat Log Backend

```bash
# Production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod logs -f backend

# Staging
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging logs -f backend
```

---

### 7.4 Jika ENV Tidak Ter-update

```bash
# Production
cd ~/petualangan-cuan/production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod down
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod up -d

# Staging
cd ~/petualangan-cuan/staging
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging down
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging up -d
```

---

### 7.5 Jika OAuth Masih Redirect ke Localhost

* Cek file `.env.prod` atau `.env.staging` di folder environment terkait
* Pastikan `GOOGLE_REDIRECT_URL` dan `FRONTEND_URL` sudah mengarah ke domain HTTPS yang benar
* Restart container backend terkait (`prod_cuan_backend` atau `stg_cuan_backend`)

---

## ✅ Selesai

Aplikasi dapat diakses di:

* 🚀 **Production**: [https://petualangancuan.rofid.me](https://petualangancuan.rofid.me)
* 🧪 **Staging**: [https://staging.petualangancuan.rofid.me](https://staging.petualangancuan.rofid.me)