# Gemini API via Cloudflare Worker Proxy

Panduan ringkas reverse proxy untuk menghubungkan **VPS Tencent** ke **Google Gemini API** (OpenAI-compatible) yang diblokir oleh jaringan Tencent/China.

---

## 1. Arsitektur

```text
[VPS Tencent] ──(Bearer PROXY_SECRET)──> [Cloudflare Worker] ──(Bearer GEMINI_API_KEY)──> [Google Gemini API]
```

* **Keamanan**: `GEMINI_API_KEY` disimpan sebagai *Secret* di Cloudflare (VPS tidak memegang API key asli Google).
* **Proteksi**: VPS wajib menyertakan `PROXY_SECRET` sebagai otentikasi Bearer token ke Worker.
* **Staging & Production**: gunakan `PROXY_SECRET` yang **berbeda** untuk masing-masing environment.
* **Catatan**: Cloudflare Worker hanya menjadi jalur penghubung (*bypass geoblock*), batas quota (RPM/RPD) tetap mengikuti akun Google Gemini Anda.

---

## 2. Endpoint & Secrets

| Lingkungan     | Endpoint OpenAI-Compatible                                                                      | Secrets di Worker                                      |
| :------------- | :---------------------------------------------------------------------------------------------- | :----------------------------------------------------- |
| **Staging**    | `https://tencent-gemini-proxy-staging.alimurrofid77.workers.dev/v1beta/openai/chat/completions` | `GEMINI_API_KEY`<br>`PROXY_SECRET` (khusus staging)    |
| **Production** | `https://tencent-gemini-proxy-prod.alimurrofid77.workers.dev/v1beta/openai/chat/completions`    | `GEMINI_API_KEY`<br>`PROXY_SECRET` (khusus production) |

> **Setup Secret:** Di Cloudflare Dashboard → **Worker** → **Settings** → **Variables and Secrets** → Tambahkan variable dengan tipe **Secret**.
>
> Saat menambahkan **`GEMINI_API_KEY`** dan **`PROXY_SECRET`**, pastikan checkbox **Secret** dicentang untuk keduanya.
>
> Gunakan `PROXY_SECRET` yang berbeda antara staging dan production.
>
> Generate secret acak dengan:
>
> ```bash
> openssl rand -base64 32
> ```

---

## 3. Source Code Cloudflare Worker

Gunakan kode ini pada Worker Staging & Production:

```js
const GOOGLE_API = "https://generativelanguage.googleapis.com";

export default {
  async fetch(request, env) {
    if (request.method !== "POST") {
      return new Response("Method Not Allowed", {
        status: 405,
        headers: { Allow: "POST" },
      });
    }

    const incomingUrl = new URL(request.url);
    const targetUrl = GOOGLE_API + incomingUrl.pathname + incomingUrl.search;

    // Validasi secret dari VPS
    const proxyAuth = request.headers.get("Authorization");

    if (proxyAuth !== `Bearer ${env.PROXY_SECRET}`) {
      return new Response("Unauthorized", { status: 401 });
    }

    // Clone headers & ganti auth dengan Gemini API Key
    const headers = new Headers(request.headers);

    headers.delete("Authorization");
    headers.set("Authorization", `Bearer ${env.GEMINI_API_KEY}`);
    headers.set("Host", "generativelanguage.googleapis.com");

    // Forward request ke Google Gemini
    const response = await fetch(targetUrl, {
      method: request.method,
      headers,
      body: request.body,
      redirect: "follow",
    });

    return new Response(response.body, {
      status: response.status,
      statusText: response.statusText,
      headers: response.headers,
    });
  },
};
```

---

## 4. Pengujian (cURL dari VPS)

```bash
curl -i -X POST "https://tencent-gemini-proxy-staging.alimurrofid77.workers.dev/v1beta/openai/chat/completions" \
  -H "Authorization: Bearer YOUR_STAGING_PROXY_SECRET" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3.5-flash-lite",
    "messages": [
      {
        "role": "user",
        "content": "Halo, tes koneksi!"
      }
    ]
  }'
```

Jika berhasil, respons harus berasal dari Google Gemini dan bukan `401 Unauthorized` dari Cloudflare Worker.

---

## 5. Konfigurasi Aplikasi (`backend-go/.env`)

```env
AI_PROVIDER=external

# Staging
EXTERNAL_AI_URL=https://tencent-gemini-proxy-staging.alimurrofid77.workers.dev/v1beta/openai/chat/completions
EXTERNAL_AI_API_KEY=YOUR_STAGING_PROXY_SECRET
EXTERNAL_AI_MODEL=gemini-3.5-flash-lite

# Production (aktifkan saat di server prod)
# EXTERNAL_AI_URL=https://tencent-gemini-proxy-prod.alimurrofid77.workers.dev/v1beta/openai/chat/completions
# EXTERNAL_AI_API_KEY=YOUR_PROD_PROXY_SECRET
# EXTERNAL_AI_MODEL=gemini-3.5-flash-lite
```

> `EXTERNAL_AI_API_KEY` **bukan** `GEMINI_API_KEY`.
>
> Nilainya adalah `PROXY_SECRET` milik environment masing-masing.

---

## 6. Troubleshooting Cepat

| Error                        | Penyebab                                                      | Solusi                                                     |
| :--------------------------- | :------------------------------------------------------------ | :--------------------------------------------------------- |
| **`400 Bad Request`**        | Format request, parameter, atau model tidak sesuai            | Periksa payload, `model`, dan format OpenAI-compatible API |
| **`401 Unauthorized`**       | `PROXY_SECRET` di VPS tidak cocok dengan yang ada di Worker   | Samakan secret antara `.env` VPS dan Cloudflare Secret     |
| **`403 Forbidden`**          | `GEMINI_API_KEY` salah atau project Google AI Studio dibatasi | Periksa kembali API Key Google di Cloudflare Secret        |
| **`429 Too Many Requests`**  | Kuota/rate limit Gemini tercapai                              | Tunggu reset kuota atau sesuaikan frekuensi request        |
| **`405 Method Not Allowed`** | Request menggunakan method selain POST                        | Pastikan request menggunakan `POST`                        |
