set working-directory := '..'

# Show available test recipes
@default:
    just --list --justfile {{source_file()}}

# Run unit tests
unit:
    go test ./... -v

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
