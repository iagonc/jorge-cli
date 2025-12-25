package encode

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/url"

	"github.com/iagonc/jorge-cli/cmd/cli/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// EncodeUsecase handles encoding operations
type EncodeUsecase struct {
	logger *zap.Logger
}

// NewEncodeUsecase creates a new EncodeUsecase
func NewEncodeUsecase(logger *zap.Logger) *EncodeUsecase {
	return &EncodeUsecase{logger: logger}
}

// Encode performs an encoding operation
func (u *EncodeUsecase) Encode(operation models.EncodeOperation, input string) *models.EncodeResult {
	result := &models.EncodeResult{
		Operation: operation,
		Input:     input,
	}

	var output string
	var err error

	switch operation {
	case models.EncodeBase64:
		output = base64.StdEncoding.EncodeToString([]byte(input))

	case models.DecodeBase64:
		decoded, e := base64.StdEncoding.DecodeString(input)
		if e != nil {
			err = e
		} else {
			output = string(decoded)
		}

	case models.EncodeURL:
		output = url.QueryEscape(input)

	case models.DecodeURL:
		decoded, e := url.QueryUnescape(input)
		if e != nil {
			err = e
		} else {
			output = decoded
		}

	case models.EncodeHex:
		output = hex.EncodeToString([]byte(input))

	case models.DecodeHex:
		decoded, e := hex.DecodeString(input)
		if e != nil {
			err = e
		} else {
			output = string(decoded)
		}

	case models.HashMD5:
		hash := md5.Sum([]byte(input))
		output = hex.EncodeToString(hash[:])

	case models.HashSHA1:
		hash := sha1.Sum([]byte(input))
		output = hex.EncodeToString(hash[:])

	case models.HashSHA256:
		hash := sha256.Sum256([]byte(input))
		output = hex.EncodeToString(hash[:])

	case models.HashSHA512:
		hash := sha512.Sum512([]byte(input))
		output = hex.EncodeToString(hash[:])

	case models.EncodeBcrypt:
		hashed, e := bcrypt.GenerateFromPassword([]byte(input), bcrypt.DefaultCost)
		if e != nil {
			err = e
		} else {
			output = string(hashed)
		}

	default:
		err = fmt.Errorf("unknown operation: %s", operation)
	}

	if err != nil {
		result.Error = err.Error()
	} else {
		result.Output = output
	}

	return result
}

// VerifyBcrypt verifies a bcrypt hash
func (u *EncodeUsecase) VerifyBcrypt(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetAvailableOperations returns list of available operations
func (u *EncodeUsecase) GetAvailableOperations() []string {
	return []string{
		"base64", "base64-decode",
		"url", "url-decode",
		"hex", "hex-decode",
		"md5", "sha1", "sha256", "sha512",
		"bcrypt",
	}
}
