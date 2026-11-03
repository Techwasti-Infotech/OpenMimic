package pseudonym

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"openmimic/engine/internal/generator"
	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// Pseudonymizer provides deterministic PII de-identification and format-preserving masking.
type Pseudonymizer struct {
	method         string
	salt           []byte
	sourceCategory string
	maskFormat     string
}

// NewPseudonymizer builds a pseudonymizer configured for the field.
func NewPseudonymizer(cfg *pb.PseudonymizeFunction) *Pseudonymizer {
	method := "sha256_salt"
	if cfg.GetMethod() != "" {
		method = cfg.GetMethod()
	}

	salt := []byte(cfg.GetSalt())
	if len(salt) == 0 {
		salt = []byte("openmimic_default_hmac_secret")
	}

	category := "email"
	if cfg.GetSourceCategory() != "" {
		category = cfg.GetSourceCategory()
	}

	return &Pseudonymizer{
		method:         method,
		salt:           salt,
		sourceCategory: category,
		maskFormat:     cfg.GetMaskFormat(),
	}
}

// Pseudonymize generates or transforms a value, guaranteeing deterministic hash or format masking.
func (p *Pseudonymizer) Pseudonymize(rng *state.DeterministicRand) any {
	// First generate realistic synthetic raw PII
	rawVal := generateRawPII(rng, p.sourceCategory)

	switch p.method {
	case "fpe_numeric":
		return p.fpeNumeric(rawVal, rng)
	case "mask_chars":
		return p.maskChars(rawVal)
	case "fpe_alphanumeric":
		return p.fpeAlphaNumeric(rawVal)
	case "sha256_salt":
		fallthrough
	default:
		h := hmac.New(sha256.New, p.salt)
		h.Write([]byte(rawVal))
		digest := hex.EncodeToString(h.Sum(nil))
		return fmt.Sprintf("psn_%s", digest[:16])
	}
}

// generateRawPII produces standard synthetic PII strings according to category.
func generateRawPII(rng *state.DeterministicRand, category string) string {
	switch strings.ToLower(category) {
	case "credit_card":
		// Synthetic 16-digit card
		return fmt.Sprintf("4%03d%04d%04d%04d", rng.Intn(1000), rng.Intn(10000), rng.Intn(10000), rng.Intn(10000))
	case "ssn":
		return fmt.Sprintf("%03d-%02d-%04d", 100+rng.Intn(800), 10+rng.Intn(89), 1000+rng.Intn(9000))
	case "phone":
		return fmt.Sprintf("+1-555-%03d-%04d", 100+rng.Intn(900), 1000+rng.Intn(9000))
	case "ip":
		return fmt.Sprintf("192.168.%d.%d", rng.Intn(255), 1+rng.Intn(254))
	case "email":
		fallthrough
	default:
		return generator.GenerateStringFaker(rng, "email", "", 0)
	}
}

// maskChars replaces sensitive middle characters with '*' while keeping prefix/suffix format.
func (p *Pseudonymizer) maskChars(raw string) string {
	if p.maskFormat != "" {
		// Custom format replacement
		return p.maskFormat
	}

	length := len(raw)
	if length <= 4 {
		return strings.Repeat("*", length)
	}

	// Keep first 2 and last 4 characters visible, mask middle
	prefix := raw[:2]
	suffix := raw[length-4:]
	maskedMiddle := strings.Repeat("*", length-6)
	return prefix + maskedMiddle + suffix
}

// fpeNumeric preserves digit lengths and hyphen delimiters deterministically.
func (p *Pseudonymizer) fpeNumeric(raw string, rng *state.DeterministicRand) string {
	h := hmac.New(sha256.New, p.salt)
	h.Write([]byte(raw))
	hashBytes := h.Sum(nil)

	var sb strings.Builder
	byteIdx := 0
	for _, ch := range raw {
		if ch >= '0' && ch <= '9' {
			// Deterministically map to another digit [0-9]
			digit := hashBytes[byteIdx%len(hashBytes)] % 10
			sb.WriteByte('0' + digit)
			byteIdx++
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}

// fpeAlphaNumeric preserves format structure with alphanumeric character substitution.
func (p *Pseudonymizer) fpeAlphaNumeric(raw string) string {
	h := hmac.New(sha256.New, p.salt)
	h.Write([]byte(raw))
	hashHex := hex.EncodeToString(h.Sum(nil))

	var sb strings.Builder
	for i, ch := range raw {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
			sb.WriteByte(hashHex[i%len(hashHex)])
		} else {
			sb.WriteRune(ch)
		}
	}
	return sb.String()
}
