# Simple Makefile for a Go project

# Build the application
build:
	@echo "Building..."
	@go build -o main.exe backend/cmd/api/main.go

# Run the application
run:
	@echo "Running the backend..."
	@go run backend/cmd/api/main.go
	# Uncomment the following lines if needed for frontend setup
	# @npm install --prefix ./frontend
	# @npm run dev --prefix ./frontend

# Create DB container
docker-run:
	@echo "Starting Docker containers..."
	@docker compose up --build

# Shutdown DB container
docker-down:
	@echo "Stopping Docker containers..."
	@docker compose down

# Clean generated files
clean:
	@echo "Cleaning up..."
	@rm -f main.exe

.PHONY: build clean run docker-run docker-down
