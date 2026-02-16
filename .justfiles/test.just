set working-directory := '..'

# Show available test recipes
@default:
    just --list --justfile {{source_file()}}

# Run all tests (unit + integration)
all: unit integration

# Run unit tests
unit:
    go test ./... -v -short

# Run integration tests
integration:
    go test ./... -v -run Integration

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
