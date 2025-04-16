.PHONY: build-frontend build-backend clean dev-frontend dev-backend dev static-build

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
	cd backend && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ../app
	@echo "Static binary built successfully: ./app"

# Default target
all: build-backend 