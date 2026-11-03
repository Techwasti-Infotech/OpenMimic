package generator

import (
	"math"
	"regexp"
	"strings"
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestUUIDGeneration(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for i := 0; i < 20; i++ {
		u := GenerateUUID(rng, false)
		if !uuidRegex.MatchString(u) {
			t.Fatalf("iter %d: invalid UUID v4 format: %s", i, u)
		}
	}

	rng2 := state.NewDeterministicRand(42)
	upper := GenerateUUID(rng2, true)
	if upper != strings.ToUpper(upper) {
		t.Fatalf("expected uppercase UUID, got %s", upper)
	}
}

func TestNormalDistribution(t *testing.T) {
	rng := state.NewDeterministicRand(100)
	mean := 50.0
	stdDev := 5.0
	n := 10000

	var sum float64
	for i := 0; i < n; i++ {
		val := GenerateNormalDistribution(rng, mean, stdDev)
		sum += val
	}
	sampleMean := sum / float64(n)
	if math.Abs(sampleMean-mean) > 0.25 {
		t.Fatalf("sample mean %f too far from expected mean %f", sampleMean, mean)
	}
}

func TestUniformDistribution(t *testing.T) {
	rng := state.NewDeterministicRand(200)
	min := 10.0
	max := 20.0

	for i := 0; i < 500; i++ {
		val := GenerateUniformDistribution(rng, min, max)
		if val < min || val > max {
			t.Fatalf("val %f outside range [%f, %f]", val, min, max)
		}
	}
}

func TestNowFunction(t *testing.T) {
	baseTime := time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)
	clock := state.NewClock(baseTime, 1*time.Second)

	rfc := GenerateNow(clock, "RFC3339")
	if rfc != "2026-06-15T10:00:00Z" {
		t.Fatalf("expected RFC3339 formatted date, got %v", rfc)
	}

	epochMs := GenerateNow(clock, "epoch_ms")
	if epochMs != baseTime.UnixMilli() {
		t.Fatalf("expected epoch_ms %d, got %v", baseTime.UnixMilli(), epochMs)
	}

	epochS := GenerateNow(clock, "epoch_s")
	if epochS != baseTime.Unix() {
		t.Fatalf("expected epoch_s %d, got %v", baseTime.Unix(), epochS)
	}

	clock.Tick()
	rfcTick := GenerateNow(clock, "RFC3339")
	if rfcTick != "2026-06-15T10:00:01Z" {
		t.Fatalf("expected tick date 2026-06-15T10:00:01Z, got %v", rfcTick)
	}
}

func TestSequentialInteger(t *testing.T) {
	tests := []struct {
		index    int64
		start    int64
		step     int64
		expected int64
	}{
		{index: 0, start: 1, step: 1, expected: 1},
		{index: 5, start: 1, step: 1, expected: 6},
		{index: 3, start: 100, step: 5, expected: 115},
		{index: 10, start: 0, step: 2, expected: 20},
	}

	for _, tt := range tests {
		res := GenerateSequentialInteger(tt.index, tt.start, tt.step)
		if res != tt.expected {
			t.Fatalf("for index=%d start=%d step=%d expected=%d, got=%d", tt.index, tt.start, tt.step, tt.expected, res)
		}
	}
}

func TestStringFaker(t *testing.T) {
	rng := state.NewDeterministicRand(300)

	name := GenerateStringFaker(rng, "name", "", 0)
	if !strings.Contains(name, " ") {
		t.Fatalf("expected full name with space, got %s", name)
	}

	email := GenerateStringFaker(rng, "email", "", 0)
	if !strings.Contains(email, "@") {
		t.Fatalf("expected valid email with @, got %s", email)
	}

	pattern := "ORDER_####"
	patResult := GenerateStringFaker(rng, "", pattern, 0)
	if !strings.HasPrefix(patResult, "ORDER_") || len(patResult) != 10 {
		t.Fatalf("expected pattern matching ORDER_####, got %s", patResult)
	}
}

func TestWeightedChoice(t *testing.T) {
	rng := state.NewDeterministicRand(400)
	options := []*pb.WeightedChoiceOption{
		{Value: "SUCCESS", Weight: 95.0},
		{Value: "FAILURE", Weight: 5.0},
	}

	counts := make(map[string]int)
	n := 10000
	for i := 0; i < n; i++ {
		res := GenerateWeightedChoice(rng, options)
		counts[res]++
	}

	failRate := float64(counts["FAILURE"]) / float64(n)
	// Expect ~5% failure rate
	if failRate < 0.035 || failRate > 0.065 {
		t.Fatalf("expected failure rate around 0.05, got %f", failRate)
	}
}

func TestMarkovChainProgression(t *testing.T) {
	st := state.NewGeneratorState(500, time.Now(), time.Second)
	chain := &pb.MarkovChainFunction{
		InitialState: "VIEW_CART",
		States: []*pb.MarkovState{
			{
				StateName: "VIEW_CART",
				Transitions: []*pb.MarkovTransition{
					{ToState: "CHECKOUT", Probability: 1.0},
				},
			},
			{
				StateName: "CHECKOUT",
				Transitions: []*pb.MarkovTransition{
					{ToState: "PURCHASED", Probability: 1.0},
				},
			},
		},
	}

	step1 := GenerateMarkovChain(st, "cart_state", chain)
	if step1 != "CHECKOUT" {
		t.Fatalf("expected transition to CHECKOUT, got %s", step1)
	}

	step2 := GenerateMarkovChain(st, "cart_state", chain)
	if step2 != "PURCHASED" {
		t.Fatalf("expected transition to PURCHASED, got %s", step2)
	}
}

func TestEntityPool(t *testing.T) {
	rng := state.NewDeterministicRand(600)
	pool := &pb.EntityPoolFunction{
		PoolName:     "users",
		PoolSize:     50,
		EntityPrefix: "usr_",
	}

	for i := 0; i < 20; i++ {
		ent := GenerateEntityPool(rng, pool)
		if !strings.HasPrefix(ent, "usr_") {
			t.Fatalf("expected entity prefix 'usr_', got %s", ent)
		}
	}
}
