package models

import "time"

// JWTResult represents a decoded JWT
type JWTResult struct {
	Raw       string                 `json:"raw"`
	Valid     bool                   `json:"valid"`
	Header    map[string]interface{} `json:"header"`
	Payload   map[string]interface{} `json:"payload"`
	Signature string                 `json:"signature"`
	Claims    *JWTClaims             `json:"claims,omitempty"`
	Errors    []string               `json:"errors,omitempty"`
}

// JWTClaims represents standard JWT claims
type JWTClaims struct {
	Issuer     string    `json:"iss,omitempty"`
	Subject    string    `json:"sub,omitempty"`
	Audience   []string  `json:"aud,omitempty"`
	ExpiresAt  time.Time `json:"exp,omitempty"`
	NotBefore  time.Time `json:"nbf,omitempty"`
	IssuedAt   time.Time `json:"iat,omitempty"`
	JWTID      string    `json:"jti,omitempty"`
	IsExpired  bool      `json:"is_expired"`
	ExpiresIn  string    `json:"expires_in,omitempty"`
}
