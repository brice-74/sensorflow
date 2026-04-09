package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/brice-74/sensorflow/internal/log"
)

type SignCSRRequest struct {
	CSRPem     string
	CommonName string
	TTLDays    int
}

type SignCSRResponse struct {
	CertPEM   string `json:"cert_pem"`
	CACertPEM string `json:"ca_cert_pem"`
	SerialHex string `json:"serial_hex"`
	NotBefore string `json:"not_before"`
	NotAfter  string `json:"not_after"`
}

type Service struct {
	cfg       Signer
	logger    log.Logger
	caCert    *x509.Certificate
	caKey     *rsa.PrivateKey
	caCertPEM []byte
}

func NewService(cfg Signer, logger log.Logger) (*Service, error) {
	if err := os.MkdirAll(cfg.CertDir, 0o750); err != nil {
		return nil, fmt.Errorf("create cert dir: %w", err)
	}

	caKeyPath := filepath.Join(cfg.CertDir, "ca.key")
	caCertPath := filepath.Join(cfg.CertDir, "ca.crt")
	serverKeyPath := filepath.Join(cfg.CertDir, "server.key")
	serverCertPath := filepath.Join(cfg.CertDir, "server.crt")

	caKey, caCert, caCertPEM, err := ensureCA(caKeyPath, caCertPath, cfg, logger)
	if err != nil {
		return nil, err
	}
	if err := ensureServerCert(serverKeyPath, serverCertPath, caKey, caCert, cfg, logger); err != nil {
		return nil, err
	}

	return &Service{
		cfg:       cfg,
		logger:    logger,
		caCert:    caCert,
		caKey:     caKey,
		caCertPEM: caCertPEM,
	}, nil
}

func (s *Service) CACertPEM() []byte {
	return s.caCertPEM
}

func (s *Service) SignClientCSR(req SignCSRRequest) (*SignCSRResponse, error) {
	if strings.TrimSpace(req.CSRPem) == "" {
		return nil, errors.New("csr_pem is required")
	}

	csr, err := parseCSR(req.CSRPem)
	if err != nil {
		return nil, err
	}
	if err := csr.CheckSignature(); err != nil {
		return nil, fmt.Errorf("invalid CSR signature: %w", err)
	}

	ttlDays := s.cfg.ClientDays
	if req.TTLDays > 0 && req.TTLDays < ttlDays {
		ttlDays = req.TTLDays
	}

	now := time.Now().UTC()
	serial, err := randomSerial()
	if err != nil {
		return nil, fmt.Errorf("failed to generate serial: %w", err)
	}

	commonName := strings.TrimSpace(req.CommonName)
	if commonName == "" {
		commonName = csr.Subject.CommonName
	}
	if commonName == "" {
		commonName = "sensorflow-client"
	}

	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    now.Add(-5 * time.Minute),
		NotAfter:     now.Add(time.Duration(ttlDays) * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		DNSNames:     csr.DNSNames,
		IPAddresses:  csr.IPAddresses,
		URIs:         csr.URIs,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, s.caCert, csr.PublicKey, s.caKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign CSR: %w", err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	s.logger.Info("client certificate signed", log.Tags{
		"serial":     serial.Text(16),
		"subject":    commonName,
		"valid_days": fmt.Sprintf("%d", ttlDays),
		"not_after":  tmpl.NotAfter.Format(time.RFC3339),
	})

	return &SignCSRResponse{
		CertPEM:   string(certPEM),
		CACertPEM: string(s.caCertPEM),
		SerialHex: serial.Text(16),
		NotBefore: tmpl.NotBefore.Format(time.RFC3339),
		NotAfter:  tmpl.NotAfter.Format(time.RFC3339),
	}, nil
}

func ensureCA(keyPath, certPath string, cfg Signer, logger log.Logger) (*rsa.PrivateKey, *x509.Certificate, []byte, error) {
	if fileExists(keyPath) && fileExists(certPath) {
		key, err := readPrivateKey(keyPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read CA key: %w", err)
		}
		cert, certPEM, err := readCertificate(certPath)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("read CA cert: %w", err)
		}
		if !cert.IsCA {
			return nil, nil, nil, fmt.Errorf("CA cert at %s is not a CA", certPath)
		}
		return key, cert, certPEM, nil
	}

	key, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate CA key: %w", err)
	}

	now := time.Now().UTC()
	serial, err := randomSerial()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("generate CA serial: %w", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: cfg.CACommonName},
		NotBefore:             now.Add(-5 * time.Minute),
		NotAfter:              now.Add(time.Duration(cfg.CADays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create CA cert: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return nil, nil, nil, fmt.Errorf("write CA key: %w", err)
	}
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return nil, nil, nil, fmt.Errorf("write CA cert: %w", err)
	}

	cert, err := x509.ParseCertificate(certDER)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse generated CA cert: %w", err)
	}
	logger.Info("generated new CA", log.Tags{"cert_path": certPath})
	return key, cert, certPEM, nil
}

func ensureServerCert(keyPath, certPath string, caKey *rsa.PrivateKey, caCert *x509.Certificate, cfg Signer, logger log.Logger) error {
	if fileExists(keyPath) && fileExists(certPath) {
		_, err := readPrivateKey(keyPath)
		if err != nil {
			return fmt.Errorf("read server key: %w", err)
		}
		cert, _, err := readCertificate(certPath)
		if err != nil {
			return fmt.Errorf("read server cert: %w", err)
		}
		if time.Now().After(cert.NotAfter) {
			logger.Info("server cert expired, regenerating", log.Tags{"expired_at": cert.NotAfter.Format(time.RFC3339)})
		} else {
			return nil
		}
	}

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate server key: %w", err)
	}

	now := time.Now().UTC()
	serial, err := randomSerial()
	if err != nil {
		return fmt.Errorf("generate server serial: %w", err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: cfg.ServerCommonName},
		NotBefore:    now.Add(-5 * time.Minute),
		NotAfter:     now.Add(time.Duration(cfg.ServerDays) * 24 * time.Hour),
		DNSNames:     cfg.ServerDNSNames,
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create server cert: %w", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	if err := os.WriteFile(keyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("write server key: %w", err)
	}
	if err := os.WriteFile(certPath, certPEM, 0o644); err != nil {
		return fmt.Errorf("write server cert: %w", err)
	}

	logger.Info("generated server cert", log.Tags{"cert_path": certPath})
	return nil
}

func parseCSR(csrPEM string) (*x509.CertificateRequest, error) {
	block, _ := pem.Decode([]byte(csrPEM))
	if block == nil || block.Type != "CERTIFICATE REQUEST" {
		return nil, errors.New("csr_pem must be a valid PEM CSR")
	}
	csr, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse CSR: %w", err)
	}
	return csr, nil
}

func readPrivateKey(path string) (*rsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, errors.New("invalid PEM private key")
	}
	if block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("unsupported private key type %s", block.Type)
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func readCertificate(path string) (*x509.Certificate, []byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, nil, errors.New("invalid PEM certificate")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	return cert, b, nil
}

func randomSerial() (*big.Int, error) {
	limit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, limit)
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
