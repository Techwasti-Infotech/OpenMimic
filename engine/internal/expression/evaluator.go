package expression

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	pb "openmimic/engine/pkg/proto/v1"
)

// Evaluator parses and evaluates runtime expressions against current row records.
type Evaluator struct {
	expr       string
	returnType string
}

// NewEvaluator creates a new Evaluator.
func NewEvaluator(cfg *pb.CalculatedExpressionFunction) *Evaluator {
	retType := "double"
	if cfg.GetReturnType() != "" {
		retType = cfg.GetReturnType()
	}
	return &Evaluator{
		expr:       strings.TrimSpace(cfg.GetExpression()),
		returnType: retType,
	}
}

// Evaluate computes the formula value given the current record state.
func (e *Evaluator) Evaluate(record map[string]any) any {
	if e.expr == "" {
		return nil
	}

	// 1. String Concatenation Mode: e.g. "first_name + ' ' + last_name"
	if strings.Contains(e.expr, "'") || strings.Contains(e.expr, "\"") || e.returnType == "string" {
		return e.evalStringConcat(record)
	}

	// 2. Numerical Expression Evaluation
	numVal := e.evalNumeric(record)
	switch e.returnType {
	case "int64", "int":
		return int64(numVal)
	case "bool":
		return numVal > 0
	case "double", "float":
		fallthrough
	default:
		return numVal
	}
}

// evalStringConcat joins string literals and field references.
func (e *Evaluator) evalStringConcat(record map[string]any) string {
	parts := strings.Split(e.expr, "+")
	var sb strings.Builder
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if (strings.HasPrefix(trimmed, "'") && strings.HasSuffix(trimmed, "'")) ||
			(strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"")) {
			// String literal
			if len(trimmed) >= 2 {
				sb.WriteString(trimmed[1 : len(trimmed)-1])
			}
		} else {
			// Field identifier
			if val, exists := record[trimmed]; exists && val != nil {
				sb.WriteString(fmt.Sprintf("%v", val))
			}
		}
	}
	return sb.String()
}

// evalNumeric tokenizes and evaluates arithmetic formulas: + - * / with field substitutions.
func (e *Evaluator) evalNumeric(record map[string]any) float64 {
	formula := e.expr
	for field, val := range record {
		if val == nil {
			continue
		}
		var fVal float64
		switch v := val.(type) {
		case float64:
			fVal = v
		case float32:
			fVal = float64(v)
		case int:
			fVal = float64(v)
		case int32:
			fVal = float64(v)
		case int64:
			fVal = float64(v)
		default:
			if parsed, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err == nil {
				fVal = parsed
			}
		}
		// Replace field identifier as distinct token
		formula = replaceToken(formula, field, fmt.Sprintf("%f", fVal))
	}

	result, err := parseArithmetic(formula)
	if err != nil {
		return 0.0
	}
	return result
}

// replaceToken ensures exact whole-word token replacement.
func replaceToken(src, token, replacement string) string {
	var sb strings.Builder
	i := 0
	for i < len(src) {
		if strings.HasPrefix(src[i:], token) {
			leftBoundary := (i == 0) || !isIdentChar(rune(src[i-1]))
			rightBoundary := (i+len(token) >= len(src)) || !isIdentChar(rune(src[i+len(token)]))
			if leftBoundary && rightBoundary {
				sb.WriteString(replacement)
				i += len(token)
				continue
			}
		}
		sb.WriteByte(src[i])
		i++
	}
	return sb.String()
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Simple recursive descent parser for basic arithmetic expressions with parentheses
type parser struct {
	tokens []string
	pos    int
}

func parseArithmetic(expr string) (float64, error) {
	tokens := tokenize(expr)
	p := &parser{tokens: tokens, pos: 0}
	val := p.parseExpr()
	return val, nil
}

func (p *parser) parseExpr() float64 {
	val := p.parseTerm()
	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "+" && op != "-" {
			break
		}
		p.pos++
		term := p.parseTerm()
		if op == "+" {
			val += term
		} else {
			val -= term
		}
	}
	return val
}

func (p *parser) parseTerm() float64 {
	val := p.parseFactor()
	for p.pos < len(p.tokens) {
		op := p.tokens[p.pos]
		if op != "*" && op != "/" {
			break
		}
		p.pos++
		factor := p.parseFactor()
		if op == "*" {
			val *= factor
		} else if factor != 0 {
			val /= factor
		}
	}
	return val
}

func (p *parser) parseFactor() float64 {
	if p.pos >= len(p.tokens) {
		return 0.0
	}
	tok := p.tokens[p.pos]
	p.pos++

	if tok == "(" {
		val := p.parseExpr()
		if p.pos < len(p.tokens) && p.tokens[p.pos] == ")" {
			p.pos++
		}
		return val
	}

	if tok == "-" {
		return -p.parseFactor()
	}

	num, _ := strconv.ParseFloat(tok, 64)
	return num
}

func tokenize(s string) []string {
	var tokens []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if unicode.IsSpace(rune(ch)) {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			continue
		}
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' || ch == '(' || ch == ')' {
			if cur.Len() > 0 {
				tokens = append(tokens, cur.String())
				cur.Reset()
			}
			tokens = append(tokens, string(ch))
			continue
		}
		cur.WriteByte(ch)
	}
	if cur.Len() > 0 {
		tokens = append(tokens, cur.String())
	}
	return tokens
}

