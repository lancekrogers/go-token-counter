set working-directory := '..'

binary := "tcount"
bin_dir := "bin"

# Show available build recipes
@default:
    just --list --justfile {{source_file()}}

# Build for Linux amd64
linux:
    GOOS=linux GOARCH=amd64 go build -o {{bin_dir}}/{{binary}}-linux-amd64 ./cmd/tcount

# Build for macOS amd64
darwin:
    GOOS=darwin GOARCH=amd64 go build -o {{bin_dir}}/{{binary}}-darwin-amd64 ./cmd/tcount

# Build for macOS arm64
darwin-arm64:
    GOOS=darwin GOARCH=arm64 go build -o {{bin_dir}}/{{binary}}-darwin-arm64 ./cmd/tcount

# Build for Windows
windows:
    GOOS=windows GOARCH=amd64 go build -o {{bin_dir}}/{{binary}}-windows-amd64.exe ./cmd/tcount

# Build all platforms
all: linux darwin darwin-arm64 windows
