package solver

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestConstraintSolverPredicates(t *testing.T) {
	constraints := []*pb.Constraint{
		{
			Id:         "c1",
			Name:       "High value transaction rule",
			Expression: "amount >= 1000 && status == SUCCESS",
		},
	}

	eval := NewConstraintEvaluator(constraints)

	// Test satisfying candidate
	passRecord := map[string]any{
		"amount": 1500.50,
		"status": "SUCCESS",
		"id":     "tx_1",
	}
	ok, err := eval.Evaluate(passRecord)
	if err != nil || !ok {
		t.Fatalf("expected record to pass constraint, got ok=%v err=%v", ok, err)
	}

	// Test failing candidate (amount too low)
	failRecord1 := map[string]any{
		"amount": 800.0,
		"status": "SUCCESS",
	}
	ok, err = eval.Evaluate(failRecord1)
	if err != nil || ok {
		t.Fatalf("expected record to fail constraint on amount, got ok=%v", ok)
	}

	// Test failing candidate (wrong status)
	failRecord2 := map[string]any{
		"amount": 2000.0,
		"status": "FAILED",
	}
	ok, err = eval.Evaluate(failRecord2)
	if err != nil || ok {
		t.Fatalf("expected record to fail constraint on status, got ok=%v", ok)
	}
}

func TestCrossFieldPredicate(t *testing.T) {
	constraints := []*pb.Constraint{
		{
			Id:         "c2",
			Name:       "Max exceeds min",
			Expression: "max_limit > min_limit",
		},
	}

	eval := NewConstraintEvaluator(constraints)

	pass := map[string]any{"max_limit": 100, "min_limit": 20}
	ok, _ := eval.Evaluate(pass)
	if !ok {
		t.Fatalf("expected pass for max_limit > min_limit")
	}

	fail := map[string]any{"max_limit": 10, "min_limit": 50}
	ok, _ = eval.Evaluate(fail)
	if ok {
		t.Fatalf("expected fail for max_limit <= min_limit")
	}
}

