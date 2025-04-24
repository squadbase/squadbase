APP := squad

.PHONY: all build test clean install uninstall crossbuild snapshot-release release-dry-run release

# Get version from git tag
GIT_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS = -ldflags "-X 'github.com/squadbase/squadbase/version.Version=$(GIT_VERSION)' -X 'github.com/squadbase/squadbase/version.BuildTime=$(BUILD_TIME)' -X 'github.com/squadbase/squadbase/version.GitCommit=$(GIT_COMMIT)'"

all: clean build test

# Build the binary
build:
	@echo "Building ${APP} CLI..."
	go build $(LDFLAGS) -o bin/${APP}

# Run tests
test:
	@echo "Running tests..."
	go test -v ./test/...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/ dist/
	go clean

# Install the binary to ~/bin and add to PATH if needed
install: build
	@echo "Installing ${APP} CLI..."
	mkdir -p $(HOME)/bin
	install -m 755 bin/${APP} $(HOME)/bin/
	@if ! echo $$PATH | grep -q "$(HOME)/bin"; then \
		echo ""; \
		echo "NOTE: $(HOME)/bin is not in your PATH."; \
		echo "Add the following line to your .profile, .bash_profile, or .zshrc:"; \
		echo "export PATH=\"\$$HOME/bin:\$$PATH\""; \
		echo "Then restart your terminal or run: source ~/.$(shell basename $$SHELL)rc"; \
	fi

# Uninstall the binary from ~/bin
uninstall:
	@echo "Uninstalling ${APP} CLI..."
	rm -f $(HOME)/bin/${APP}
	@echo "${APP} CLI has been removed."

# Cross-compile for different platforms
crossbuild:
	@echo "Cross-compiling for various platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/${APP}-linux-amd64 
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/${APP}-darwin-amd64 
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/${APP}-darwin-arm64 
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/${APP}-windows-amd64.exe 

# Create a development/test build with goreleaser
snapshot-release:
	@echo "Creating a snapshot release..."
	goreleaser release --clean --snapshot --config .goreleaser.yml

# Dry run a release with goreleaser
release-dry-run:
	@echo "Dry run release process..."
	goreleaser release --clean --skip=publish --snapshot --config .goreleaser.yml

# Create a release with goreleaser (requires a valid git tag)
release:
	@echo "Creating a release..."
	@set -o allexport; source .env.goreleaser; set +o allexport; \
	goreleaser release --clean --config .goreleaser.yml