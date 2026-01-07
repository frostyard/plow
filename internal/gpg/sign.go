// Package gpg provides GPG signing functionality.
package gpg

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Signer handles GPG signing operations.
type Signer struct {
	KeyID      string // Optional: specific key ID to use
	Passphrase string // Optional: passphrase from environment
}

// NewSigner creates a new GPG signer.
func NewSigner(keyID string) *Signer {
	return &Signer{
		KeyID:      keyID,
		Passphrase: os.Getenv("GPG_PASSPHRASE"),
	}
}

// SignRelease signs the Release file, creating Release.gpg and InRelease.
func (s *Signer) SignRelease(distDir string) error {
	releasePath := filepath.Join(distDir, "Release")
	releaseGpgPath := filepath.Join(distDir, "Release.gpg")
	inReleasePath := filepath.Join(distDir, "InRelease")

	// Remove old signatures (ignore errors, files may not exist)
	_ = os.Remove(releaseGpgPath)
	_ = os.Remove(inReleasePath)

	// Create detached signature (Release.gpg)
	if err := s.signDetached(releasePath, releaseGpgPath); err != nil {
		return fmt.Errorf("create Release.gpg: %w", err)
	}

	// Create inline signature (InRelease)
	if err := s.signInline(releasePath, inReleasePath); err != nil {
		return fmt.Errorf("create InRelease: %w", err)
	}

	return nil
}

func (s *Signer) signDetached(inputPath, outputPath string) error {
	args := []string{
		"--batch",
		"--yes",
		"--armor",
		"--detach-sign",
		"--output", outputPath,
	}

	if s.KeyID != "" {
		args = append(args, "--default-key", s.KeyID)
	}

	if s.Passphrase != "" {
		args = append(args, "--pinentry-mode", "loopback", "--passphrase-fd", "0")
	}

	args = append(args, inputPath)

	cmd := exec.Command("gpg", args...)
	if s.Passphrase != "" {
		cmd.Stdin = bytes.NewReader([]byte(s.Passphrase))
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	return nil
}

func (s *Signer) signInline(inputPath, outputPath string) error {
	args := []string{
		"--batch",
		"--yes",
		"--armor",
		"--clearsign",
		"--output", outputPath,
	}

	if s.KeyID != "" {
		args = append(args, "--default-key", s.KeyID)
	}

	if s.Passphrase != "" {
		args = append(args, "--pinentry-mode", "loopback", "--passphrase-fd", "0")
	}

	args = append(args, inputPath)

	cmd := exec.Command("gpg", args...)
	if s.Passphrase != "" {
		cmd.Stdin = bytes.NewReader([]byte(s.Passphrase))
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}

	return nil
}

// ExportPublicKey exports the public key in ASCII-armored format.
func (s *Signer) ExportPublicKey(outputPath string) error {
	args := []string{
		"--armor",
		"--export",
	}

	if s.KeyID != "" {
		args = append(args, s.KeyID)
	}

	cmd := exec.Command("gpg", args...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, output, 0644)
}
