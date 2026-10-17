# 🚪 AGENTS.md — WA Gateway (Go / Whatsmeow)

Area-specific guide for `wa-gateway/`. The global rules in the root [AGENTS.md](../AGENTS.md) still apply.

## ⚡ 1. Commands & Verification

```bash
cd wa-gateway/src

# Build & run
go build -o whatsapp.exe .     # Windows (Linux/macOS: -o whatsapp)
go run . rest                  # REST API mode
go run . mcp                   # MCP server mode

# Dependencies & hygiene
go mod tidy
go fmt ./...

# --- Verification (MUST pass before finishing a task) ---
go vet ./...
go test ./...
```

## 🏗️ 2. Architecture

- **Domain-Driven Design**: business logic is split into domain packages under `domains/` (`app`, `chat`, `group`, `message`, `send`, `user`, `newsletter`).
- **Clean Architecture**: clear separation between UI, use cases, and infrastructure.
- **Cobra CLI**: command pattern with separate commands for the `rest` and `mcp` modes.

| Directory | Contents |
|---|---|
| `src/cmd/` | CLI commands (root, rest, mcp) |
| `src/domains/` | Business domain logic |
| `src/usecase/` | Application use cases bridging domains and UI |
| `src/infrastructure/` | External integrations (WhatsApp, database) |
| `src/ui/` | Interface layers (REST API, MCP server, WebSocket) |
| `src/validations/` | Input validation logic |
| `src/pkg/` | Shared utilities and helpers |

**Modes**: REST mode = Fiber web server with HTML templates, WebSocket, middleware stack. MCP mode = Model Context Protocol server with SSE transport for AI agent integration. Both modes **cannot run simultaneously** (whatsmeow limitation) — never design a feature that assumes both are active at once.

## 📦 3. Key Dependencies

- `go.mau.fi/whatsmeow` — WhatsApp Web protocol implementation
- `github.com/gofiber/fiber/v2` — web framework for REST mode
- `github.com/mark3labs/mcp-go` — MCP server implementation
- `github.com/spf13/cobra` — CLI framework
- `github.com/spf13/viper` — configuration management

Go version: **1.26** (see `src/go.mod`). FFmpeg is required for media processing. HTML templates and assets are embedded into the binary via Go's embed feature.

## ⚠️ 4. Critical Rule: Device ID vs JID

Two distinct identifiers, NEVER confused:

- **Device ID** (`devices.device_id`): user-assigned alias (e.g. `"busine"`) or auto-generated UUID — key in `DeviceManager.devices`, returned by `DeviceInstance.ID()` and `ResolveDevice()`.
- **JID** (`devices.jid`): full WhatsApp JID (`62896...@s.whatsapp.net`) — derived from `client.Store.ID.ToNonAD().String()`, returned by `DeviceInstance.JID()`.

The `chats`/`messages` tables store `device_id` as the **JID**, not the alias. When querying chat storage always use `DeviceInstance.JID()` (falling back to `ID()` when the JID is empty) — the `deviceChatStorage` wrapper (`src/infrastructure/whatsapp/chatstorage_wrapper.go`) handles this automatically in `loadFromRegistry()` / `ensureInstance()`; do not bypass it.

Key files: `src/infrastructure/whatsapp/device_instance.go`, `device_manager.go`, `chatstorage_wrapper.go`.

## 🔧 5. Configuration & Storage

- Configuration priority: **CLI flags > environment variables > `.env`**. Every env var can be overridden by a CLI flag.
- Env template: [wa-gateway/src/.env.example](src/.env.example) only. Reading the real `.env` is forbidden (global rule).
- **Main DB** (`DB_URI`): WhatsApp connection data (SQLite by default, PostgreSQL supported).
- **Chat storage**: separate SQLite database (`storages/chatstorage.db`), enabled via `WHATSAPP_CHAT_STORAGE=true`.
- **Keys DB** (`DB_KEYS_URI`): in-memory SQLite for encryption keys.
- Media files live in `src/statics/media/`, databases in `src/storages/`.
- Multi-device: multiple WhatsApp accounts can be connected in one server instance, with per-device auto-reconnect and connection monitoring.

## 📌 6. Other Rules

- All source code lives in `src/` — no Go files outside it.
- Media processing (compression, webhook) is documented in `docs/` (`webhook-payload.md`, `chatwoot.md`).
