set working-directory := '..'

BUILDTOOL := "go run ./internal/buildutil"

# Show available test recipes
@default:
    just --list --justfile {{source_file()}}

# Run all tests (unit + integration, with dashboard)
all:
    @{{BUILDTOOL}} all

# Run unit tests (with dashboard)
unit:
    @{{BUILDTOOL}} test

# Run integration tests (with dashboard)
integration:
    @{{BUILDTOOL}} integration

# Run tests with coverage
coverage:
    go test ./... -coverprofile=coverage.out
    go tool cover -func=coverage.out

# Run tests with coverage HTML report
coverage-html:
    go test ./... -coverprofile=coverage.out
    go tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
    go test ./... -bench=. -benchmem
