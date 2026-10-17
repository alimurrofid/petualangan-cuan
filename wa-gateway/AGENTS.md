# 🚪 AGENTS.md — WA Gateway (Go / Whatsmeow)

Area-specific guide for `wa-gateway/`. The complete architecture & commands already exist in [`wa-gateway/CLAUDE.md`](CLAUDE.md) — **read that file as the primary guide for this area** (equivalent to an AGENTS.md: commands, DDD architecture, databases, key dependencies).

## ⚡ 1. Mandatory Verification

```bash
cd wa-gateway/src
go vet ./...
go test ./...
```

Both MUST pass before finishing a task in this area.

## ⚠️ 2. Critical Rule: Device ID vs JID

Two distinct identifiers, NEVER confused (details in CLAUDE.md §Device ID vs JID):

- **Device ID** (`devices.device_id`): user-assigned alias (e.g. `"busine"`) or auto-generated UUID — key in `DeviceManager.devices`, returned by `ResolveDevice()`.
- **JID** (`devices.jid`): full WhatsApp JID (`62896...@s.whatsapp.net`) — returned by `DeviceInstance.JID()`.
- The `chats`/`messages` tables store `device_id` as the **JID**, not the alias. When querying chat storage always use `DeviceInstance.JID()` (falling back to `ID()` when the JID is empty) — the `deviceChatStorage` wrapper (`src/infrastructure/whatsapp/chatstorage_wrapper.go`) handles this automatically; do not bypass it.

## 📌 3. Other Rules

- REST mode & MCP mode **cannot run simultaneously** (whatsmeow limitation) — never design a feature that assumes both are active at once.
- All source lives in `src/`; media in `src/statics/media/`, databases in `src/storages/`.
- Configuration priority: CLI flags > env vars > `.env`; template is only at [wa-gateway/src/.env.example](src/.env.example). Reading the real `.env` is forbidden (global rule).
