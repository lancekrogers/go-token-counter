set working-directory := '..'

binary := "tcount"
bin_dir := "bin"
version := `git describe --tags --always --dirty 2>/dev/null || echo "dev"`
ldflags := "-s -w -X main.version=" + version

# Show available build recipes
[private]
@default:
    just --list --justfile {{ source_file() }}

# Build for Linux amd64
linux:
    GOOS=linux GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o {{ bin_dir }}/{{ binary }}-linux-amd64 ./cmd/tcount

# Build for Linux arm64
linux-arm64:
    GOOS=linux GOARCH=arm64 go build -ldflags '{{ ldflags }}' -o {{ bin_dir }}/{{ binary }}-linux-arm64 ./cmd/tcount

# Build for macOS amd64
darwin:
    GOOS=darwin GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o {{ bin_dir }}/{{ binary }}-darwin-amd64 ./cmd/tcount

# Build for macOS arm64
darwin-arm64:
    GOOS=darwin GOARCH=arm64 go build -ldflags '{{ ldflags }}' -o {{ bin_dir }}/{{ binary }}-darwin-arm64 ./cmd/tcount

# Build for Windows
windows:
    GOOS=windows GOARCH=amd64 go build -ldflags '{{ ldflags }}' -o {{ bin_dir }}/{{ binary }}-windows-amd64.exe ./cmd/tcount

# Build all platforms
all: linux linux-arm64 darwin darwin-arm64 windows

# Snapshot release (local dry-run)
release-snapshot:
    goreleaser release --snapshot --clean
