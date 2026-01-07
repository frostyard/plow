# Integration Guide

This guide explains how to integrate your project with Plow to automatically publish Debian packages.

## Prerequisites

- Your project builds `.deb` packages and attaches them to GitHub releases
- The organization secrets are configured (see [setup.md](setup.md))
- The `plow` repository has been initialized and has at least one release

## Quick Start

Add this workflow to your repository at `.github/workflows/publish-deb.yml`:

```yaml
name: Publish to APT Repository

on:
  release:
    types: [published]

jobs:
  publish:
    uses: frostyard/plow/.github/workflows/publish-deb.yml@main
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.DEB_GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.DEB_GPG_PASSPHRASE }}
      REPO_DEPLOY_KEY: ${{ secrets.DEB_REPO_DEPLOY_KEY }}
```

That's it! When you create a release with a `.deb` file attached:
- Full releases → published to `stable`
- Pre-releases → published to `testing`

## Workflow Inputs

The reusable workflow accepts these optional inputs:

| Input | Default | Description |
|-------|---------|-------------|
| `distribution` | `auto` | Target distribution: `stable`, `testing`, or `auto` |
| `deb_pattern` | `*_amd64.deb` | Glob pattern to match `.deb` files in release assets |
| `keep_versions` | `5` | Number of versions to keep per package |

### Examples

#### Explicit Distribution

Always publish to `stable`:

```yaml
jobs:
  publish:
    uses: frostyard/plow/.github/workflows/publish-deb.yml@main
    with:
      distribution: stable
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.DEB_GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.DEB_GPG_PASSPHRASE }}
      REPO_DEPLOY_KEY: ${{ secrets.DEB_REPO_DEPLOY_KEY }}
```

#### Custom Package Pattern

If your package has a specific naming convention:

```yaml
jobs:
  publish:
    uses: frostyard/plow/.github/workflows/publish-deb.yml@main
    with:
      deb_pattern: 'myapp_*_amd64.deb'
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.DEB_GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.DEB_GPG_PASSPHRASE }}
      REPO_DEPLOY_KEY: ${{ secrets.DEB_REPO_DEPLOY_KEY }}
```

#### Keep More Versions

Keep the last 10 versions instead of 5:

```yaml
jobs:
  publish:
    uses: frostyard/plow/.github/workflows/publish-deb.yml@main
    with:
      keep_versions: 10
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.DEB_GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.DEB_GPG_PASSPHRASE }}
      REPO_DEPLOY_KEY: ${{ secrets.DEB_REPO_DEPLOY_KEY }}
```

## Building .deb Packages

The workflow expects `.deb` files to be attached to your GitHub release. Here are some common approaches:

### Using goreleaser

If you're using Go with goreleaser, add nfpm configuration:

```yaml
# .goreleaser.yml
nfpms:
  - id: myapp
    package_name: myapp
    vendor: Frostyard
    homepage: https://github.com/frostyard/myapp
    maintainer: Frostyard <support@frostyard.org>
    description: My awesome application
    license: MIT
    formats:
      - deb
    bindir: /usr/bin
```

### Using nfpm Directly

Create an `nfpm.yaml`:

```yaml
name: myapp
arch: amd64
platform: linux
version: ${VERSION}
maintainer: Frostyard <support@frostyard.org>
description: My awesome application
vendor: Frostyard
homepage: https://github.com/frostyard/myapp
license: MIT
contents:
  - src: ./myapp
    dst: /usr/bin/myapp
```

Build with: `nfpm package --packager deb`

### Using dpkg-deb

For manual packaging:

```bash
mkdir -p myapp_1.0.0_amd64/DEBIAN
mkdir -p myapp_1.0.0_amd64/usr/bin

# Create control file
cat > myapp_1.0.0_amd64/DEBIAN/control << EOF
Package: myapp
Version: 1.0.0
Architecture: amd64
Maintainer: Frostyard <support@frostyard.org>
Description: My awesome application
EOF

# Copy binary
cp myapp myapp_1.0.0_amd64/usr/bin/

# Build package
dpkg-deb --build myapp_1.0.0_amd64
```

## Complete Example Workflow

Here's a complete workflow that builds and publishes a Go application:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - name: Build
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          GOOS=linux GOARCH=amd64 go build -o myapp ./cmd/myapp
      
      - name: Package
        run: |
          VERSION=${GITHUB_REF#refs/tags/v}
          
          mkdir -p pkg/DEBIAN pkg/usr/bin
          cp myapp pkg/usr/bin/
          
          cat > pkg/DEBIAN/control << EOF
          Package: myapp
          Version: ${VERSION}
          Architecture: amd64
          Maintainer: Frostyard <support@frostyard.org>
          Description: My awesome application
          EOF
          
          dpkg-deb --build pkg myapp_${VERSION}_amd64.deb
      
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: '*.deb'
          generate_release_notes: true

  publish:
    needs: build
    uses: frostyard/plow/.github/workflows/publish-deb.yml@main
    secrets:
      GPG_PRIVATE_KEY: ${{ secrets.DEB_GPG_PRIVATE_KEY }}
      GPG_PASSPHRASE: ${{ secrets.DEB_GPG_PASSPHRASE }}
      REPO_DEPLOY_KEY: ${{ secrets.DEB_REPO_DEPLOY_KEY }}
```

## Troubleshooting

### No .deb files found

Check that:
1. Your release has `.deb` files attached
2. The `deb_pattern` matches your file names
3. The release is published (not draft)

### Permission denied

Ensure your repository can access the organization secrets. Go to your organization's secret settings and check the "Repository access" configuration.

### Package not appearing

1. Check the workflow run logs for errors
2. Verify the package was added to the correct distribution
3. Check `https://frostyard.github.io/plow/dists/<dist>/main/binary-amd64/Packages`

### Concurrent publish conflicts

The workflow uses GitHub's concurrency controls to queue concurrent publishes. If you're seeing issues, check that no other workflow is stuck or failing.
