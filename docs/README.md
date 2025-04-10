# Build

### Local Build

```bash
# Standard build
make build

# Run tests
make test

# Cross-compile for all platforms
make crossbuild

# Install locally
make install

# Uninstall
make uninstall
```

### Version Information

Version information is obtained from Git at build time:

- `GIT_VERSION`: Current Git tag (or "dev" if none exists)
- `GIT_COMMIT`: Current Git commit hash
- `BUILD_TIME`: Build timestamp

## Release Process

### Test Build (Snapshot)

You can create a snapshot build for development or testing purposes:

```bash
# Create a snapshot build using GoReleaser
make snapshot-release

# Dry-run the release process
make release-dry-run
```

This will generate binaries in the `dist/` directory without publishing them as an official release.

### Official Release

Steps to perform an official release:

1. Create a new version tag:

```bash
# Release a new stable version
git tag v1.0.0
git push origin v1.0.0

# Or a pre-release version
git tag v1.0.0-beta.1
git push origin v1.0.0-beta.1
```

2. Execute the release with GoReleaser:

```bash
make release
```

This performs the following:
- Builds binaries for multiple platforms
- Uploads them to the GitHub release page
- Updates the Homebrew formula (macOS/Linux)
- Updates the Scoop manifest (Windows)
- Generates checksums and changelog

## Alpha and Beta Releases

Pre-releases (alpha, beta, rc) follow the same process as regular releases:

1. Create a pre-release tag:
```bash
git tag v1.0.0-alpha.1
git push origin v1.0.0-alpha.1
```

2. Run the release:
```bash
make release
```

GoReleaser automatically detects if the tag includes "alpha", "beta", or "rc", and marks it as a pre-release on GitHub.

## Command Reference

### Makefile Commands

- `make build`: Build the binary
- `make test`: Run tests
- `make clean`: Remove build artifacts
- `make install`: Install the binary locally
- `make uninstall`: Uninstall the binary
- `make crossbuild`: Cross-compile for multiple platforms
- `make snapshot-release`: Create a snapshot build using GoReleaser
- `make release-dry-run`: Dry-run the release process
- `make release`: Perform the official release (requires Git tag)
