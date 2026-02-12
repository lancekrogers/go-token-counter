set dotenv-load := false

mod build '.justfiles/build.just'
mod test '.justfiles/test.just'

binary := "tcount"
bin_dir := "bin"

# Show available recipes
@default:
    just --list

# Build the binary
build: fmt vet
    go build -o {{bin_dir}}/{{binary}} ./cmd/tcount

# Build only (no fmt/vet)
build-only:
    go build -o {{bin_dir}}/{{binary}} ./cmd/tcount

# Format code
fmt:
    go fmt ./...

# Run go vet
vet:
    go vet ./...

# Run linter
lint:
    golangci-lint run ./...

# Clean build artifacts
clean:
    rm -rf {{bin_dir}}

# Download dependencies
deps:
    go mod tidy
    go mod download

# Install binary to GOPATH/bin
install: build
    go install ./cmd/tcount

# Uninstall binary
uninstall:
    rm -f $(go env GOPATH)/bin/{{binary}}
