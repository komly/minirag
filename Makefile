.PHONY: build-frontend build-backend clean dev-frontend dev-backend dev static-build static-build-all

# Build the frontend React app
build-frontend:
	cd backend/cmd/minirag/frontend && pnpm install && pnpm run build

# Build the backend with embedded frontend
build-backend: build-frontend
	cd backend && go build -o ../app

# Clean build artifacts
clean:
	rm -rf backend/cmd/minirag/frontend/dist
	rm -f app

# Development mode - run frontend in watch mode
dev-frontend:
	@echo "Starting frontend dev server..."
	cd backend/cmd/minirag/frontend && pnpm run dev

# Development mode - run backend server
dev-backend:
	@echo "Waiting for frontend server to be ready..."
	@while ! curl -s http://localhost:5173 > /dev/null; do sleep 1; done
	@echo "Frontend server is ready, starting backend..."
	cd backend/cmd/minirag && go run main.go -dev

# Development mode - run both frontend and backend
dev:
	@echo "Starting development mode..."
	@echo "Frontend will be available at http://localhost:5173"
	@echo "Backend will be available at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@echo ""
	@echo "Starting frontend and backend in parallel..."
	@make -j 2 dev-frontend dev-backend

# Build a static binary with embedded frontend
static-build: build-frontend
	@echo "Building static binary..."
	mkdir -p bin
	cd backend/cmd/minirag && go build -ldflags="-w -s" -o ../../../bin/minirag
	@echo "Static binary built successfully: ./bin/minirag"

# Build static binaries for multiple platforms (macOS, Linux, Windows)
static-build-all: build-frontend
	@echo "Building static binary for Linux (amd64)..."
	mkdir -p bin
	cd backend/cmd/minirag && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ../../../bin/minirag-linux-amd64
	@echo "Building static binary for macOS (amd64)..."
	cd backend/cmd/minirag && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-w -s" -o ../../../bin/minirag-darwin-amd64
	@echo "Building static binary for macOS (arm64)..."
	cd backend/cmd/minirag && CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o ../../../bin/minirag-darwin-arm64
	@echo "Building static binary for Windows (amd64)..."
	cd backend/cmd/minirag && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-w -s" -o ../../../bin/minirag-windows-amd64.exe
	@echo "Static binaries built successfully in ./bin: minirag-linux-amd64, minirag-darwin-amd64, minirag-darwin-arm64, minirag-windows-amd64.exe"

# Default target
all: build-backend 