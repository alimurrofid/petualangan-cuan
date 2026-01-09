### 1. Diagram Arsitektur (Micro-Services Lokal)

Semua layanan berjalan dalam satu jaringan Docker (`docker-compose`), saling berkomunikasi via HTTP internal, kecuali Whisper yang dijalankan via Command Line (CLI) oleh Go.

```mermaid
graph TD
    User[User (Web/WA)] -->|Request| WA[WA Gateway Service]
    User -->|Request| Web[Frontend Vue]
    
    subgraph "Docker Host (Server)"
        WA -->|Webhook HTTP| Backend[Backend Go (Main App)]
        Web -->|API HTTP| Backend
        
        Backend -->|Exec CLI| Whisper[Whisper.cpp Binary]
        Backend -->|HTTP POST| OCR[Python OCR Service]
        Backend -->|HTTP POST| LLM[Llama Server (Qwen 0.5B)]
        
        Backend -->|SQL| DB[(PostgreSQL)]
    end

```

---

### 2. Rincian Komponen

| Komponen | Teknologi / Tool | Peran & Tugas | Konsumsi Resource (Estimasi) |
| --- | --- | --- | --- |
| **Main Backend** | **Go (Fiber)** | Mengatur logika bisnis, menyimpan ke DB, menerima Webhook WA, dan menjadi "orkestrator" AI.
| **Logic Engine** | **Llama.cpp Server** | Menjalankan model **Qwen 2.5 (3B)**. Menerima teks mentah, mengubahnya menjadi JSON terstruktur.
| **Voice Engine** | **Whisper.cpp** | Binary C++ yang dipanggil Go. Mengubah Audio (.wav) menjadi Teks.
| **Vision Engine** | **Python (FastAPI)** | Microservice kecil menjalankan **PaddleOCR**. Menerima Gambar, mengembalikan Teks mentah.
| **Chat Gateway** | **Go-WA-Multidevice** | Menangani koneksi WhatsApp. Meneruskan pesan user ke Backend via Webhook.

---

### 3. Alur Kerja (Data Flow)

#### A. Skenario: User Kirim Voice Note di WA

1. **User:** Kirim VN: *"Habis beli bensin 50 ribu"* (Format OGG).
2. **WA Gateway:** Terima pesan  Download file  Kirim JSON Payload ke Webhook Backend Go.
3. **Backend Go:**
* Terima webhook.
* Convert OGG ke WAV menggunakan `ffmpeg`.
* Jalankan perintah: `./bin/whisper -f file.wav`.
* Dapat output teks: *"Habis beli bensin 50 ribu"*.


4. **Backend Go:** Kirim teks ke **Llama Server** dengan prompt: *"Ekstrak ke JSON"*.
5. **Llama Server:** Balas: `{"item": "Bensin", "amount": 50000, "category": "Transport"}`.
6. **Backend Go:** Simpan ke DB  Kirim balasan *"Transaksi Bensin 50.000 tercatat!"* ke WA Gateway.

#### B. Skenario: User Upload Foto Struk

1. **User:** Upload foto struk via Web/WA.
2. **Backend Go:** Kirim file gambar ke `http://ocr-service:8000/scan`.
3. **OCR Service:** Baca gambar  Balas teks mentah struk.
4. **Backend Go:** Kirim teks struk ke **Llama Server**.
5. **Llama Server:** Analisis teks struk  Ekstrak Total & Nama Toko  Balas JSON.
6. **Backend Go:** Simpan transaksi.

---

### 4. Struktur Folder Proyek

Ini menggabungkan kode `petualangan-cuan`, `go-whatsapp`, dan service baru.

```text
/petualangan-cuan
├── docker-compose.yml          # Mengatur 5 Service (Backend, DB, LLM, OCR, WA)
├── ai-models/                  # Folder host untuk menyimpan file .gguf
│   └── qwen2.5-3b-instruct.gguf
│
├── backend-go/
│   ├── Dockerfile
│   ├── bin/                    # Binary Whisper & Modelnya
│   │   ├── main                # whisper executable
│   │   └── ggml-base.bin       # whisper model
│   ├── internal/
│   │   ├── handler/
│   │   └── service/
│   └── ...
├── frontend-vue/               # Frontend Vue.js
│   ├── Dockerfile
│   ├── public/
│   ├── src/
│   └── ...
├── ocr-service/                # OCR Service
│   ├── Dockerfile
│   ├── main.py                 # FastAPI app
│   └── requirements.txt
│
└── wa-gateway/                 # WA Gateway Service
    ├── Dockerfile
    └── ...

```

---