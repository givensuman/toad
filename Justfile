# Print help message
help:
    #!/usr/bin/env bash
    just --list

# Build program binary
build:
    #!/usr/bin/env bash
    mkdir -p bin
    cd src
    go build -ldflags="-s -w" -trimpath -o bin/toad .

# Install program binary
install:
    #!/usr/bin/env bash
    cd src
    go install

# Run the program
run *args:
    #!/usr/bin/env bash
    export DEBUG=true
    cd src
    go run {{ args }}

# Run program tests
test:
    #!/usr/bin/env bash
    cd src
    go test -parallel=1 -v ./...
    # go tool cover -func=coverage.out | sort -rnk3

# Clean development environment
clean:
    @rm -rf coverage.out bin/

# Display test coverage
cover:
    cd src
    go test -v -race $(shell go list ./... | grep -v /vendor/) -v -coverprofile=coverage.out
    go tool cover -func=coverage.out

# Format, lint and vet; all in one!
check: fmt lint vet

alias format := fmt
# Format program files
fmt:
    #!/usr/bin/env bash
    just --unstable --fmt
    cd src
    gofmt -w -s -l .

# Lint program files
lint:
    #!/usr/bin/env bash
    cd src
    golangci-lint run

# Vet program files
vet:
    #!/usr/bin/env bash
    cd src
    go vet ./...
