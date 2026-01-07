# Plow

Plow is a Debian package repository manager designed for GitHub Actions. It enables automatic publishing of `.deb` packages from GitHub releases to a GitHub Pages-hosted APT repository.

## Features

- **GitHub Actions Integration**: Reusable workflow for publishing packages from any repository
- **Automatic Distribution Selection**: Pre-releases go to `testing`, full releases go to `stable`
- **GPG Signing**: Automatic signing of repository metadata
- **Version Pruning**: Keeps only the N most recent versions of each package
- **Zero Server Infrastructure**: Everything runs in GitHub Actions, hosted on GitHub Pages

## Quick Start

### For End Users (Installing Packages)

```bash
# Add the GPG key
curl -fsSL https://frostyard.github.io/plow/public.key | sudo gpg --dearmor -o /usr/share/keyrings/frostyard.gpg

# Add the repository
echo "deb [signed-by=/usr/share/keyrings/frostyard.gpg] https://frostyard.github.io/plow stable main" | sudo tee /etc/apt/sources.list.d/frostyard.list

# Update and install packages
sudo apt update
sudo apt install <package-name>
```

### For Repository Administrators

See [docs/setup.md](docs/setup.md) for initial setup instructions.

### For Project Maintainers

See [docs/integration.md](docs/integration.md) for integrating your project.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Your Project Repository                       │
│  (e.g., frostyard/myapp)                                            │
├─────────────────────────────────────────────────────────────────────┤
│  on: release (published)                                            │
│       │                                                             │
│       ▼                                                             │
│  Calls: frostyard/plow/.github/workflows/publish-deb.yml@main       │
└───────┼─────────────────────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     frostyard/plow (gh-pages branch)                 │
├─────────────────────────────────────────────────────────────────────┤
│  dists/                                                             │
│  ├── stable/main/binary-amd64/Packages{,.gz,.xz}                   │
│  └── testing/main/binary-amd64/Packages{,.gz,.xz}                  │
│  pool/main/<package>/<package>_<version>_<arch>.deb                │
│  public.key                                                         │
└─────────────────────────────────────────────────────────────────────┘
```

## CLI Usage

Plow includes a CLI tool for local repository management:

```bash
# Initialize a new repository
plow init

# Add a package
plow add mypackage_1.0.0_amd64.deb --dist stable

# Regenerate index files
plow index --dist stable

# Sign the repository
plow sign --dist stable

# Prune old versions (keep 5)
plow prune --keep-versions 5
```

## Repository Structure

```
repo/
├── dists/
│   ├── stable/
│   │   ├── main/
│   │   │   └── binary-amd64/
│   │   │       ├── Packages
│   │   │       ├── Packages.gz
│   │   │       └── Packages.xz
│   │   ├── Release
│   │   ├── Release.gpg
│   │   └── InRelease
│   └── testing/
│       └── (same structure)
├── pool/
│   └── main/
│       └── <first-letter>/
│           └── <package-name>/
│               └── <package>_<version>_amd64.deb
├── public.key
└── index.html
```

## License

MIT
