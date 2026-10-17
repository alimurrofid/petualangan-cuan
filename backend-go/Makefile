run:
	go run cmd/api/main.go

fresh:
	go run cmd/api/main.go -fresh

seed:
	go run cmd/api/main.go -seed

fresh-seed:
	go run cmd/api/main.go -fresh -seed

# --- Verification targets (AI agent & CI self-check) ---

vet:
	go vet ./...

build:
	go build -o bin/api ./cmd/api

test:
	go test ./...

# Jalankan seluruh verifikasi backend selesai task:
check: vet test build

.PHONY: run fresh seed fresh-seed vet build test check
