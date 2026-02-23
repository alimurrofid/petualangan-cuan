<div align="center">
  <h1>🌟 Petualangan Cuan 🌟</h1>
  <p><i>An intelligent personal finance and wealth-tracking application powered by Local AI and WhatsApp integration</i></p>

  <!-- Badges -->
  <p>
    <img src="https://img.shields.io/badge/Vue.js-35495E?style=for-the-badge&logo=vuedotjs&logoColor=4FC08D" alt="Vue.js">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
    <img src="https://img.shields.io/badge/PostgreSQL-316192?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
    <img src="https://img.shields.io/badge/Docker-2CA5E0?style=for-the-badge&logo=docker&logoColor=white" alt="Docker">
    <img src="https://img.shields.io/badge/Grafana-F46800?style=for-the-badge&logo=grafana&logoColor=white" alt="Grafana">
    <img src="https://img.shields.io/badge/WhatsApp-25D366?style=for-the-badge&logo=whatsapp&logoColor=white" alt="WhatsApp">
  </p>
</div>

---

## 🚀 Key Features

### 👤 User Features
- **Transaction Tracking:** Record and categorize daily income and expenses effortlessly.
- **Wallet Management:** Keep track of balances across multiple wallets/accounts.
- **Saving Goals:** Set financial aspirations and monitor progress.
- **Debt Management:** Track lending and borrowing ledgers natively.
- **Wishlist:** Compile a list of desired items and compare prices.
- **Financial Health Metrics:** Get summaries and visualized financial health insights.
- **AI Financial Advisor:** An intelligent chatbot accessible via the Web Interface and WhatsApp, providing contextual voice, text, and image receipt analysis for automated transaction entry.

### ⚙️ Technical / DevOps Features
- **Microservices Architecture:** Independently scalable backend API, frontend SPA, and WhatsApp messaging gateway.
- **Local AI Infrastructure:** Integrated local LLM (`google_gemma-3-4b-it` via `llama.cpp`) and local Whisper model (`faster-whisper-server`) for private, offline data processing.
- **Distributed Tracing:** Seamless request correlation (`X-Request-Id`) across handlers, services, and middleware.
- **PLG Monitoring Stack:** Complete self-hosted observability using **Promtail, Loki, and Grafana** paired with robust structured JSON logging via `zerolog`.
- **Uptime Monitoring:** Real-time service uptime tracking via Uptime Kuma.

---

## 🛠️ Tech Stack & Architecture

### 🗺️ System Architecture

```text
                            👤 Users
              ┌────────────────┴────────────────┐
              ▼                                 ▼
       🌐 Web Browser                    💬 WhatsApp App
     (Vue.js Frontend)                 (WhatsApp Platform)
              │                                 │
              │                                 ▼
              │                        🚪 WA Gateway
              │                        (Whatsmeow + SQLite)
              │                                 │
              └────────────────┬────────────────┘
                               ▼
                      🚀 Go Backend API 
                   (Fiber + GORM + ZeroLog)
                               │
         ┌─────────────────────┼─────────────────────┐
         ▼                     ▼                     ▼
   🐘 PostgreSQL       🧠 Local AI Server    🔍 Observability Stack
     (Database)        (Llama.cpp/Whisper)           │
                                                     ▼
                                                📄 Promtail
                                                     │
                                                     ▼
                                                  📚 Loki
                                                     │
                                                     ▼
                                                📊 Grafana
```

### 🧰 Technology Details

| Component | Technology | Description |
|-----------|------------|-------------|
| **Frontend** | Vue.js 3, TypeScript, Vite, Pnpm, TailwindCSS | High-performance, responsive single-page application. |
| **Backend API** | Go, Fiber, GORM, PostgreSQL, ZeroLog | Lightning-fast REST API orchestrating business logic and database operations. |
| **WhatsApp Gateway**| Go, Whatsmeow, SQLite | Webhook-driven bot acting as a bridge between user WhatsApp messages and the Backend AI. |
| **AI Infrastructure**| Llama.cpp, Faster-Whisper | Localized, privacy-first intent parsing and voice transcription. (Gemma model configurable). |
| **Monitoring** | Promtail, Loki, Grafana, Uptime Kuma | Full-stack log aggregation, querying, and system health visualization. |
| **Deployment** | Docker & Docker Compose | Containerized environments for reproducible, cross-platform local deployment. |

---

## 🗂️ Repository Structure

```text
/petualangan-cuan
├── backend-go/          # Go Fiber REST API, strict MVC/Service layers
├── frontend-vue/        # Vue 3 SPA frontend (Vite)
├── wa-gateway/          # Go-based whatsapp integration via Whatsmeow
├── ai-models/           # Quantized GGUF models and Whisper assets
├── monitoring/          # Promtail configs and Grafana provisioning
├── uploads/             # Persistent volume for transaction receipts/audio
└── docker-compose.yml   # Infrastructure orchestration script
```

---

## 💻 Getting Started (Local Development)

### 1. Prerequisites
Ensure you have the following installed on your host machine:
- **[Docker Engine & Docker Compose](https://docs.docker.com/get-docker/)**
- **[Go 1.22+](https://go.dev/dl/)** (Optional, for native backend running)
- **[Node.js 18+](https://nodejs.org/) & [Pnpm](https://pnpm.io/)** (Optional, for frontend hacking)

### 2. Environment Variables
You need to configure the environment files before starting the containers.

```bash
# In the root, copy the master environment template
cp .env.example .env

# Navigate into the frontend and copy its environment template
cd frontend-vue
cp .env.example .env

# Navigate into the wa-gateway and copy its environment template
cd ../wa-gateway
cp .env.example .env
```
*(Ensure you modify the generated `.env` files with secure secrets according to your local setup).*

### 3. Running the Stack
Launch the entire infrastructure using Docker Compose from the root directory:

```bash
# Build and start all containers in detached mode
docker compose up -d --build
```

### 4. Access Points
Once the stack is spun up, the following services will be available on your localhost:

| Service | Protocol | URL / Port | Credentials (If Auth Required) |
|---------|----------|------------|--------------------------------|
| **Frontend UI** | HTTP | `http://localhost:5173` | *(Public SPA)* |
| **Backend API** | HTTP | `http://localhost:8080/api` | *(Token Based)* |
| **WA Gateway** | HTTP | `http://localhost:3000` | Basic Auth in `.env` |
| **Grafana** | HTTP | `http://localhost:3002` | `admin` / `admincuan` |
| **Uptime Kuma** | HTTP | `http://localhost:3001` | *Create on first visit* |
| **Local LLM** | HTTP | `http://localhost:8081` | *(Internal API)* |
| **Whisper Server**| HTTP | `http://localhost:8000` | *(Internal API)* |

---

## 📊 Observability & Monitoring

Petualangan Cuan takes observability seriously by employing the **Grafana, Loki, Promtail (PLG)** stack directly alongside the application. 

- **Accessing Logs:** Open Grafana at [http://localhost:3002](http://localhost:3002) (Login: `admin` / `admincuan`). A pre-provisioned dashboard *"Petualangan Cuan Logs Monitoring"* is automatically imported on startup.
- **Traceability:** The Go backend leverages strict `zerolog` instrumentation. Every incoming HTTP request is assigned a unique `X-Request-Id`, which is propagated down into the service wrappers.
- **Debugging Example:** 
  If a user reports an error at 14:00, you can query Grafana Loki to trace the exact lifecycle:
  ```logql
  {container="petualangan_cuan_backend"} | json | request_id="a1b2c3d4-e5f6-g7h8..."
  ```
  This immediately isolates the exact SQL syntax error, AI timeout, or warning without polluting context.

---

## 📜 API Documentation

The REST API strictly adheres to Swagger Open API specifications.

Once the backend container is running, access the interactive auto-generated documentation via:
* **Swagger UI:** [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

This interface can be used to manually invoke protected endpoints and inspect required JSON payloads.
