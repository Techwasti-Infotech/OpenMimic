package generator

import (
	"fmt"
	"strings"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// GenerateUUID generates standard v4 UUIDs deterministically based on state RNG.
func GenerateUUID(rng *state.DeterministicRand, uppercase bool) string {
	u := fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		rng.Uint32(),
		uint16(rng.Uint32()),
		uint16(rng.Uint32()&0x0fff|0x4000), // version 4
		uint16(rng.Uint32()&0x3fff|0x8000), // variant 1
		rng.Uint64()&0xffffffffffff,
	)
	if uppercase {
		return strings.ToUpper(u)
	}
	return u
}

// GenerateNormalDistribution generates a float according to a Gaussian distribution.
func GenerateNormalDistribution(rng *state.DeterministicRand, mean, stdDev float64) float64 {
	return mean + rng.NormFloat64()*stdDev
}

// GenerateUniformDistribution generates a float uniformly distributed between min and max.
func GenerateUniformDistribution(rng *state.DeterministicRand, min, max float64) float64 {
	if max <= min {
		return min
	}
	return min + rng.Float64()*(max-min)
}

// GenerateNow outputs a timestamp based on the logical clock.
func GenerateNow(clock *state.Clock, format string) any {
	t := clock.Now()
	if format == "" || format == "RFC3339" {
		return t.Format(time.RFC3339)
	}
	switch format {
	case "epoch_ms":
		return t.UnixMilli()
	case "epoch_s":
		return t.Unix()
	default:
		return t.Format(format)
	}
}

// GenerateSequentialInteger outputs start + (index * step).
func GenerateSequentialInteger(index int64, start, step int64) int64 {
	if step == 0 {
		step = 1
	}
	return start + (index * step)
}

// GenerateStringFaker produces deterministic synthetic strings for categories like names, emails, cities, etc.
func GenerateStringFaker(rng *state.DeterministicRand, category, pattern string, index int64) string {
	firstNames := []string{"Alice", "Bob", "Charlie", "Diana", "Ethan", "Fiona", "George", "Hannah", "Ian", "Julia"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Miller", "Davis", "Wilson", "Taylor", "Anderson"}
	cities := []string{"New York", "London", "Tokyo", "Berlin", "San Francisco", "Bengaluru", "Sydney", "Paris", "Toronto", "Singapore"}
	domains := []string{"example.com", "testmail.org", "mimicdata.io", "samplecloud.net"}

	if pattern != "" {
		var b strings.Builder
		for _, r := range pattern {
			if r == '#' {
				b.WriteString(fmt.Sprintf("%d", rng.Intn(10)))
			} else {
				b.WriteRune(r)
			}
		}
		return b.String()
	}

	switch strings.ToLower(category) {
	case "name", "full_name":
		fn := firstNames[rng.Intn(len(firstNames))]
		ln := lastNames[rng.Intn(len(lastNames))]
		return fmt.Sprintf("%s %s", fn, ln)
	case "first_name":
		return firstNames[rng.Intn(len(firstNames))]
	case "last_name":
		return lastNames[rng.Intn(len(lastNames))]
	case "email":
		fn := strings.ToLower(firstNames[rng.Intn(len(firstNames))])
		ln := strings.ToLower(lastNames[rng.Intn(len(lastNames))])
		dom := domains[rng.Intn(len(domains))]
		return fmt.Sprintf("%s.%s%d@%s", fn, ln, rng.Intn(1000), dom)
	case "city":
		return cities[rng.Intn(len(cities))]
	case "country":
		countries := []string{"United States", "United Kingdom", "Germany", "Japan", "India", "Canada", "Australia", "France"}
		return countries[rng.Intn(len(countries))]
	default:
		return fmt.Sprintf("val_%d_%04d", index, rng.Intn(10000))
	}
}

// GenerateWeightedChoice picks a choice according to relative weights.
func GenerateWeightedChoice(rng *state.DeterministicRand, options []*pb.WeightedChoiceOption) string {
	if len(options) == 0 {
		return ""
	}
	var totalWeight float64
	for _, opt := range options {
		totalWeight += opt.Weight
	}
	if totalWeight <= 0 {
		return options[0].Value
	}

	r := rng.Float64() * totalWeight
	var current float64
	for _, opt := range options {
		current += opt.Weight
		if r <= current {
			return opt.Value
		}
	}
	return options[len(options)-1].Value
}

// GenerateMarkovChain steps through state transitions according to transition probabilities.
func GenerateMarkovChain(st *state.GeneratorState, key string, chain *pb.MarkovChainFunction) string {
	if chain == nil || len(chain.States) == 0 {
		return "INITIAL"
	}

	// Retrieve current state from state machine
	currentState := chain.InitialState
	if val, found := st.GetVar(key); found {
		if s, ok := val.(string); ok && s != "" {
			currentState = s
		}
	}

	// Find transition rules for currentState
	var activeTransitions []*pb.MarkovTransition
	for _, stateDef := range chain.States {
		if stateDef.StateName == currentState {
			activeTransitions = stateDef.Transitions
			break
		}
	}

	if len(activeTransitions) == 0 {
		return currentState
	}

	var totalWeight float64
	for _, tr := range activeTransitions {
		totalWeight += tr.Probability
	}

	r := st.RNG.Float64() * totalWeight
	nextState := currentState
	var current float64
	for _, tr := range activeTransitions {
		current += tr.Probability
		if r <= current {
			nextState = tr.ToState
			break
		}
	}

	st.SetVar(key, nextState)
	return nextState
}

// GenerateEntityPool picks an entity from a fixed, shared pool to maintain foreign-key consistency.
func GenerateEntityPool(rng *state.DeterministicRand, pool *pb.EntityPoolFunction) string {
	if pool == nil {
		return "ent_0"
	}
	size := pool.PoolSize
	if size <= 0 {
		size = 100
	}
	prefix := pool.EntityPrefix
	if prefix == "" {
		prefix = "id_"
	}

	entityIdx := rng.Intn(int(size))
	return fmt.Sprintf("%s%05d", prefix, entityIdx)
}
