package expression

import (
	"math"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestNumericExpression(t *testing.T) {
	eval := NewEvaluator(&pb.CalculatedExpressionFunction{
		Expression: "quantity * unit_price * (1 - discount)",
		ReturnType: "double",
	})

	row := map[string]any{
		"quantity":   5.0,
		"unit_price": 20.0,
		"discount":   0.10,
	}

	result := eval.Evaluate(row).(float64)
	expected := 5.0 * 20.0 * 0.90 // 90.0
	if math.Abs(result-expected) > 0.001 {
		t.Errorf("expected %f, got %f", expected, result)
	}
}

func TestStringConcatenation(t *testing.T) {
	eval := NewEvaluator(&pb.CalculatedExpressionFunction{
		Expression: "first_name + ' ' + last_name",
		ReturnType: "string",
	})

	row := map[string]any{
		"first_name": "Alice",
		"last_name":  "Smith",
	}

	result := eval.Evaluate(row).(string)
	if result != "Alice Smith" {
		t.Errorf("expected 'Alice Smith', got '%s'", result)
	}
}

