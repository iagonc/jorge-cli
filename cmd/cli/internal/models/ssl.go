package models

import "time"

// SSLCertInfo contains detailed certificate information
type SSLCertInfo struct {
	Subject         string    `json:"subject"`
	Issuer          string    `json:"issuer"`
	ValidFrom       time.Time `json:"valid_from"`
	ValidUntil      time.Time `json:"valid_until"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	IsExpired       bool      `json:"is_expired"`
	IsValid         bool      `json:"is_valid"`
	SerialNumber    string    `json:"serial_number"`
	SignatureAlgo   string    `json:"signature_algo"`
	DNSNames        []string  `json:"dns_names"`
}

// SSLChainCert represents a certificate in the chain
type SSLChainCert struct {
	Subject string `json:"subject"`
	Issuer  string `json:"issuer"`
	IsCA    bool   `json:"is_ca"`
	Level   int    `json:"level"` // 0 = leaf, 1+ = intermediate/root
}

// SSLCheckResult contains the complete SSL check result
type SSLCheckResult struct {
	Host         string         `json:"host"`
	Port         int            `json:"port"`
	Certificate  SSLCertInfo    `json:"certificate"`
	Chain        []SSLChainCert `json:"chain"`
	TLSVersion   string         `json:"tls_version"`
	CipherSuite  string         `json:"cipher_suite"`
	IsChainValid bool           `json:"is_chain_valid"`
	Errors       []string       `json:"errors,omitempty"`
	Warnings     []string       `json:"warnings,omitempty"`
}
