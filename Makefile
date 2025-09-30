.PHONY: generate clean build test

# Generate all code
generate: generate-storage generate-server generate-client

# Generate server code with storage
generate-server:
	@echo "Generating server code..."
	go run cmd/codegen/main.go \
		-type=server \
		-output=./cmd/server \
		-package=main \
		-module=github.com/openchami/inventory \
		-tidy=false

# Generate storage code
generate-storage:
	@echo "Generating storage code..."
	@mkdir -p internal/storage
	go run cmd/codegen/main.go \
		-type=storage \
		-output=./internal/storage \
		-package=storage \
		-module=github.com/openchami/inventory \
		-tidy=false

# Generate client library
generate-client:
	@echo "Generating client library..."
	@mkdir -p pkg/client
	go run cmd/codegen/main.go \
		-type=client \
		-output=./pkg/client \
		-package=client \
		-module=github.com/openchami/inventory \
		-tidy=false

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -f cmd/server/*_generated.go
	rm -f cmd/server/handlers_*.go
	rm -f pkg/client/*.go
	rm -f internal/storage/*_generated.go

# Post-generation cleanup
post-generate:
	@echo "Running post-generation tasks..."
	go mod tidy
	go fmt ./...

# Build everything
build: generate post-generate
	@echo "Building applications..."
	go build -o bin/server ./cmd/server
	go build -o bin/crawler ./cmd/crawler
	go build -o bin/codegen ./cmd/codegen

# Test everything
test: generate post-generate
	go test ./...

# Development workflow
dev: clean generate post-generate build test
	@echo "Development build complete!"

# Install tools
install-tools:
	go install ./cmd/codegen
	go install ./cmd/crawler
