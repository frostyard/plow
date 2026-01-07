# Setup Guide

This guide walks through the initial setup of Plow for your organization.

## Prerequisites

- A GitHub organization (e.g., `frostyard`)
- Admin access to create organization secrets
- A machine with GPG installed (for key generation)

## Step 1: Generate GPG Key

Generate a dedicated GPG key for signing packages. This key will be stored as a GitHub secret.

### Option A: Without Passphrase (Simpler)

```bash
gpg --batch --gen-key <<EOF
%no-protection
Key-Type: RSA
Key-Length: 4096
Key-Usage: sign
Name-Real: Frostyard Debian Repository
Name-Email: debian@frostyard.org
Expire-Date: 0
EOF
```

### Option B: With Passphrase (More Secure)

```bash
gpg --full-generate-key
```

When prompted:
- Key type: RSA (sign only) or RSA and RSA
- Key size: 4096
- Expiration: 0 (no expiration) or your preference
- Real name: `Frostyard Debian Repository`
- Email: `debian@frostyard.org` (or your preference)
- Passphrase: Choose a strong passphrase

## Step 2: Export the Keys

### Find your key ID

```bash
gpg --list-secret-keys --keyid-format long
```

Output will look like:
```
sec   rsa4096/ABC123DEF4567890 2024-01-01 [SC]
      ABCDEF1234567890ABCDEF1234567890ABCDEF12
uid                 [ultimate] Frostyard Debian Repository <debian@frostyard.org>
```

The key ID is `ABC123DEF4567890` (the part after `rsa4096/`).

### Export private key (for GitHub secret)

```bash
gpg --armor --export-secret-keys YOUR_KEY_ID | base64 -w0 > private-key-base64.txt
```

The contents of `private-key-base64.txt` will be your `DEB_GPG_PRIVATE_KEY` secret.

### Export public key (for the repository)

```bash
gpg --armor --export YOUR_KEY_ID > public.key
```

Keep this file - you'll add it to the repository later.

## Step 3: Create Deploy Key

The reusable workflow needs to push to the `plow` repository's `gh-pages` branch. Create an SSH deploy key:

```bash
ssh-keygen -t ed25519 -C "plow-deploy-key" -f plow-deploy-key -N ""
```

This creates:
- `plow-deploy-key` (private key) - goes to organization secret
- `plow-deploy-key.pub` (public key) - goes to repository deploy keys

## Step 4: Configure GitHub

### Add Organization Secrets

Go to your organization settings → Secrets and variables → Actions → New organization secret.

Add these secrets:

| Secret Name | Value | Notes |
|-------------|-------|-------|
| `DEB_GPG_PRIVATE_KEY` | Contents of `private-key-base64.txt` | Base64-encoded GPG private key |
| `DEB_GPG_PASSPHRASE` | Your GPG passphrase | Leave empty if no passphrase |
| `DEB_REPO_DEPLOY_KEY` | Contents of `plow-deploy-key` | SSH private key |

**Important**: Set "Repository access" to allow all repositories (or select specific ones).

### Add Deploy Key to Repository

Go to `frostyard/plow` → Settings → Deploy keys → Add deploy key:

- Title: `Plow Deploy Key`
- Key: Contents of `plow-deploy-key.pub`
- Allow write access: ✓ (checked)

## Step 5: Initialize the Repository

### Enable GitHub Pages

Go to `frostyard/plow` → Settings → Pages:
- Source: Deploy from a branch
- Branch: `gh-pages` / `/ (root)`

### Run the Init Workflow

Go to Actions → "Initialize GitHub Pages" → Run workflow.

This creates the `gh-pages` branch with the initial directory structure.

### Add Public Key

After the init workflow completes, add your public key:

```bash
git clone git@github.com:frostyard/plow.git
cd plow
git checkout gh-pages
cp /path/to/public.key .
git add public.key
git commit -m "Add GPG public key"
git push
```

## Step 6: Create a Release

Create a release of `plow` itself so the workflow can download the binary:

```bash
git checkout main
git tag v0.1.0
git push origin v0.1.0
```

This triggers the release workflow which builds and uploads the `plow` binary.

## Verification

Test the setup by:

1. Creating a release in another repository that uses the publish workflow
2. Checking that the package appears in `https://frostyard.github.io/plow/dists/stable/main/binary-amd64/Packages`

## Security Notes

- The GPG private key is stored encrypted in GitHub secrets
- Only workflows in your organization can access organization secrets
- The deploy key only has write access to the `plow` repository
- Consider key rotation annually or after any security incident

## Troubleshooting

### "Failed to download plow binary"

Make sure you've created at least one release of the `plow` repository.

### "Permission denied" when pushing to gh-pages

Check that the deploy key has write access and is correctly added to both the repository and the organization secret.

### GPG signing fails

Verify the GPG key was exported correctly:
```bash
echo "$DEB_GPG_PRIVATE_KEY" | base64 -d | gpg --list-packets
```

This should show packet information without errors.
