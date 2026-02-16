set dotenv-load := false

mod release '.justfiles/build.just'
mod test '.justfiles/test.just'

binary := "tcount"
bin_dir := "bin"
BUILDTOOL := "go run ./internal/buildutil"

# Show available recipes
@default:
    just --list

# Build the binary (with dashboard)
build:
    @{{BUILDTOOL}} build

# Build only (no vet, with dashboard)
build-only:
    @{{BUILDTOOL}} build-only

# Format code
fmt:
    go fmt ./...

# Run go vet
vet:
    go vet ./...

# Run linter
lint:
    golangci-lint run ./...

# Clean build artifacts (with dashboard)
clean:
    @{{BUILDTOOL}} clean

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
