package solver

import (
	"fmt"
	"strconv"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// ConstraintEvaluator parses and evaluates predicates against candidate event records.
type ConstraintEvaluator struct {
	constraints []*pb.Constraint
}

// NewConstraintEvaluator creates an evaluator from a slice of proto constraints.
func NewConstraintEvaluator(constraints []*pb.Constraint) *ConstraintEvaluator {
	return &ConstraintEvaluator{
		constraints: constraints,
	}
}

// Evaluate checks if candidate record satisfies all defined constraints.
func (ce *ConstraintEvaluator) Evaluate(record map[string]any) (bool, error) {
	if len(ce.constraints) == 0 {
		return true, nil
	}

	for _, c := range ce.constraints {
		expr := strings.TrimSpace(c.GetExpression())
		if expr == "" {
			continue
		}

		satisfied, err := evaluateExpression(expr, record)
		if err != nil {
			return false, fmt.Errorf("constraint '%s' evaluation error: %w", c.GetName(), err)
		}
		if !satisfied {
			return false, nil
		}
	}

	return true, nil
}

// evaluateExpression supports "&&" (AND) and "||" (OR) compound expressions.
func evaluateExpression(expr string, record map[string]any) (bool, error) {
	// First handle AND clauses (all must pass)
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			passed, err := evaluateExpression(strings.TrimSpace(part), record)
			if err != nil {
				return false, err
			}
			if !passed {
				return false, nil
			}
		}
		return true, nil
	}

	// Handle OR clauses (at least one must pass)
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			passed, err := evaluateExpression(strings.TrimSpace(part), record)
			if err != nil {
				return false, err
			}
			if passed {
				return true, nil
			}
		}
		return false, nil
	}

	// Base single predicate: field op value or fieldA op fieldB
	return evaluatePredicate(expr, record)
}

func evaluatePredicate(pred string, record map[string]any) (bool, error) {
	// Recognized operators in order of length
	operators := []string{">=", "<=", "!=", "==", ">", "<", "="}

	var foundOp string
	var opIdx int = -1
	for _, op := range operators {
		idx := strings.Index(pred, op)
		if idx != -1 {
			foundOp = op
			opIdx = idx
			break
		}
	}

	if foundOp == "" || opIdx == -1 {
		return false, fmt.Errorf("unrecognized operator in predicate: '%s'", pred)
	}

	leftKey := strings.TrimSpace(pred[:opIdx])
	rightRaw := strings.TrimSpace(pred[opIdx+len(foundOp):])

	leftVal, exists := record[leftKey]
	if !exists {
		// Field not found in record
		return false, nil
	}

	// Clean quotes from right side if string literal
	rightValStr := strings.Trim(rightRaw, `"'`)

	// Normalize "=" to "=="
	if foundOp == "=" {
		foundOp = "=="
	}

	// Numeric comparison if both can be parsed as float
	leftNum, leftIsNum := toFloat(leftVal)
	rightNum, errRightNum := strconv.ParseFloat(rightValStr, 64)

	// Check if right side is another field reference (cross-field predicate)
	if errRightNum != nil {
		if otherVal, otherExists := record[rightRaw]; otherExists {
			if otherNum, otherIsNum := toFloat(otherVal); otherIsNum && leftIsNum {
				return compareFloat(leftNum, foundOp, otherNum)
			}
			return compareString(fmt.Sprintf("%v", leftVal), foundOp, fmt.Sprintf("%v", otherVal))
		}
	}

	if leftIsNum && errRightNum == nil {
		return compareFloat(leftNum, foundOp, rightNum)
	}

	// Fallback to string comparison
	return compareString(fmt.Sprintf("%v", leftVal), foundOp, rightValStr)
}

func toFloat(val any) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f, true
		}
	}
	return 0, false
}

func compareFloat(left float64, op string, right float64) (bool, error) {
	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	case ">":
		return left > right, nil
	case ">=":
		return left >= right, nil
	case "<":
		return left < right, nil
	case "<=":
		return left <= right, nil
	default:
		return false, fmt.Errorf("unsupported numeric operator: %s", op)
	}
}

func compareString(left, op, right string) (bool, error) {
	switch op {
	case "==":
		return left == right, nil
	case "!=":
		return left != right, nil
	default:
		return false, fmt.Errorf("unsupported string operator: %s", op)
	}
}

