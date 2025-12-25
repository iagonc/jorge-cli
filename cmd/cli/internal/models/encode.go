package models

// EncodeOperation represents an encoding/decoding operation
type EncodeOperation string

const (
	EncodeBase64    EncodeOperation = "base64"
	DecodeBase64    EncodeOperation = "base64-decode"
	EncodeURL       EncodeOperation = "url"
	DecodeURL       EncodeOperation = "url-decode"
	EncodeHex       EncodeOperation = "hex"
	DecodeHex       EncodeOperation = "hex-decode"
	HashMD5         EncodeOperation = "md5"
	HashSHA1        EncodeOperation = "sha1"
	HashSHA256      EncodeOperation = "sha256"
	HashSHA512      EncodeOperation = "sha512"
	EncodeBcrypt    EncodeOperation = "bcrypt"
)

// EncodeResult represents the result of an encoding operation
type EncodeResult struct {
	Operation EncodeOperation `json:"operation"`
	Input     string          `json:"input"`
	Output    string          `json:"output"`
	Error     string          `json:"error,omitempty"`
}
