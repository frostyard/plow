# Plow - Debian Repository Manager

## Project Overview

Plow is a Debian package repository manager designed for GitHub Actions. It enables automatic publishing of `.deb` packages from GitHub releases to a GitHub Pages-hosted APT repository.

## Development Guidelines

### Before Committing Changes

Always run the following commands to verify changes:

```bash
make test    # Run tests with race detection
make lint    # Run linters (vet, fmt-check, golangci-lint)
make build   # Build the binary
```

Or run all three with:

```bash
make all
```

### Code Style

- Follow standard Go conventions
- Run `make fmt` to format code before committing
- All exported functions must have documentation comments
- Tests should be in `*_test.go` files alongside the code they test

### Project Structure

```
plow/
├── cmd/plow/           # CLI entrypoint
├── internal/
│   ├── cli/            # CLI commands (cobra)
│   ├── deb/            # .deb file parsing and version comparison
│   ├── gpg/            # GPG signing wrapper
│   └── repo/           # Repository structure and metadata
├── .github/workflows/  # GitHub Actions workflows
└── docs/               # Documentation
```

### Key Packages

- `internal/deb`: Parses `.deb` files, extracts control metadata, handles Debian version comparison
- `internal/repo`: Manages repository directory structure, generates Packages/Release files
- `internal/gpg`: Wraps GPG CLI for signing Release files
- `internal/cli`: Cobra CLI commands (init, add, index, sign, prune)

### Testing

- Run `make test` for standard test run
- Run `make test-coverage` to generate HTML coverage report
- Tests use the standard `testing` package

### Workflows

- `ci.yml`: Runs on push/PR to main, builds and tests
- `release.yml`: Triggered on version tags, builds binaries for multiple platforms
- `init-pages.yml`: Manual workflow to initialize gh-pages branch
- `publish-deb.yml`: Reusable workflow called by other repos to publish packages
