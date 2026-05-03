.PHONY: init run lint test build build-web build-all clean version

MODULE := github.com/skylunna/luner
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

ifeq ($(OS),Windows_NT)
    # Windows 环境
    EXT := .exe
    RM := powershell -NoProfile -Command "Remove-Item -Force -Recurse -ErrorAction SilentlyContinue"
else
    # Linux / macOS
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
        EXT :=
    else
        EXT :=
    endif
    RM := rm -rf
endif

build-web:
	@echo "→ Building web console..."
	cd web && npm install --silent && npm run build
	@echo "✓ Web console built → internal/console/dist/"

build: build-web
	go build -ldflags="-s -w \
		-X $(MODULE)/cmd/luner.Version=$(VERSION) \
		-X $(MODULE)/cmd/luner.commit=$(COMMIT) \
		-X $(MODULE)/cmd/luner.buildDate=$(DATE)" \
		-o bin/luner$(EXT) ./cmd/luner
	@echo "✓ Built bin/luner$(EXT) (version=$(VERSION) commit=$(COMMIT))"

build-all: build

run:
	go run ./cmd/luner -config config/config.yaml

version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

clean:
	$(RM) bin/
	@echo " Cleaned bin/"

init:
	@echo "🔧 Installing dev tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo " Done"

test:
	go test -race -v ./...

lint:
	golangci-lint run ./...