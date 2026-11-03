package pseudonym

import (
	"strings"
	"testing"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestPseudonymizeSha256Salt(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	cfg := &pb.PseudonymizeFunction{
		Method:         "sha256_salt",
		Salt:           "my_secret_key",
		SourceCategory: "email",
	}

	p := NewPseudonymizer(cfg)
	res := p.Pseudonymize(rng).(string)

	if !strings.HasPrefix(res, "psn_") {
		t.Errorf("expected psn_ prefix, got %s", res)
	}
}

func TestPseudonymizeFPENumeric(t *testing.T) {
	rng := state.NewDeterministicRand(99)
	cfg := &pb.PseudonymizeFunction{
		Method:         "fpe_numeric",
		Salt:           "token_key",
		SourceCategory: "ssn",
	}

	p := NewPseudonymizer(cfg)
	res := p.Pseudonymize(rng).(string)

	// SSN format: 000-00-0000 (length 11, hyphens at positions 3 and 6)
	if len(res) != 11 {
		t.Errorf("expected SSN length 11, got %d (%s)", len(res), res)
	}
	if res[3] != '-' || res[6] != '-' {
		t.Errorf("expected hyphens preserved at positions 3 and 6 in %s", res)
	}
}

func TestPseudonymizeMaskChars(t *testing.T) {
	rng := state.NewDeterministicRand(77)
	cfg := &pb.PseudonymizeFunction{
		Method:         "mask_chars",
		SourceCategory: "credit_card",
	}

	p := NewPseudonymizer(cfg)
	res := p.Pseudonymize(rng).(string)

	if len(res) != 16 {
		t.Fatalf("expected credit card length 16, got %d", len(res))
	}
	if !strings.Contains(res, "******") {
		t.Errorf("expected middle characters masked with *, got %s", res)
	}
}
