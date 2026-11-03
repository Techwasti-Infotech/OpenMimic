package state

import (
	"testing"
)

func TestGlobalEntityRegistry(t *testing.T) {
	reg := NewGlobalEntityRegistry()
	rng := NewDeterministicRand(42)

	// Sample from empty pool
	_, ok := reg.SampleEntity("users", rng)
	if ok {
		t.Fatalf("expected empty pool to return false")
	}

	// Register 5 active users
	for i := 1; i <= 5; i++ {
		reg.RegisterEntity("users", "usr_"+string(rune('A'+i-1)))
	}

	if reg.PoolSize("users") != 5 {
		t.Fatalf("expected pool size 5, got %d", reg.PoolSize("users"))
	}

	// Sample 20 child foreign keys - all must be members of the registered pool
	validUsers := map[string]bool{
		"usr_A": true, "usr_B": true, "usr_C": true, "usr_D": true, "usr_E": true,
	}

	for i := 0; i < 20; i++ {
		sample, found := reg.SampleEntity("users", rng)
		if !found {
			t.Fatalf("expected entity to be found")
		}
		if !validUsers[sample] {
			t.Fatalf("orphaned entity sampled: %s", sample)
		}
	}
}
