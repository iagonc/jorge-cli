package ssl

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
)

// TLSDialer interface for testability
type TLSDialer interface {
	DialContext(ctx context.Context, network, addr string) (*tls.Conn, error)
}

// DefaultTLSDialer implements TLSDialer using standard library
type DefaultTLSDialer struct {
	Timeout time.Duration
}

func (d *DefaultTLSDialer) DialContext(ctx context.Context, network, addr string) (*tls.Conn, error) {
	dialer := &net.Dialer{
		Timeout: d.Timeout,
	}

	conn, err := tls.DialWithDialer(dialer, network, addr, &tls.Config{
		InsecureSkipVerify: false,
	})
	if err != nil {
		// Try again with insecure to get cert info even if invalid
		conn, err = tls.DialWithDialer(dialer, network, addr, &tls.Config{
			InsecureSkipVerify: true,
		})
	}
	return conn, err
}

// SSLCheckUsecase handles SSL certificate checking
type SSLCheckUsecase struct {
	Logger *zap.Logger
	Dialer TLSDialer
}

// NewSSLCheckUsecase creates a new SSLCheckUsecase
func NewSSLCheckUsecase(logger *zap.Logger) *SSLCheckUsecase {
	return &SSLCheckUsecase{
		Logger: logger,
		Dialer: &DefaultTLSDialer{Timeout: 10 * time.Second},
	}
}

// CheckSSL performs an SSL certificate check on the given host and port
func (u *SSLCheckUsecase) CheckSSL(ctx context.Context, host string, port int) (*models.SSLCheckResult, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	u.Logger.Info("Checking SSL certificate", zap.String("address", addr))

	result := &models.SSLCheckResult{
		Host: host,
		Port: port,
	}

	// Dial TLS connection
	conn, err := u.Dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Connection failed: %v", err))
		return result, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Get connection state
	state := conn.ConnectionState()

	// Get TLS version
	result.TLSVersion = tlsVersionString(state.Version)
	result.CipherSuite = tls.CipherSuiteName(state.CipherSuite)

	// Process certificates
	if len(state.PeerCertificates) == 0 {
		result.Errors = append(result.Errors, "No certificates received")
		return result, fmt.Errorf("no certificates received")
	}

	// Process leaf certificate
	leafCert := state.PeerCertificates[0]
	result.Certificate = processCertificate(leafCert)

	// Process certificate chain
	for i, cert := range state.PeerCertificates {
		chainCert := models.SSLChainCert{
			Subject: cert.Subject.String(),
			Issuer:  cert.Issuer.String(),
			IsCA:    cert.IsCA,
			Level:   i,
		}
		result.Chain = append(result.Chain, chainCert)
	}

	// Verify certificate chain
	opts := x509.VerifyOptions{
		DNSName:       host,
		Intermediates: x509.NewCertPool(),
	}

	for _, cert := range state.PeerCertificates[1:] {
		opts.Intermediates.AddCert(cert)
	}

	_, err = leafCert.Verify(opts)
	result.IsChainValid = err == nil
	if err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Chain verification: %v", err))
	}

	// Add warnings for expiring certificates
	if result.Certificate.DaysUntilExpiry <= 30 && result.Certificate.DaysUntilExpiry > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Certificate expires in %d days", result.Certificate.DaysUntilExpiry))
	}

	u.Logger.Info("SSL check completed",
		zap.String("host", host),
		zap.Bool("valid", result.Certificate.IsValid),
		zap.Int("days_until_expiry", result.Certificate.DaysUntilExpiry))

	return result, nil
}

func processCertificate(cert *x509.Certificate) models.SSLCertInfo {
	now := time.Now()
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	return models.SSLCertInfo{
		Subject:         cert.Subject.String(),
		Issuer:          cert.Issuer.String(),
		ValidFrom:       cert.NotBefore,
		ValidUntil:      cert.NotAfter,
		DaysUntilExpiry: daysUntilExpiry,
		IsExpired:       now.After(cert.NotAfter),
		IsValid:         now.After(cert.NotBefore) && now.Before(cert.NotAfter),
		SerialNumber:    cert.SerialNumber.String(),
		SignatureAlgo:   cert.SignatureAlgorithm.String(),
		DNSNames:        cert.DNSNames,
	}
}

func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("Unknown (0x%04x)", version)
	}
}
