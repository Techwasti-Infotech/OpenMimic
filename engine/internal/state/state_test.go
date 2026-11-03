package state

import (
	"testing"
	"time"
)

func TestClockDeterministicProgress(t *testing.T) {
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	step := 1 * time.Second
	c := NewClock(base, step)

	if !c.Now().Equal(base) {
		t.Fatalf("expected initial time %v, got %v", base, c.Now())
	}

	t1 := c.Tick()
	expected1 := base.Add(1 * time.Second)
	if !t1.Equal(expected1) {
		t.Fatalf("expected tick 1 to be %v, got %v", expected1, t1)
	}

	t2 := c.Tick()
	expected2 := base.Add(2 * time.Second)
	if !t2.Equal(expected2) {
		t.Fatalf("expected tick 2 to be %v, got %v", expected2, t2)
	}

	c.Reset()
	if !c.Now().Equal(base) {
		t.Fatalf("expected time after reset to be %v, got %v", base, c.Now())
	}
}

func TestDeterministicRand(t *testing.T) {
	seed := int64(987654321)
	r1 := NewDeterministicRand(seed)
	r2 := NewDeterministicRand(seed)

	for i := 0; i < 100; i++ {
		f1 := r1.Float64()
		f2 := r2.Float64()
		if f1 != f2 {
			t.Fatalf("iter %d: Float64 mismatch %v vs %v", i, f1, f2)
		}

		n1 := r1.NormFloat64()
		n2 := r2.NormFloat64()
		if n1 != n2 {
			t.Fatalf("iter %d: NormFloat64 mismatch %v vs %v", i, n1, n2)
		}

		u1 := r1.Uint32()
		u2 := r2.Uint32()
		if u1 != u2 {
			t.Fatalf("iter %d: Uint32 mismatch %v vs %v", i, u1, u2)
		}

		randInt1 := r1.Intn(50)
		randInt2 := r2.Intn(50)
		if randInt1 != randInt2 {
			t.Fatalf("iter %d: Intn mismatch %v vs %v", i, randInt1, randInt2)
		}
	}
}

func TestGeneratorStateVariables(t *testing.T) {
	gs := NewGeneratorState(42, time.Now(), time.Second)
	gs.SetVar("counter", 100)

	val, found := gs.GetVar("counter")
	if !found || val != 100 {
		t.Fatalf("expected counter=100, got %v, found: %v", val, found)
	}

	_, found = gs.GetVar("nonexistent")
	if found {
		t.Fatalf("expected nonexistent to return found=false")
	}
}

