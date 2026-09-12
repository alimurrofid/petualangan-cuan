# 🤖 AGENTS.MD - Petualangan Cuan Developer & AI Agent Guide

Selamat datang di repository **Petualangan Cuan**. Dokumen ini adalah panduan standar dan pedoman operasional utama bagi seluruh **AI Coding Agent** dan developer yang berkontribusi pada codebase ini.

---

## 🛡️ 1. Aturan Keamanan & Pengelolaan Rahasia (CRITICAL)

> [!CAUTION]
> ### KEBIJAKAN PRIVASI & FILE .ENV
> 1. **DILARANG KERAS MEMBACA / MENGAKSES / MENCETAK FILE `.env` ASLI** (`.env`, `.env.prod`, `.env.staging`, atau file rahasia lokal lainnya).
> 2. **HANYA GUNAKAN TEMPLATE CONTOH**: Semua referensi variabel lingkungan **HANYA BOLEH** merujuk pada file [.env.example](file:///.env.example) atau `frontend-vue/.env.example`.
> 3. **JANGAN PERNAH MENULISKAN KREDENSIAL/SECRET ASLI** ke dalam source code, file markdown, commit message, atau log output.
> 4. Saat membuat fitur baru yang membutuhkan variabel konfigurasi, daftarkan variabel baru tersebut di `.env.example` beserta komentar penjelas.

---

## 🧹 2. Prinsip Rekayasa Kode (Clean Code & No Legacy Baggage)

> [!IMPORTANT]
> ### TIDAK BUTUH BACKWARD COMPATIBILITY
> - **Hapus Dead Code / Sisa Implementasi Lama**: Jika ada fungsi, entity, interface, komponen, styling, file, atau route lama yang sudah diganti atau tidak digunakan lagi, **WAJIB LANGSUNG DIHAPUS**.
> - **Dilarang Menulis Adaptor Retro / Legacy Shims**: Jangan membuat wrapper atau layer fallback hanya demi mempertahankan kompatibilitas dengan implementasi lama yang sudah tidak dipakai. Kode harus ramping, modern, dan to-the-point.
> - **Cegah Halusinasi AI**: Menyimpan sisa kode lama yang tidak terpakai menyebabkan AI Agent berikutnya berhalusinasi dan mengimpor modul yang salah. Jaga kebersihan codebase setiap saat.

> [!WARNING]
> ### ATURAN KETAT PENANGANAN DATA & LARANGAN FALLBACK `||`
> - **DILARANG MENGGUNAKAN FALLBACK GANDA `||` PADA PROPERTY/KOLOM**:
>   - ❌ **SALAH**: `const name = item.title || item.name || item.description || ''`
>   - ❌ **SALAH**: `const userId = data.user_id || data.userId || data.ownerId`
>   - Menulis fallback properti semacam ini menutupi kesalahan penamaan skema/kontrak API dan memicu AI agent berikutnya mengambil nama kolom yang salah.
> - **FAIL-FAST (BIARKAN ERROR MELEDAK EKSPLISIT)**:
>   - Jika terjadi error logika, schema mismatch, atau request gagal, **biarkan error terjadi secara eksplisit** dan tangani dengan error-handling yang tepat. Jangan ditutupi dengan try-catch kosong atau default value palsu.
> - **STANDAR NILAI KOSONG**:
>   - **Data Layer (API / State / Database / Logic)**: Gunakan **`null`** (atau `undefined` sesuai strict TypeScript type).
>   - **UI / Display Layer (Template / View)**: Jika suatu data bernilai kosong/null, tampilkan tanda strip **`-`** (dash), bukan placeholder karangan.

---

## 🗺️ 3. Arsitektur Sistem & Tech Stack

### 3.1 Gambaran Arsitektur

```text
                            👤 Pengguna
              ┌─────────────────┴─────────────────┐
              ▼                                   ▼
        🌐 Web Browser                     💬 WhatsApp App
      (Vue.js Frontend)                  (WhatsApp Platform)
              │                                   │
              │                                   ▼
              │                         🚪 WA Gateway
              │                         (Whatsmeow + SQLite)
              │                                   │
              └─────────────────┬─────────────────┘
                                ▼
                       🚀 Go Backend API 
                    (Fiber + GORM + ZeroLog)
                                │
          ┌─────────────────────┼─────────────────────┐
          ▼                     ▼                     ▼
    🐘 PostgreSQL        🧠 AI Service        🔍 Observability Stack
      (Database)      (Local LLM/Whisper             (PLG)
                      atau External AI)               │
                                                      ▼
                                                 📄 Promtail
                                                      │
                                                      ▼
                                                   📚 Loki
                                                      │
                                                      ▼
                                                 📊 Grafana
```

### 3.2 Detail Komponen

| Komponen | Teknologi | Deskripsi & Tanggung Jawab |
|---|---|---|
| **Backend API** | Go 1.21+, Fiber v2, GORM, Zerolog, Swagger | REST API melayani autentikasi, transaksi, dompet, wishlist, hutang-piutang, analitik finansial, serta orkestrasi AI. |
| **Frontend Web** | Vue 3 (`<script setup>`), TypeScript, Vite, Tailwind CSS v4, shadcn-vue, Pinia, ApexCharts, pnpm | SPA interaktif, dashboard finansial, chat advisor, analitik grafik, dan manajemen aset. |
| **WhatsApp Gateway** | Go, Whatsmeow, SQLite | Jembatan penerima pesan WhatsApp yang mem-forward webhook ke Backend API untuk auto-entry transaksi. |
| **AI Subsystem** | Llama.cpp (Local GGUF), Faster-Whisper, External AI Provider Strategy (Gemini, OpenRouter, etc.) | Parsing intent transaksi via teks/suara/gambar (receipt OCR) dan financial advisor. |
| **Database** | PostgreSQL 15 | Penyimpanan data relasional utama. |
| **Observability** | Promtail, Loki, Grafana, Uptime Kuma | Structured logging JSON (`X-Request-Id` tracing), monitoring metrik, dan uptime tracking. |
| **Containerization**| Docker & Docker Compose | Standardisasi runtime lokal, staging, dan produksi. |

---

## 🗂️ 4. Struktur Direktori Proyek

```text
petualangan-cuan/
├── .env.example              # Template variabel lingkungan utama
├── AGENTS.MD                 # Panduan agen AI & developer (file ini)
├── README.md                 # Dokumentasi umum repositori
├── SETUPSERVER.md            # Panduan deployment VPS & CI/CD
├── docker-compose.yml        # Orchestration Docker lokal
├── docker-compose.staging.yml# Orchestration staging
├── docker-compose.prod.yml   # Orchestration production
│
├── backend-go/               # 🚀 Backend Core Service (Go)
│   ├── cmd/api/main.go       # Entry point utama aplikasi
│   ├── docs/                 # Swagger auto-generated documentation
│   ├── internal/
│   │   ├── config/           # Konfigurasi App, Database, AI, JWT, OAuth
│   │   ├── entity/           # GORM Data Models & Entity structs
│   │   ├── handler/          # HTTP Request Handlers (Fiber controllers)
│   │   ├── provider/ai/      # AI Provider (Local LLM, Whisper, External AI)
│   │   ├── repository/       # Data Access Object (DAO / DB queries)
│   │   ├── seeder/           # Seeder database awal
│   │   └── service/          # Business logic utama
│   ├── pkg/
│   │   ├── middleware/       # JWT Auth, Logger, RequestID middleware
│   │   └── utils/            # Helper fungsi, response formatter, encryption
│   ├── Dockerfile
│   ├── Makefile              # Shortcut perintah Go
│   └── go.mod / go.sum
│
├── frontend-vue/             # 🌐 Frontend Web Client (Vue 3)
│   ├── src/
│   │   ├── assets/           # Asset statis, CSS, ikon
│   │   ├── components/       # Reusable UI components & shadcn-vue parts
│   │   │   └── ui/           # Atom komponen UI (Button, Dialog, Card, dsb.)
│   │   ├── composables/      # Vue Composition API hooks
│   │   ├── layouts/          # Layout shell (Sidebar, Topbar, Auth layout)
│   │   ├── lib/              # Utils, formatting, Axios client interceptors
│   │   ├── router/           # Vue Router guards & route definitions
│   │   ├── stores/           # Pinia stores (Auth, Transaction, Wallet, dsb.)
│   │   ├── types/            # TypeScript interfaces & type definitions
│   │   └── views/            # Halaman utama aplikasi (Dashboard, Auth, dsb.)
│   ├── .env.example          # Template konfigurasi frontend (VITE_API_BASE_URL)
│   ├── package.json          # Dependency pnpm
│   ├── vite.config.ts        # Konfigurasi Vite & Tailwind v4
│   └── tsconfig.json
│
├── wa-gateway/               # 🚪 WhatsApp Integration Service
│   ├── src/                  # Kode Go Whatsmeow gateway
│   └── docker/
│
├── ai-models/                # 🧠 Model Weights & Cache
│   ├── google_gemma-3-4b-it-Q4_K_M.gguf
│   └── mmproj-google_gemma-3-4b-it-f16.gguf
│
└── monitoring/               # 🔍 Observability Configurations
    ├── grafana/provisioning/ # Dashboard & datasource pre-provisioning
    └── promtail-config.yml   # Konfigurasi scrape Docker logs ke Loki
```

---

## ⚡ 5. Perintah & Workflow Pengembangan (Commands)

### 5.1 Menjalankan dengan Docker Compose (Rekomendasi Utama)
```bash
# Menjalankan seluruh stack secara lokal
docker compose up -d --build

# Melihat log spesifik service
docker compose logs -f backend
docker compose logs -f wa-gateway

# Menghentikan stack
docker compose down
```

### 5.2 Backend Go (Manual / Native Development)
Masuk ke direktori `backend-go/`:
```bash
# Mengunduh dependensi
go mod tidy

# Menjalankan server backend
go run cmd/api/main.go
# atau menggunakan Makefile:
make run

# Reset database & migrate ulang:
make fresh

# Menjalankan seeder:
make seed

# Reset database, migrasi, dan seed sekaligus:
make fresh-seed
```

### 5.3 Frontend Vue (pnpm Development)
Masuk ke direktori `frontend-vue/`:
```bash
# Install dependencies (wajib pakai pnpm)
pnpm install

# Menjalankan Vite dev server (HMR di http://localhost:5173)
pnpm run dev

# Type-check dan build bundle produksi
pnpm run build

# Preview build produksi
pnpm run preview
```

---

## 📐 6. Konvensi Koding & Standar Implementasi

### 6.1 Backend Go Guidelines
1. **Pola Layer Terpisah**:
   - `Handler`: Hanya bertugas memvalidasi input HTTP, mem-parse body/query/param, memanggil Service, dan mengembalikan format JSON standar.
   - `Service`: Berisi seluruh aturan bisnis, validasi logika, orkestrasi transaksi DB, dan pemanggilan provider AI.
   - `Repository`: Hanya berisi query database GORM langsung.
   - `Entity`: Definisi struct model GORM yang bersih.
2. **Error Handling & Response Format**:
   - Selalu periksa error secara eksplisit (`if err != nil`).
   - Gunakan format response JSON yang konsisten melalui helper utility.
3. **Structured Logging & Tracing**:
   - Gunakan `zerolog` dengan menyertakan context request (`request_id`).
   - Pastikan log tidak mencetak data sensitif seperti password plain-text atau secret token.

### 6.2 Frontend Vue 3 Guidelines
1. **Composition API & `<script setup lang="ts">`**:
   - Seluruh komponen baru wajib menggunakan Composition API dengan `<script setup lang="ts">`.
2. **Strict Typing (TypeScript)**:
   - Hindari penggunaan tipe `any`. Definisikan interface data secara akurat di folder `types/` atau pada store terkait.
3. **State Management (Pinia)**:
   - Kelola data global dan pemanggilan API dalam Pinia Store yang terisolasi per domain (misal `auth.ts`, `transaction.ts`, `wallet.ts`).
4. **Styling (Tailwind CSS v4)**:
   - Gunakan class utility Tailwind CSS v4 secara konsisten.
   - Manfaatkan komponen reusable berbasis `shadcn-vue` yang sudah ada di `src/components/ui/`.
5. **UI Handling State**:
   - Setiap pemanggilan data harus memiliki visual feedback yang jelas (Loading skeleton/spinner, Empty state dengan tanda `-`, dan Error toast/alert).

---

## ✅ 7. SOP & Checklist AI Agent Sebelum Menyelesaikan Task

Sebelum menyelesaikan pekerjaan (menyelesaikan prompt user), setiap AI Agent **wajib** memeriksa:

1. [ ] **Apakah ada dead code sisa?**
   - Hapus function, variable, import, atau file yang sudah tidak terpakai lagi. Jangan tinggalkan sisa implementasi usang.
2. [ ] **Apakah ada fallback properti sembrono (`||`)?**
   - Pastikan tidak ada kode seperti `a.field1 || a.field2`. Jika data kosong, pastikan menggunakan `null` di state/data dan `-` di tampilan.
3. [ ] **Apakah ada kebocoran file `.env`?**
   - Pastikan file `.env` asli tidak tersentuh/terbaca dan konfigurasi hanya mengacu pada `.env.example`.
4. [ ] **Apakah kode lolos validasi tipe / linting?**
   - Untuk backend: pastikan kode Go valid (`go vet`, `go build`).
   - Untuk frontend: pastikan tipe TypeScript valid (`pnpm run build` / `vue-tsc`).
5. [ ] **Apakah arsitektur tetap terjaga?**
   - Jangan bypass layer (misal Handler memanggil Repository langsung tanpa melalui Service).
