# 🚀 Panduan Lengkap Setup Server & Deployment

## **Petualangan Cuan**
*(VPS Ubuntu + Docker Compose + GitHub Actions CI/CD + GHCR + Nginx + Cloudflare HTTPS)*

Dokumen ini menjelaskan panduan **end-to-end deployment** aplikasi **Petualangan Cuan** ke VPS Ubuntu dari server baru (*clean install*) hingga aplikasi berjalan stabil di lingkungan **Production** dan **Staging** yang terisolasi.

---

## 📑 Daftar Isi

1. [Prasyarat & Inventarisasi Konfigurasi](#1-prasyarat--inventarisasi-konfigurasi)
2. [Spesifikasi Minimum & Rekomendasi VPS](#2-spesifikasi-minimum--rekomendasi-vps)
3. [Inisialisasi & Pengamanan Sistem Ubuntu](#3-inisialisasi--pengamanan-sistem-ubuntu)
4. [User Non-Root, SSH Key, & Hardening](#4-user-non-root-ssh-key--hardening)
5. [Konfigurasi Firewall (UFW & Cloud Security Group)](#5-konfigurasi-firewall-ufw--cloud-security-group)
6. [Instalasi Docker, NVM, & Tooling Server](#6-instalasi-docker-nvm--tooling-server)
7. [Arsitektur Database PostgreSQL (Containerized)](#7-arsitektur-database-postgresql-containerized)
8. [Autentikasi GitHub Container Registry (GHCR)](#8-autentikasi-github-container-registry-ghcr)
9. [Struktur Direktori Server Multi-Environment](#9-struktur-direktori-server-multi-environment)
10. [Shared Assets & Model AI Lokal (Opsional)](#10-shared-assets--model-ai-lokal-opsional)
11. [Matriks & Konfigurasi Environment Variables](#11-matriks--konfigurasi-environment-variables)
12. [Setup Environment Production](#12-setup-environment-production)
13. [Setup Environment Staging](#13-setup-environment-staging)
14. [Setup Domain Cloudflare, SSL Origin, & Nginx Reverse Proxy](#14-setup-domain-cloudflare-ssl-origin--nginx-reverse-proxy)
15. [Bootstrapping & Deployment Manual Pertama Kali](#15-bootstrapping--deployment-manual-pertama-kali)
16. [Otomasi CI/CD via GitHub Actions](#16-otomasi-cicd-via-github-actions)
17. [Health Check & Monitoring (Loki, Promtail, Grafana, Uptime Kuma)](#17-health-check--monitoring)
18. [Prosedur Backup & Restore Database](#18-prosedur-backup--restore-database)
19. [Prosedur Rollback & Disaster Recovery](#19-prosedur-rollback--disaster-recovery)
20. [Panduan Troubleshooting & FAQ](#20-panduan-troubleshooting--faq)
21. [Checklist Deployment Final](#21-checklist-deployment-final)

---

## 1. Prasyarat & Inventarisasi Konfigurasi

### 1.1 Komponen Aplikasi Aktual
Berdasarkan source code repository:
* **Backend**: Go 1.24 (Fiber v2, GORM, PostgreSQL driver, Zerolog)
* **Frontend**: Vue 3 (Vite, TailwindCSS, Axios, Pinia, Nginx static runtime)
* **WhatsApp Gateway**: Go WhatsApp bridge (SQLite session, REST & Webhook)
* **Database**: PostgreSQL 15 Alpine (dijalankan via Docker Container terisolasi per environment)
* **AI Provider**: Fleksibel antara `external` (OpenRouter/Gemini API) atau `local` (Gemma 3 GGUF via llama.cpp server + Faster Whisper)
* **Observability**: Grafana Loki (Log collector), Promtail (Log scraper dari Docker socket), Grafana (Dashboard), Uptime Kuma (Uptime monitor)
* **Reverse Proxy**: Nginx pada Host VPS + Cloudflare Full (Strict) SSL Origin Certificate

### 1.2 Daftar File Konfigurasi Terkait
| File | Lokasi Repository | Fungsi Utama |
|---|---|---|
| docker-compose.prod.yml | Root | Orkestrasi stack Production |
| docker-compose.staging.yml | Root | Orkestrasi stack Staging |
| .env.example | Root | Template master environment variables |
| .github/workflows/deploy.yml | Root | CI/CD build & deploy otomatis ke Production (branch `main`) |
| .github/workflows/deploy-staging.yml | Root | CI/CD build & deploy otomatis ke Staging (branch `staging`) |
| backend-go/Dockerfile | `backend-go/` | Multi-stage Docker build untuk Go backend binary |
| frontend-vue/Dockerfile | `frontend-vue/` | Multi-stage Docker build Vite SPA + Nginx |
| monitoring/promtail-config.yml | `monitoring/` | Konfigurasi log scraping Docker container |

---

## 2. Spesifikasi Minimum & Rekomendasi VPS

### Opsi A — Mode AI External (Rekomendasi untuk VPS Budget / 2GB - 4GB RAM)
Menggunakan `AI_PROVIDER=external` (OpenRouter / Google Gemini API). Container LLM dan Whisper tidak dijalankan (`COMPOSE_PROFILES=` kosong), sehingga sangat hemat sumber daya.
* **CPU**: 2 vCPU
* **RAM**: 2 GB (Minimum) / 4 GB (Rekomendasi)
* **Storage**: 25 GB – 40 GB SSD / NVMe
* **OS**: Ubuntu 22.04 LTS atau Ubuntu 24.04 LTS (64-bit x86_64 / amd64)

### Opsi B — Mode AI Local (Inference On-Server)
Menggunakan `AI_PROVIDER=local` dan `COMPOSE_PROFILES=local-ai` (menjalankan Gemma 3 4B GGUF + Faster-Whisper CPU).
* **CPU**: 4 vCPU atau lebih
* **RAM**: 8 GB – 16 GB
* **Storage**: 60 GB – 100 GB SSD
* **Catatan**: LLM quantization Q4_K_M membutuhkan alokasi RAM sekitar ~3 GB, Whisper butuh ~1 GB.

---

## 3. Inisialisasi & Pengamanan Sistem Ubuntu

Jalankan langkah-langkah berikut pada VPS Ubuntu baru (login awal via console cloud provider atau terminal SSH sebagai `root`).

### 3.1 Update & Upgrade Sistem
* **Tujuan**: Memastikan semua package OS memiliki patch keamanan terbaru.
* **Prasyarat**: Akses user `root` atau sudoers.
* **Lokasi Eksekusi**: Terminal VPS.

```bash
sudo apt update && sudo apt upgrade -y
sudo apt autoremove -y
```

### 3.2 Konfigurasi Timezone Indonesia
* **Tujuan**: Menyelaraskan timestamp sistem dan log server dengan Waktu Indonesia Barat (WIB).
* **Lokasi Eksekusi**: Terminal VPS.

```bash
sudo timedatectl set-timezone Asia/Jakarta
```

* **Cara Verifikasi**:
```bash
timedatectl
```
*Hasil yang diharapkan: `Time zone: Asia/Jakarta (WIB, +0700)`.*

---

## 4. User Non-Root, SSH Key, & Hardening

> [!CAUTION]
> Jangan menjalankan aplikasi atau operasional harian langsung menggunakan akun `root`.

### 4.1 Membuat User Sudo Non-Root
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Buat user baru (misal: ubuntu)
sudo adduser ubuntu

# 2. Berikan hak akses sudo
sudo usermod -aG sudo ubuntu
```

### 4.2 Setup SSH Key Authentication (Laptop → Server)
Gunakan algoritma **ED25519** untuk keamanan dan performa terbaik.

#### 🅰️ Generate SSH Key di Laptop (Terminal Lokal)
* **Lokasi Eksekusi**: Terminal Laptop.

```bash
ssh-keygen -t ed25519 -C "admin@petualangancuan"
```
*File yang dihasilkan:*
* Private Key: `~/.ssh/id_ed25519` (simpan aman, jangan pernah di-commit ke Git)
* Public Key: `~/.ssh/id_ed25519.pub`

#### 🅱️ Kirim Public Key ke VPS
* **Opsi 1 (Otomatis via `ssh-copy-id` dari Laptop):**
```bash
ssh-copy-id -i ~/.ssh/id_ed25519.pub ubuntu@IP_PUBLIC_VPS
```

* **Opsi 2 (Manual di VPS):**
```bash
# Di VPS (sebagai user ubuntu):
mkdir -p ~/.ssh
nano ~/.ssh/authorized_keys
# Paste isi dari file id_ed25519.pub laptop Anda ke sini, lalu simpan (Ctrl+O, Enter, Ctrl+X)

chmod 700 ~/.ssh
chmod 600 ~/.ssh/authorized_keys
```

### 4.3 Setup Host Alias di Laptop (`~/.ssh/config` atau `ssh tencent`)
Agar tidak perlu mengetikkan `ssh -i ... ubuntu@IP_SERVER` setiap kali login, buat konfigurasi alias SSH di laptop Anda:
* **Windows**: `C:\Users\<USERNAME>\.ssh\config`
* **Linux / macOS**: `~/.ssh/config`

Tambahkan blok konfigurasi berikut:

```text
# Opsi 1: Menggunakan Hostname DNS Cloudflare (setelah DNS A Record ssh.rofid.me dibuat)
Host tencent
    HostName ssh.rofid.me
    User ubuntu
    IdentityFile ~/.ssh/id_ed25519
    # Di Windows bisa menggunakan path absolut jika perlu:
    # IdentityFile C:\Users\<NAMA_USER>\.ssh\id_ed25519

# Opsi 2: Menggunakan IP Publik Langsung
Host petualangancuan-vps
    HostName IP_PUBLIC_VPS
    User ubuntu
    IdentityFile ~/.ssh/id_ed25519
```

*Uji koneksi dari laptop:*
```bash
ssh tencent
# atau: ssh petualangancuan-vps
```

### 4.4 Hardening SSH Daemon di Server
* **Tujuan**: Menonaktifkan login password dan login langsung user root untuk mencegah serangan brute-force.
* **Lokasi Eksekusi**: Terminal VPS.

```bash
sudo nano /etc/ssh/sshd_config
```
Pastikan opsi berikut dikonfigurasi:
```text
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
```

Uji konfigurasi dan restart service SSH:
```bash
sudo sshd -t && sudo systemctl restart ssh
```

> [!WARNING]
> **PENTING**: Jangan tutup sesi terminal SSH saat ini sebelum membuka tab terminal baru dan memverifikasi login berhasil via `ssh petualangancuan-vps`.

---

### 4.5 Setup Terminal Zsh, Oh My Zsh, Plugin, & Tema `af-magic`
Pasang shell **Zsh** dan framework **Oh My Zsh** lengkap dengan tema `af-magic`, auto-suggestions, dan syntax highlighting untuk produktivitas di terminal server:
* **Lokasi Eksekusi**: Terminal VPS (sebagai user `ubuntu`).

```bash
# 1. Install Zsh, Git, & Curl
sudo apt install -y zsh git curl

# 2. Install Oh My Zsh (mode non-interaktif)
sh -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended

# 3. Download Plugin zsh-autosuggestions
git clone https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions

# 4. Download Plugin zsh-syntax-highlighting
git clone https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting

# 5. Set tema af-magic di ~/.zshrc
sed -i 's/ZSH_THEME="robbyrussell"/ZSH_THEME="af-magic"/' ~/.zshrc

# 6. Aktifkan plugin di ~/.zshrc
sed -i 's/plugins=(git)/plugins=(git zsh-autosuggestions zsh-syntax-highlighting)/' ~/.zshrc

# 7. Jadikan Zsh sebagai default shell untuk user saat ini (ubuntu)
sudo chsh -s $(which zsh) $USER

# 8. Masuk ke shell Zsh
zsh
```

---

## 5. Konfigurasi Firewall (UFW & Cloud Security Group)

### 5.1 Cloud Security Group (Tencent Cloud / AWS / GCP / DigitalOcean)
Pastikan aturan **Inbound / Masuk** pada dashboard Cloud Provider mengizinkan:

| Protocol | Port / Range | Source (IP Asal) | Policy / Action | Keterangan / Deskripsi |
|---|---|---|---|---|
| **TCP** | `22` | `0.0.0.0/0` (atau IP Laptop) | **ACCEPT / ALLOW** | Akses SSH administratif |
| **TCP** | `80` | `0.0.0.0/0` | **ACCEPT / ALLOW** | HTTP traffic (Nginx redirect to HTTPS) |
| **TCP** | `443` | `0.0.0.0/0` | **ACCEPT / ALLOW** | HTTPS traffic (Cloudflare SSL) |
| **ALL / ICMP** | `ALL` | `0.0.0.0/0` | **ACCEPT / ALLOW** | Ping / Network Health Check (Opsional) |

> [!IMPORTANT]
> Port database (5432, 5555), backend internal (8080, 8888), WA Gateway (3003, 3330), dan monitoring (3001, 3002, 3100, 3110, 3301, 3302) **TIDAK PERLU** dibuka di firewall publik cloud karena hanya diakses oleh Nginx reverse proxy internal atau jaringan localhost (`127.0.0.1`).

### 5.2 Konfigurasi UFW di VPS
* **Lokasi Eksekusi**: Terminal VPS.

```bash
sudo apt install -y ufw

# Set default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# Izinkan SSH (WAJIB sebelum enable agar tidak terkunci)
sudo ufw allow 22/tcp

# Izinkan HTTP & HTTPS untuk Nginx
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Aktifkan UFW
sudo ufw enable
```

*Verifikasi status UFW:*
```bash
sudo ufw status verbose
```

---

## 6. Instalasi Docker, NVM, & Tooling Server

### 6.1 Install Docker Engine & Docker Compose Plugin
Pasang Docker resmi dari repository Docker CE (bukan package default ubuntu `docker.io` yang usang).

* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Install dependensi pendukung
sudo apt update
sudo apt install -y ca-certificates curl gnupg git unzip zip htop ncdu nginx build-essential

# 2. Tambahkan GPG key resmi Docker
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# 3. Daftarkan repository Docker
echo \
  "Types: deb
URIs: https://download.docker.com/linux/ubuntu
Suites: $(. /etc/os-release && echo "${UBUNTU_CODENAME:-$VERSION_CODENAME}")
Components: stable
Architectures: $(dpkg --print-architecture)
Signed-By: /etc/apt/keyrings/docker.asc" | \
sudo tee /etc/apt/sources.list.d/docker.sources > /dev/null

# 4. Install Docker Engine & Docker Compose Plugin
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# 5. Masukkan user ubuntu ke group docker
sudo usermod -aG docker $USER
```

*Verifikasi instalasi (keluar dan login kembali agar group docker aktif):*
```bash
docker --version
docker compose version
```
*Hasil yang diharapkan: Docker version 24+ / 27+ dan Docker Compose version v2+.*

---

### 6.2 Install NVM & Node.js 22 LTS
Pasang **Node Version Manager (NVM)** dan gunakan **Node.js versi 22 (LTS)** untuk keperluan scripting atau tooling lokal server jika diperlukan:
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Download dan jalankan script instalasi NVM
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash

# 2. Muat environment NVM ke sesi shell saat ini
source ~/.zshrc 2>/dev/null || source ~/.bashrc

# 3. Cek nvm
nvm --version

# 4. Install dan set default Node.js 22 LTS
nvm install 22
nvm use 22
nvm alias default 22

# 5. Verifikasi instalasi Node & NPM
node -v
npm -v
```

---

## 7. Arsitektur Database PostgreSQL (Containerized)

> [!NOTE]
> **PENTING — BUKAN INSTALASI NATIVE APT**:
> Proyek Petualangan Cuan menggunakan **Containerized PostgreSQL 15 Alpine** yang berjalan di dalam Docker network masing-masing environment ([docker-compose.prod.yml](file:///f:/Project/petualangan-cuan/docker-compose.prod.yml) dan [docker-compose.staging.yml](file:///f:/Project/petualangan-cuan/docker-compose.staging.yml)).
> **JANGAN** menginstal `postgresql` via `apt` pada host Ubuntu karena akan memakan memori dan menyebabkan konflik port 5432 pada host.

### 7.1 Rincian Isolasi Database
| Item | Production Database | Staging Database |
|---|---|---|
| **Service Name** | `postgres` | `postgres` |
| **Container Name** | `prod_cuan_db` | `stg_cuan_db` |
| **Image** | `postgres:15-alpine` | `postgres:15-alpine` |
| **Volume Data (Persisten)** | `postgres_data_prod` | `postgres_data_staging` |
| **Docker Internal Hostname** | `postgres` | `postgres` |
| **Docker Internal Port** | `5432` | `5432` |
| **Host Port Mapping** | `5432:5432` | `5555:5432` |
| **Network** | `app-network` (bridge) | `staging-network` (bridge) |
| **Auto-Migration** | GORM AutoMigrate otomatis saat backend menyala | GORM AutoMigrate otomatis saat backend menyala |

Database dan user diinisialisasi secara otomatis oleh image `postgres:15-alpine` saat pertama kali container dibuat berdasarkan variabel `POSTGRES_USER`, `POSTGRES_PASSWORD`, dan `POSTGRES_DB` yang bersumber dari file `.env.prod` atau `.env.staging`.

---

## 8. Autentikasi GitHub Container Registry (GHCR)

Agar VPS dapat mengunduh (*pull*) image Docker private dari GitHub Container Registry (GHCR):

1. **Buat Personal Access Token (Classic) di GitHub**:
   - Buka link langsung: [https://github.com/settings/tokens](https://github.com/settings/tokens) (atau navigasi via **GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)**)
   - Klik **Generate new token (classic)**
   - Masukkan **Note** (misal: `VPS Petualangan Cuan GHCR`)
   - Atur **Expiration** sesuai kebutuhan (misal: `90 days` atau `No expiration`)
   - Centang scope minimum yang dibutuhkan:
     * `read:packages` (WAJIB — untuk mengunduh / pull container images dari GHCR)
     * `repo` (Akses repositori jika image berstatus private)
     * `workflow` (Opsional jika berinteraksi dengan GitHub Actions)
   - Klik tombol **Generate token** di bagian bawah dan salin token yang muncul (token hanya ditampilkan satu kali).

2. **Login Docker ke GHCR di Server VPS**:
   * **Lokasi Eksekusi**: Terminal VPS.

```bash
echo "GITHUB_PAT_ANDA" | docker login ghcr.io -u USERNAME_GITHUB --password-stdin
```
*Hasil yang diharapkan: `Login Succeeded`.*

---

## 9. Struktur Direktori Server Multi-Environment

Buat hierarki folder server untuk memisahkan konfigurasi, volume data upload, database, dan shared assets:

```text
/home/ubuntu/petualangan-cuan/
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
│   ├── uploads_staging/
│   └── wa-gateway-data_staging/
└── production/
    ├── .env.prod
    ├── docker-compose.prod.yml
    ├── monitoring/
    │   ├── promtail-config.yml
    │   └── grafana/
    ├── uploads_prod/
    └── wa-gateway-data_prod/
```

### Eksekusi Pembuatan Folder di VPS
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Folder Shared AI Models
mkdir -p ~/petualangan-cuan/shared/ai-models/whisper

# 2. Folder Production
mkdir -p ~/petualangan-cuan/production/uploads_prod
mkdir -p ~/petualangan-cuan/production/wa-gateway-data_prod
mkdir -p ~/petualangan-cuan/production/monitoring/grafana/provisioning/dashboards
mkdir -p ~/petualangan-cuan/production/monitoring/grafana/provisioning/datasources

# 3. Folder Staging
mkdir -p ~/petualangan-cuan/staging/uploads_staging
mkdir -p ~/petualangan-cuan/staging/wa-gateway-data_staging
mkdir -p ~/petualangan-cuan/staging/monitoring/grafana/provisioning/dashboards
mkdir -p ~/petualangan-cuan/staging/monitoring/grafana/provisioning/datasources
```

### 9.1 Matriks Port & Nama Container (Production vs Staging)

| Service | Container Production | Port Host Prod | Container Staging | Port Host Staging | Port Internal Container | Keterangan |
|---|---|---|---|---|---|---|
| **Frontend** | `prod_cuan_frontend` | `3000` | `stg_cuan_frontend` | `3300` | `80` | Web UI Vue 3 (SPA Nginx) |
| **Backend API** | `prod_cuan_backend` | `8080` | `stg_cuan_backend` | `8888` | `8080` | REST API (Go Fiber) |
| **PostgreSQL** | `prod_cuan_db` | `5432` | `stg_cuan_db` | `5555` | `5432` | Database PostgreSQL 15 |
| **WA Gateway** | `prod_cuan_wa_gateway` | `3003` | `stg_cuan_wa_gateway` | `3330` | `3000` | WhatsApp Bridge & Webhook |
| **Local LLM** | `prod_cuan_llm` | `8081` | `stg_cuan_llm` | `8881` | `8080` | Gemma 3 llama.cpp server |
| **Whisper** | `prod_cuan_whisper` | `8000` | `stg_cuan_whisper` | `8800` | `8000` | Faster Whisper STT |
| **Loki** | `prod_cuan_loki` | `3100` | `stg_cuan_loki` | `3110` | `3100` | Log Collector |
| **Promtail** | `prod_cuan_promtail` | - | `stg_cuan_promtail` | - | - | Log Scraper Docker |
| **Grafana** | `prod_cuan_grafana` | `3002` | `stg_cuan_grafana` | `3302` | `3000` | Observability Dashboard |
| **Uptime Kuma** | `prod_cuan_uptime` | `3001` | `stg_cuan_uptime` | `3301` | `3001` | Health & Uptime Monitor |

---

### 9.2 Atur Ownership & Permission Folder Uploads
Container backend Go berjalan dengan user non-root `appuser:appgroup` (UID `100`, GID `101`). Folder bind mount harus memiliki permission yang tepat agar upload file struk/gambar transaksi tidak mengalami *permission denied*:

```bash
# Production uploads permission
sudo chown -R 100:101 ~/petualangan-cuan/production/uploads_prod
sudo chmod -R 755 ~/petualangan-cuan/production/uploads_prod

# Staging uploads permission
sudo chown -R 100:101 ~/petualangan-cuan/staging/uploads_staging
sudo chmod -R 755 ~/petualangan-cuan/staging/uploads_staging
```

---

## 10. Shared Assets & Model AI Lokal (Opsional)

Jika menggunakan mode **AI Lokal** (`AI_PROVIDER=local` dan `COMPOSE_PROFILES=local-ai`), siapkan model GGUF di direktori shared:

### 🅰️ Opsi A — Download Langsung di Server VPS:
* **Lokasi Eksekusi**: Terminal VPS.

```bash
cd ~/petualangan-cuan/shared/ai-models

# Gemma 3 4B GGUF Model
wget -c -O google_gemma-3-4b-it-Q4_K_M.gguf \
"https://huggingface.co/bartowski/google_gemma-3-4b-it-GGUF/resolve/main/google_gemma-3-4b-it-Q4_K_M.gguf?download=true"

# Gemma 3 Multimodal Projector (Image Understanding)
wget -c -O mmproj-google_gemma-3-4b-it-f16.gguf \
"https://huggingface.co/bartowski/google_gemma-3-4b-it-GGUF/resolve/main/mmproj-google_gemma-3-4b-it-f16.gguf?download=true"
```

### 🅱️ Opsi B — Upload dari Laptop (Jika File Sudah Didownload di Lokal):
* **Lokasi Eksekusi**: Terminal Laptop.

```bash
scp ai-models/google_gemma-3-4b-it-Q4_K_M.gguf ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/shared/ai-models/
scp ai-models/mmproj-google_gemma-3-4b-it-f16.gguf ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/shared/ai-models/
```

*(Jika menggunakan `AI_PROVIDER=external`, langkah ini dapat dilewati).*

---

## 11. Matriks & Konfigurasi Environment Variables

### 11.1 Matriks Environment Variables
| Variable | Definisi & Konsumsi | Production | Staging | Secret? | Deskripsi |
|---|---|---|---|---|---|
| `PORT` | Backend Internal | `8080` | `8080` | Tidak | Port listen backend di dalam container |
| `TZ` | System | `Asia/Jakarta` | `Asia/Jakarta` | Tidak | Timezone aplikasi |
| `DB_HOST` | Backend connection | `postgres` | `postgres` | Tidak | Hostname service PostgreSQL di docker network |
| `DB_PORT` | Postgres host map | `5432` | `5555` | Tidak | Port host yang diexpose oleh container DB |
| `DB_USER` | PostgreSQL & Backend | `cuanpsqlprod` | `cuanpsqlstg` | Tidak | Username PostgreSQL |
| `DB_PASSWORD` | PostgreSQL & Backend | *[Generated]* | *[Generated]* | **YA** | Password PostgreSQL |
| `DB_NAME` | PostgreSQL & Backend | `cuan_prod` | `cuan_stg` | Tidak | Nama database |
| `JWT_SECRET` | Backend Auth | *[Generated]* | *[Generated]* | **YA** | Kunci token JWT (min. 32 karakter acak) |
| `JWT_ACCESS_EXPIRY` | Backend Auth | `15m` | `15m` | Tidak | Masa aktif access token |
| `JWT_REFRESH_EXPIRY` | Backend Auth | `72h` | `72h` | Tidak | Masa aktif refresh token |
| `GOOGLE_CLIENT_ID` | Backend OAuth | *[Dari Google Console]* | *[Dari Google Console]* | Tidak | Client ID Google OAuth |
| `GOOGLE_CLIENT_SECRET` | Backend OAuth | *[Dari Google Console]* | *[Dari Google Console]* | **YA** | Client Secret Google OAuth |
| `GOOGLE_REDIRECT_URL` | Backend OAuth | `https://petualangancuan.rofid.me/api/auth/google/callback` | `https://stagingpetualangancuan.rofid.me/api/auth/google/callback` | Tidak | Callback URL OAuth publik |
| `FRONTEND_URL` | Backend CORS | `https://petualangancuan.rofid.me` | `https://stagingpetualangancuan.rofid.me` | Tidak | Domain frontend untuk CORS policy |
| `AI_PROVIDER` | Backend AI Service | `external` / `local` | `external` / `local` | Tidak | Pilihan provider AI |
| `COMPOSE_PROFILES` | Docker Compose | *(kosongkan jika external)* | *(kosongkan jika external)* | Tidak | Profile compose (`local-ai` jika mode lokal) |
| `EXTERNAL_AI_URL` | Backend AI Service | `https://openrouter.ai/api/v1/chat/completions` | `https://openrouter.ai/api/v1/chat/completions` | Tidak | Endpoint external AI API |
| `EXTERNAL_AI_API_KEY` | Backend AI Service | *[OpenRouter API Key]* | *[OpenRouter API Key]* | **YA** | API Key external AI provider |
| `EXTERNAL_AI_MODEL` | Backend AI Service | `google/gemini-2.0-flash-001` | `google/gemini-2.0-flash-001` | Tidak | Model external yang digunakan |
| `WA_GATEWAY_USERNAME` | WA Gateway Auth | `adminwaprod` | `adminwastaging` | Tidak | Basic auth username WA Gateway |
| `WA_GATEWAY_PASSWORD` | WA Gateway Auth | *[Generated]* | *[Generated]* | **YA** | Basic auth password WA Gateway |
| `WA_WEBHOOK_URL` | WA Gateway -> Backend | `http://prod_cuan_backend:8080/api/webhook/whatsapp` | `http://stg_cuan_backend:8080/api/webhook/whatsapp` | Tidak | Endpoint webhook internal |
| `WHATSAPP_WEBHOOK_SECRET` | Backend WA HMAC | *[Generated 32-char]* | *[Generated 32-char]* | **YA** | Secret HMAC webhook WhatsApp |
| `WA_GATEWAY_URL` | Backend -> WA Gateway | `http://prod_cuan_wa_gateway:3000` | `http://stg_cuan_wa_gateway:3000` | Tidak | URL internal WA Gateway |
| `GRAFANA_USER` | Grafana Security | `admingrafanaprod` | `admingrafanastaging` | Tidak | Username admin Grafana |
| `GRAFANA_PASSWORD` | Grafana Security | *[Generated]* | *[Generated]* | **YA** | Password admin Grafana |

---

## 12. Setup Environment Production

### 12.1 File `.env.prod`
* **Lokasi File**: `~/petualangan-cuan/production/.env.prod`
* **Cara Membuat**:
```bash
nano ~/petualangan-cuan/production/.env.prod
```

```env
# Server Configuration
PORT=8080
TZ=Asia/Jakarta

# Database Configuration (PostgreSQL Docker)
DB_HOST=postgres
DB_USER=cuanpsqlprod
DB_PASSWORD=GANTI_DENGAN_PASSWORD_PROD_YANG_KUAT
DB_NAME=cuan_prod
DB_PORT=5432

# Security & Authentication (Generate via: openssl rand -base64 32)
JWT_SECRET=GANTI_DENGAN_JWT_SECRET_RANDOM_PROD_32_CHAR
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=72h

# Google OAuth 2.0
GOOGLE_CLIENT_ID=GANTI_DENGAN_GOOGLE_CLIENT_ID_PROD.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GANTI_DENGAN_GOOGLE_CLIENT_SECRET_PROD
GOOGLE_REDIRECT_URL=https://petualangancuan.rofid.me/api/auth/google/callback

# Frontend Integration (CORS)
FRONTEND_URL=https://petualangancuan.rofid.me

# AI Configuration (external = hemat RAM, local = on-server GGUF)
AI_PROVIDER=external
COMPOSE_PROFILES=

# External AI Settings (Aktif jika AI_PROVIDER=external)
EXTERNAL_AI_URL=https://openrouter.ai/api/v1/chat/completions
EXTERNAL_AI_API_KEY=GANTI_DENGAN_API_KEY_OPENROUTER_ATAU_GEMINI
EXTERNAL_AI_MODEL=google/gemini-2.0-flash-001

# Local AI Settings (Hanya jika AI_PROVIDER=local dan COMPOSE_PROFILES=local-ai)
LOCAL_LLM_URL=http://llm-server:8080
LOCAL_WHISPER_URL=http://whisper-server:8000

# WhatsApp Gateway
WA_GATEWAY_USERNAME=adminwaprod
WA_GATEWAY_PASSWORD=GANTI_PASSWORD_WA_PROD
WA_WEBHOOK_URL=http://prod_cuan_backend:8080/api/webhook/whatsapp
WHATSAPP_WEBHOOK_SECRET=GANTI_SECRET_WEBHOOK_WA_PROD
WA_GATEWAY_URL=http://prod_cuan_wa_gateway:3000

# Observability (Grafana)
GRAFANA_USER=admingrafanaprod
GRAFANA_PASSWORD=GANTI_PASSWORD_GRAFANA_PROD

# Host Port Mapping (Production)
FRONTEND_PORT=3000
BACKEND_PORT=8080
LLM_PORT=8081
WHISPER_PORT=8000
WA_GATEWAY_PORT=3003
LOKI_PORT=3100
GRAFANA_PORT=3002
UPTIME_KUMA_PORT=3001
```

---

## 13. Setup Environment Staging

### 13.1 File `.env.staging`
* **Lokasi File**: `~/petualangan-cuan/staging/.env.staging`
* **Cara Membuat**:
```bash
nano ~/petualangan-cuan/staging/.env.staging
```

```env
# Server Configuration
PORT=8080
TZ=Asia/Jakarta

# Database Configuration (PostgreSQL Docker - Staging)
DB_HOST=postgres
DB_USER=cuanpsqlstg
DB_PASSWORD=GANTI_DENGAN_PASSWORD_STAGING_YANG_KUAT
DB_NAME=cuan_stg
DB_PORT=5555

# Security & Authentication
JWT_SECRET=GANTI_DENGAN_JWT_SECRET_RANDOM_STAGING_32_CHAR
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=72h

# Google OAuth 2.0 (Staging)
GOOGLE_CLIENT_ID=GANTI_DENGAN_GOOGLE_CLIENT_ID_STAGING.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=GANTI_DENGAN_GOOGLE_CLIENT_SECRET_STAGING
GOOGLE_REDIRECT_URL=https://stagingpetualangancuan.rofid.me/api/auth/google/callback

# Frontend Integration (CORS)
FRONTEND_URL=https://stagingpetualangancuan.rofid.me

# AI Configuration
AI_PROVIDER=external
COMPOSE_PROFILES=

# External AI Settings
EXTERNAL_AI_URL=https://openrouter.ai/api/v1/chat/completions
EXTERNAL_AI_API_KEY=GANTI_DENGAN_API_KEY_OPENROUTER_ATAU_GEMINI
EXTERNAL_AI_MODEL=google/gemini-2.0-flash-001

# Local AI Settings
LOCAL_LLM_URL=http://llm-server:8080
LOCAL_WHISPER_URL=http://whisper-server:8000

# WhatsApp Gateway
WA_GATEWAY_USERNAME=adminwastaging
WA_GATEWAY_PASSWORD=GANTI_PASSWORD_WA_STAGING
WA_WEBHOOK_URL=http://stg_cuan_backend:8080/api/webhook/whatsapp
WHATSAPP_WEBHOOK_SECRET=GANTI_SECRET_WEBHOOK_WA_STAGING
WA_GATEWAY_URL=http://stg_cuan_wa_gateway:3000

# Observability (Grafana Staging)
GRAFANA_USER=admingrafanastaging
GRAFANA_PASSWORD=GANTI_PASSWORD_GRAFANA_STAGING

# Host Port Mapping (Staging Offset)
FRONTEND_PORT=3300
BACKEND_PORT=8888
LLM_PORT=8881
WHISPER_PORT=8800
WA_GATEWAY_PORT=3330
LOKI_PORT=3110
GRAFANA_PORT=3302
UPTIME_KUMA_PORT=3301
```

### 13.2 Salin File Compose & Monitoring dari Laptop ke VPS
* **Lokasi Eksekusi**: Terminal Laptop (di root direktori proyek).

```bash
# 1. Salin Docker Compose
scp docker-compose.prod.yml ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/production/
scp docker-compose.staging.yml ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/staging/

# 2. Salin Konfigurasi Monitoring
scp -r monitoring/* ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/production/monitoring/
scp -r monitoring/* ubuntu@IP_PUBLIC_VPS:~/petualangan-cuan/staging/monitoring/
```

---

## 14. Setup Domain Cloudflare, SSL Origin, & Nginx Reverse Proxy

### 14.1 Konfigurasi DNS & SSL di Cloudflare
1. Buka dashboard **Cloudflare → DNS → Records**:
   * `A` Record `petualangancuan` → `IP_PUBLIC_VPS` (Proxy Status: ☁️ **Proxied / Orange Cloud**)
   * `A` Record `stagingpetualangancuan` → `IP_PUBLIC_VPS` (Proxy Status: ☁️ **Proxied / Orange Cloud**)
   * `A` Record `ssh` → `IP_PUBLIC_VPS` (Proxy Status: 🔘 **DNS Only / Grey Cloud**)
2. Buka **Cloudflare → SSL/TLS → Overview**:
   * Set Encryption Mode: **Full (strict)**

### 14.2 Pasang Cloudflare Origin Certificate di VPS
1. Buat Origin Certificate di **Cloudflare → SSL/TLS → Origin Server → Create Certificate** (Validitas 15 tahun untuk domain `*.rofid.me`, `rofid.me`).
2. Pasang di server VPS:
   * **Lokasi Eksekusi**: Terminal VPS.

```bash
sudo mkdir -p /etc/ssl/cloudflare

# Salin isi Origin Certificate ke origin.crt
sudo nano /etc/ssl/cloudflare/origin.crt

# Salin isi Private Key ke origin.key
sudo nano /etc/ssl/cloudflare/origin.key

# Kunci permission
sudo chmod 644 /etc/ssl/cloudflare/origin.crt
sudo chmod 600 /etc/ssl/cloudflare/origin.key
```

### 14.3 Konfigurasi Nginx Virtual Host Production
* **File**: `/etc/nginx/sites-available/petualangancuan_prod`

```bash
sudo nano /etc/nginx/sites-available/petualangancuan_prod
```

```nginx
# HTTP -> HTTPS Redirect
server {
    listen 80;
    server_name petualangancuan.rofid.me;
    return 301 https://petualangancuan.rofid.me$request_uri;
}

# HTTPS Server (Production)
server {
    listen 443 ssl http2;
    server_name petualangancuan.rofid.me;

    client_max_body_size 25M;

    ssl_certificate     /etc/ssl/cloudflare/origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/origin.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # 1. FRONTEND (Vue SPA - Port 3000)
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 2. BACKEND API (Go Fiber - Port 8080)
    location /api/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 3. AI STREAMING SSE (Disable Buffering for smooth chat stream)
    location /api/ai/chat/stream {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;

        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 4. STATIC UPLOADS (Receipts, Images, Audio)
    location /uploads/ {
        proxy_pass http://127.0.0.1:8080/uploads/;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 14.4 Konfigurasi Nginx Virtual Host Staging
* **File**: `/etc/nginx/sites-available/petualangancuan_staging`

```bash
sudo nano /etc/nginx/sites-available/petualangancuan_staging
```

```nginx
# HTTP -> HTTPS Redirect
server {
    listen 80;
    server_name stagingpetualangancuan.rofid.me;
    return 301 https://stagingpetualangancuan.rofid.me$request_uri;
}

# HTTPS Server (Staging)
server {
    listen 443 ssl http2;
    server_name stagingpetualangancuan.rofid.me;

    client_max_body_size 25M;

    ssl_certificate     /etc/ssl/cloudflare/origin.crt;
    ssl_certificate_key /etc/ssl/cloudflare/origin.key;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;

    # 1. FRONTEND (Vue SPA - Staging Port 3300)
    location / {
        proxy_pass http://127.0.0.1:3300;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 2. BACKEND API (Go Fiber - Staging Port 8888)
    location /api/ {
        proxy_pass http://127.0.0.1:8888;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 3. AI STREAMING SSE
    location /api/ai/chat/stream {
        proxy_pass http://127.0.0.1:8888;
        proxy_http_version 1.1;

        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 4. STATIC UPLOADS (Staging)
    location /uploads/ {
        proxy_pass http://127.0.0.1:8888/uploads/;
        proxy_http_version 1.1;

        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 14.5 Aktifkan Virtual Host & Restart Nginx
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# Aktifkan konfigurasi
sudo ln -sf /etc/nginx/sites-available/petualangancuan_prod /etc/nginx/sites-enabled/
sudo ln -sf /etc/nginx/sites-available/petualangancuan_staging /etc/nginx/sites-enabled/

# Hapus default Nginx page jika ada
sudo rm -f /etc/nginx/sites-enabled/default

# Test sintaks Nginx
sudo nginx -t

# Reload/Restart Nginx
sudo systemctl restart nginx
```

---

## 15. Bootstrapping & Deployment Manual Pertama Kali

Sebelum CI/CD otomatis dijalankan, jalankan container pertama kali di server untuk memastikan semua image dapat di-pull dan service berjalan sempurna:

### 15.1 Start Staging Stack
* **Lokasi Eksekusi**: Terminal VPS.

```bash
cd ~/petualangan-cuan/staging
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging pull
docker compose --env-file .env.staging -f docker-compose.staging.yml -p cuan-staging up -d
```

### 15.2 Start Production Stack
* **Lokasi Eksekusi**: Terminal VPS.

```bash
cd ~/petualangan-cuan/production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod pull
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod up -d
```

### 15.3 Verifikasi Status Container & User Runtime
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Cek semua container aktif
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# 2. Verifikasi user non-root (appuser:appgroup, UID 100, GID 101) di backend container
docker exec -it prod_cuan_backend id appuser
docker exec -it stg_cuan_backend id appuser
```
*Output yang diharapkan:*
```text
uid=100(appuser) gid=101(appgroup)
```

---

## 16. Otomasi CI/CD via GitHub Actions

### 16.1 Konfigurasi GitHub Repository Secrets
Buka **GitHub Repository → Settings → Secrets and variables → Actions → New repository secret**:

| Secret Name | Nilai / Deskripsi |
|---|---|
| `SERVER_IP` | IP Publik VPS Anda |
| `SERVER_USER` | User SSH VPS (`ubuntu`) |
| `SSH_PRIVATE_KEY` | Isi lengkap private key laptop (`~/.ssh/id_ed25519`) |

### 16.2 Alur Eksekusi CI/CD
* **Branch `staging`** → Memicu [.github/workflows/deploy-staging.yml](file:///f:/Project/petualangan-cuan/.github/workflows/deploy-staging.yml):
  1. Build Docker images: Backend, Frontend, WA Gateway dengan tag `:staging`.
  2. Push image ke GHCR (`ghcr.io/alimurrofid/petualangan-cuan/*:staging`).
  3. SSH ke VPS dan melakukan `docker compose pull && docker compose up -d` di folder `staging/`.
* **Branch `main`** → Memicu [.github/workflows/deploy.yml](file:///f:/Project/petualangan-cuan/.github/workflows/deploy.yml):
  1. Build Docker images: Backend, Frontend, WA Gateway dengan tag `:latest`.
  2. Push image ke GHCR (`ghcr.io/alimurrofid/petualangan-cuan/*:latest`).
  3. SSH ke VPS dan melakukan `docker compose pull && docker compose up -d` di folder `production/`.

---

## 17. Health Check & Monitoring

### 17.1 Stack Monitoring yang Tersedia
* **Loki**: Menerima log terpusat dari Promtail.
* **Promtail**: Membaca stream log langsung dari `/var/run/docker.sock` dan docker container runtime.
* **Grafana**: Visualisasi log dan status sistem.
* **Uptime Kuma**: Pengecekan ketersediaan service secara real-time.

### 17.2 Port Akses Monitoring (via SSH Tunneling / Localhost)
Untuk keamanan, akses dashboard Grafana dan Uptime Kuma dianjurkan menggunakan SSH Tunneling dari laptop:

```bash
# Tunnel Grafana Production (3002) dan Uptime Kuma Production (3001) ke Laptop
ssh -L 3002:127.0.0.1:3002 -L 3001:127.0.0.1:3001 petualangancuan-vps
```
Akses di browser laptop:
* Grafana: `http://localhost:3002` (Login sesuai `GRAFANA_USER` dan `GRAFANA_PASSWORD`)
* Uptime Kuma: `http://localhost:3001`

---

## 18. Prosedur Backup & Restore Database

Karena database berjalan di dalam container Docker (`prod_cuan_db` dan `stg_cuan_db`), proses backup dan restore dilakukan menggunakan `docker exec` dan utility `pg_dump` bawaan container.

### 18.1 Backup Database PostgreSQL

#### 🅰️ Backup Production:
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# Buat folder backup
mkdir -p ~/petualangan-cuan/backups

# Eksekusi pg_dump terkompresi
docker exec -t prod_cuan_db pg_dump -U cuanpsqlprod cuan_prod | gzip > ~/petualangan-cuan/backups/backup_prod_$(date +%Y%m%d_%H%M%S).sql.gz
```

#### 🅱️ Backup Staging:
```bash
docker exec -t stg_cuan_db pg_dump -U cuanpsqlstg cuan_stg | gzip > ~/petualangan-cuan/backups/backup_stg_$(date +%Y%m%d_%H%M%S).sql.gz
```

### 18.2 Backup File Uploads & Session WhatsApp
```bash
# Backup uploads
tar -czvf ~/petualangan-cuan/backups/uploads_prod_$(date +%Y%m%d).tar.gz -C ~/petualangan-cuan/production uploads_prod

# Backup WA Gateway data
tar -czvf ~/petualangan-cuan/backups/wa_data_prod_$(date +%Y%m%d).tar.gz -C ~/petualangan-cuan/production wa-gateway-data_prod
```

### 18.3 Restore Database PostgreSQL
* **Tujuan**: Memulihkan database dari file backup `.sql.gz`.
* **Lokasi Eksekusi**: Terminal VPS.

```bash
# 1. Hentikan traffic backend sementara
cd ~/petualangan-cuan/production
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod stop backend

# 2. Restore data ke container database
gunzip -c ~/petualangan-cuan/backups/backup_prod_YYYYMMDD_HHMMSS.sql.gz | docker exec -i prod_cuan_db psql -U cuanpsqlprod -d cuan_prod

# 3. Nyalakan kembali backend
docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod start backend
```

---

## 19. Prosedur Rollback & Disaster Recovery

### 19.1 Penanganan Deployment Gagal
Jika deployment baru menyebabkan error pada backend atau frontend:

1. **Periksa Log Container Terkait**:
   ```bash
   cd ~/petualangan-cuan/production
   docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod logs -f backend
   ```
2. **Rollback Source Code via Git**:
   ```bash
   # Di laptop:
   git revert HEAD
   git push origin main
   ```
   *GitHub Actions akan otomatis me-rebuild commit sebelumnya dan melakukan deploy ulang.*

3. **Rollback Manual Cepat di Server**:
   Jika ingin segera beralih ke image lokal sebelumnya:
   ```bash
   # Cek daftar image yang tersedia
   docker images | grep petualangan-cuan

   # Restart container
   docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod up -d --force-recreate
   ```

---

## 20. Panduan Troubleshooting & FAQ

### 1. Error: `bind: address already in use` (Port 5432)
* **Penyebab**: Terdapat service native PostgreSQL yang berjalan di host Ubuntu.
* **Solusi**:
  ```bash
  sudo systemctl stop postgresql
  sudo systemctl disable postgresql
  ```

### 2. Error: Upload File Struk Gagal / `Permission Denied`
* **Penyebab**: Folder `uploads_prod` atau `uploads_staging` tidak memiliki hak kepemilikan user runtime backend (`appuser`, UID `100`).
* **Solusi**:
  ```bash
  sudo chown -R 100:101 ~/petualangan-cuan/production/uploads_prod
  sudo chmod -R 755 ~/petualangan-cuan/production/uploads_prod
  ```

### 3. Google OAuth Error / Redirect ke Localhost
* **Penyebab**: Variabel `GOOGLE_REDIRECT_URL` atau `FRONTEND_URL` di `.env.prod` masih menggunakan `localhost`, atau container backend belum di-restart setelah file `.env` diubah.
* **Solusi**:
  1. Pastikan isi `GOOGLE_REDIRECT_URL=https://petualangancuan.rofid.me/api/auth/google/callback` di `.env.prod`.
  2. Pastikan Google Cloud Console memiliki Authorized Redirect URI yang persis sama.
  3. Restart container backend:
     ```bash
     docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod restart backend
     ```

### 4. Server Hang / Out of Memory (OOM) saat Menjalankan AI
* **Penyebab**: Mengaktifkan `COMPOSE_PROFILES=local-ai` pada VPS dengan RAM < 8GB.
* **Solusi**:
  Gunakan mode AI Eksternal. Ubah `.env.prod`:
  ```env
  AI_PROVIDER=external
  COMPOSE_PROFILES=
  ```
  Lalu recreate container:
  ```bash
  docker compose --env-file .env.prod -f docker-compose.prod.yml -p cuan-prod up -d --remove-orphans
  ```

### 5. Nilai Variabel `.env` Tidak Ter-update pada Container
* **Penyebab**: `docker compose restart` tidak memuat ulang file `.env` baru ke container yang sedang berjalan.
* **Solusi**:
  Lakukan `down` dan `up -d` secara eksplisit:
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

## 21. Checklist Deployment Final

Gunakan checklist ini sebelum merilis sistem ke pengguna:

- [ ] Hostname alias SSH laptop berfungsi (`ssh petualangancuan-vps`).
- [ ] SSH root login & password authentication telah dinonaktifkan di VPS.
- [ ] UFW aktif dan hanya membuka port 22, 80, dan 443.
- [ ] Cloudflare DNS `A` Record terkonfigurasi (Proxied untuk web, DNS-Only untuk SSH).
- [ ] Cloudflare Origin Certificate terpasang di `/etc/ssl/cloudflare/`.
- [ ] Nginx virtual hosts (`petualangancuan_prod` & `petualangancuan_staging`) aktif dan lulus uji `nginx -t`.
- [ ] File `.env.prod` dan `.env.staging` terisi lengkap dengan credentials aman.
- [ ] Folder `uploads_prod` dan `uploads_staging` memiliki permission `100:101`.
- [ ] GitHub Secrets (`SERVER_IP`, `SERVER_USER`, `SSH_PRIVATE_KEY`) terpasang di repository GitHub.
- [ ] Container Production dan Staging berjalan beriringan tanpa bentrok port host.
- [ ] Endpoint login Google OAuth dan Webhook WhatsApp berhasil diuji via HTTPS domain.
- [ ] Prosedur backup database via `pg_dump` berhasil diverifikasi.

---

**Aplikasi Siap Digunakan:**
* 🚀 **Production**: [https://petualangancuan.rofid.me](https://petualangancuan.rofid.me)
* 🧪 **Staging**: [https://stagingpetualangancuan.rofid.me](https://stagingpetualangancuan.rofid.me)