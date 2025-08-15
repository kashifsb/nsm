package cert

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kashifsb/nsm/pkg/logger"
)

type Manager struct {
	dataDir   string
	certsDir  string
	mkcertBin string
}

type CertificateInfo struct {
	CertPath string
	KeyPath  string
	Domain   string
	Created  bool
}

func NewManager(dataDir string) (*Manager, error) {
	certsDir := filepath.Join(dataDir, "certs")
	if err := os.MkdirAll(certsDir, 0o755); err != nil {
		return nil, fmt.Errorf("create certs directory: %w", err)
	}

	// Check for mkcert
	mkcertBin, err := exec.LookPath("mkcert")
	if err != nil {
		return nil, fmt.Errorf("mkcert not found: %w", err)
	}

	return &Manager{
		dataDir:   dataDir,
		certsDir:  certsDir,
		mkcertBin: mkcertBin,
	}, nil
}

func (m *Manager) EnsureCertificate(domain string, force bool) (*CertificateInfo, error) {
	if domain == "" || domain == "localhost" {
		domain = "localhost"
	}

	certPath := filepath.Join(m.certsDir, fmt.Sprintf("%s.pem", domain))
	keyPath := filepath.Join(m.certsDir, fmt.Sprintf("%s-key.pem", domain))

	info := &CertificateInfo{
		CertPath: certPath,
		KeyPath:  keyPath,
		Domain:   domain,
	}

	// Check if certificate exists and is valid
	if !force && m.certificateExists(certPath, keyPath) {
		if err := m.validateCertificate(certPath, keyPath, domain); err == nil {
			logger.Info("Using existing certificate", "domain", domain)
			return info, nil
		}

		var err error
		logger.Warn("Existing certificate is invalid, recreating", "domain", domain, "error", err)
	}

	// Create new certificate
	logger.Info("Creating new certificate", "domain", domain)
	if err := m.createCertificate(domain, certPath, keyPath); err != nil {
		return nil, fmt.Errorf("create certificate: %w", err)
	}

	info.Created = true
	return info, nil
}

func (m *Manager) IsMkcertInstalled() bool {
	_, err := exec.LookPath("mkcert")
	return err == nil
}

func (m *Manager) InstallCA() error {
	logger.Info("Installing mkcert CA")

	cmd := exec.Command(m.mkcertBin, "-install")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("install CA failed: %w\nOutput: %s", err, output)
	}

	return nil
}

func (m *Manager) GetCALocation() (string, error) {
	cmd := exec.Command(m.mkcertBin, "-CAROOT")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("get CA root: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (m *Manager) certificateExists(certPath, keyPath string) bool {
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	return certErr == nil && keyErr == nil
}

func (m *Manager) createCertificate(domain, certPath, keyPath string) error {
	args := []string{
		"-cert-file", certPath,
		"-key-file", keyPath,
		domain,
	}

	// Add common localhost variants
	if domain == "localhost" {
		args = append(args, "127.0.0.1", "::1")
	}

	cmd := exec.Command(m.mkcertBin, args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mkcert failed: %w\nOutput: %s", err, output)
	}

	// Verify the certificate was created
	if !m.certificateExists(certPath, keyPath) {
		return fmt.Errorf("certificate files not found after creation")
	}

	return nil
}

func (m *Manager) validateCertificate(certPath, keyPath, domain string) error {
	// Read certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("read certificate: %w", err)
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return fmt.Errorf("failed to decode PEM certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse certificate: %w", err)
	}

	// Check expiration
	if time.Now().After(cert.NotAfter) {
		return fmt.Errorf("certificate expired on %s", cert.NotAfter.Format("2006-01-02"))
	}

	// Check if expires soon (within 30 days)
	if time.Now().Add(30 * 24 * time.Hour).After(cert.NotAfter) {
		logger.Warn("Certificate expires soon", "domain", domain, "expires", cert.NotAfter)
	}

	// Verify hostname
	if err := cert.VerifyHostname(domain); err != nil {
		return fmt.Errorf("certificate not valid for %s: %w", domain, err)
	}

	// Test loading the key pair
	_, err = tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return fmt.Errorf("invalid key pair: %w", err)
	}

	return nil
}

func (m *Manager) ListCertificates() ([]CertificateInfo, error) {
	var certs []CertificateInfo

	entries, err := os.ReadDir(m.certsDir)
	if err != nil {
		return nil, fmt.Errorf("read certs directory: %w", err)
	}

	certFiles := make(map[string]string)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".pem") && !strings.HasSuffix(name, "-key.pem") {
			domain := strings.TrimSuffix(name, ".pem")
			certFiles[domain] = filepath.Join(m.certsDir, name)
		}
	}

	for domain, certPath := range certFiles {
		keyPath := filepath.Join(m.certsDir, fmt.Sprintf("%s-key.pem", domain))

		if m.certificateExists(certPath, keyPath) {
			certs = append(certs, CertificateInfo{
				CertPath: certPath,
				KeyPath:  keyPath,
				Domain:   domain,
			})
		}
	}

	return certs, nil
}

func (m *Manager) CleanupExpiredCertificates() error {
	certs, err := m.ListCertificates()
	if err != nil {
		return err
	}

	for _, cert := range certs {
		if err := m.validateCertificate(cert.CertPath, cert.KeyPath, cert.Domain); err != nil {
			logger.Info("Removing expired/invalid certificate", "domain", cert.Domain, "error", err)

			os.Remove(cert.CertPath)
			os.Remove(cert.KeyPath)
		}
	}

	return nil
}
