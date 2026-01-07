# User Installation Guide

This guide explains how to add the Frostyard APT repository to your Debian/Ubuntu system.

## Supported Systems

- Debian 11 (Bullseye) and later
- Ubuntu 20.04 and later
- Any Debian-based distribution

## Installation

### Step 1: Add the GPG Key

Download and install the repository signing key:

```bash
curl -fsSL https://frostyard.github.io/plow/public.key | sudo gpg --dearmor -o /usr/share/keyrings/frostyard.gpg
```

### Step 2: Add the Repository

For **stable** releases (recommended for production):

```bash
echo "deb [signed-by=/usr/share/keyrings/frostyard.gpg] https://frostyard.github.io/plow stable main" | sudo tee /etc/apt/sources.list.d/frostyard.list
```

For **testing** releases (pre-releases, beta versions):

```bash
echo "deb [signed-by=/usr/share/keyrings/frostyard.gpg] https://frostyard.github.io/plow testing main" | sudo tee /etc/apt/sources.list.d/frostyard.list
```

### Step 3: Update Package Lists

```bash
sudo apt update
```

### Step 4: Install Packages

```bash
sudo apt install <package-name>
```

## One-Line Installation

For convenience, here's a one-liner that does everything:

```bash
curl -fsSL https://frostyard.github.io/plow/public.key | sudo gpg --dearmor -o /usr/share/keyrings/frostyard.gpg && echo "deb [signed-by=/usr/share/keyrings/frostyard.gpg] https://frostyard.github.io/plow stable main" | sudo tee /etc/apt/sources.list.d/frostyard.list && sudo apt update
```

## Verifying the Installation

Check that the repository is correctly configured:

```bash
apt-cache policy
```

You should see an entry for `https://frostyard.github.io/plow`.

List available packages:

```bash
apt-cache search --names-only "^.*$" | grep -i frostyard
```

Or browse the package index directly:

```bash
curl -s https://frostyard.github.io/plow/dists/stable/main/binary-amd64/Packages
```

## Switching Between Stable and Testing

To switch from stable to testing (or vice versa):

```bash
# Edit the repository configuration
sudo nano /etc/apt/sources.list.d/frostyard.list

# Change 'stable' to 'testing' or vice versa
# Save and exit

# Update package lists
sudo apt update
```

## Removing the Repository

To remove the Frostyard repository:

```bash
# Remove the repository configuration
sudo rm /etc/apt/sources.list.d/frostyard.list

# Remove the GPG key
sudo rm /usr/share/keyrings/frostyard.gpg

# Update package lists
sudo apt update
```

Note: This does not uninstall packages that were installed from the repository. To uninstall a package:

```bash
sudo apt remove <package-name>
```

## Troubleshooting

### GPG Key Issues

If you get a GPG error, try removing and re-adding the key:

```bash
sudo rm /usr/share/keyrings/frostyard.gpg
curl -fsSL https://frostyard.github.io/plow/public.key | sudo gpg --dearmor -o /usr/share/keyrings/frostyard.gpg
sudo apt update
```

### HTTPS Issues

Ensure you have the `apt-transport-https` package installed:

```bash
sudo apt install apt-transport-https ca-certificates
```

### 404 Errors

If you get 404 errors, check:
1. The repository URL is correct
2. GitHub Pages is enabled for the plow repository
3. The distribution name is correct (`stable` or `testing`)

### Package Not Found

If a package isn't found after adding the repository:
1. Run `sudo apt update` to refresh package lists
2. Check if the package exists: `apt-cache show <package-name>`
3. Verify the package architecture matches your system

## Security

All packages in this repository are signed with GPG. The signature is verified automatically when you install packages. If verification fails, APT will refuse to install the package.

The public key fingerprint can be verified with:

```bash
gpg --show-keys /usr/share/keyrings/frostyard.gpg
```
