// Package expr implements the `expr` applet: evaluate arithmetic and
// string expressions.
//
// Supported operators (POSIX, lowest to highest precedence):
//
//	|             logical OR
//	&             logical AND
//	= != < <= > >=  comparison (numeric if both sides parse, else lexical)
//	+ -           additive
//	* / %         multiplicative
//
// Plus the keywords `length STRING` and `substr STRING POS LEN`.
//
// Regex match (STRING : REGEX) is intentionally deferred to a later wave.
package expr

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:  "expr",
		Help:  "evaluate arithmetic and string expressions",
		Usage: usage,
		Main:  Main,
	})
}

const usage = `Usage: expr EXPRESSION
Evaluate EXPRESSION and write the result to standard output.

Operators (lowest to highest precedence):
  EXPR1 | EXPR2     EXPR1 if neither null nor zero, else EXPR2
  EXPR1 & EXPR2     EXPR1 if neither is null/zero, else 0
  EXPR1 < EXPR2 ... arithmetic or lexical comparison (= != < <= > >=)
  EXPR1 + EXPR2     arithmetic addition / subtraction
  EXPR1 * EXPR2     arithmetic multiplication / division / modulo
  ( EXPRESSION )    grouping
  length STRING     length of STRING
  substr STRING POS LENGTH   substring of STRING

Exit status: 0 if EXPRESSION is neither null nor zero,
             1 if EXPRESSION is null or zero,
             2 if EXPRESSION is syntactically invalid.
`

// Main is the applet entry point.
func Main(argv []string) int {
	if len(argv) < 2 {
		ioutil.Errf("expr: missing operand")
		return 2
	}
	p := &parser{tokens: argv[1:]}
	v, err := p.expr()
	if err != nil {
		ioutil.Errf("expr: %v", err)
		return 2
	}
	if !p.eof() {
		ioutil.Errf("expr: syntax error: unexpected %q", p.peek())
		return 2
	}
	_, _ = fmt.Fprintln(ioutil.Stdout, v.String())
	if v.IsZero() {
		return 1
	}
	return 0
}

// value is either an integer or a string. Operators that need numeric
// semantics promote with Int().
type value struct {
	s   string
	n   int64
	num bool
}

func valStr(s string) value {
	v := value{s: s}
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		v.n = n
		v.num = true
	}
	return v
}

func valInt(n int64) value {
	return value{s: strconv.FormatInt(n, 10), n: n, num: true}
}

func (v value) String() string { return v.s }

func (v value) Int() (int64, bool) { return v.n, v.num }

func (v value) IsZero() bool {
	if v.s == "" {
		return true
	}
	if v.num && v.n == 0 {
		return true
	}
	return false
}

type parser struct {
	tokens []string
	pos    int
}

func (p *parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.pos]
}

func (p *parser) eof() bool { return p.pos >= len(p.tokens) }

func (p *parser) eat() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	t := p.tokens[p.pos]
	p.pos++
	return t
}

// expr -> orExpr
func (p *parser) expr() (value, error) {
	return p.orExpr()
}

func (p *parser) orExpr() (value, error) {
	left, err := p.andExpr()
	if err != nil {
		return value{}, err
	}
	for p.peek() == "|" {
		p.eat()
		right, err := p.andExpr()
		if err != nil {
			return value{}, err
		}
		if !left.IsZero() {
			// left wins; ignore right
		} else if !right.IsZero() {
			left = right
		} else {
			left = valInt(0)
		}
	}
	return left, nil
}

func (p *parser) andExpr() (value, error) {
	left, err := p.cmpExpr()
	if err != nil {
		return value{}, err
	}
	for p.peek() == "&" {
		p.eat()
		right, err := p.cmpExpr()
		if err != nil {
			return value{}, err
		}
		if left.IsZero() || right.IsZero() {
			left = valInt(0)
		}
	}
	return left, nil
}

func (p *parser) cmpExpr() (value, error) {
	left, err := p.addExpr()
	if err != nil {
		return value{}, err
	}
	for {
		op := p.peek()
		switch op {
		case "=", "!=", "<", "<=", ">", ">=":
			p.eat()
			right, err := p.addExpr()
			if err != nil {
				return value{}, err
			}
			left = compare(left, right, op)
		default:
			return left, nil
		}
	}
}

func compare(a, b value, op string) value {
	if ai, ok1 := a.Int(); ok1 {
		if bi, ok2 := b.Int(); ok2 {
			return cmpBool(intCmp(ai, bi, op))
		}
	}
	return cmpBool(strCmp(a.s, b.s, op))
}

func intCmp(a, b int64, op string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

func strCmp(a, b, op string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case ">":
		return a > b
	case ">=":
		return a >= b
	}
	return false
}

func cmpBool(b bool) value {
	if b {
		return valInt(1)
	}
	return valInt(0)
}

func (p *parser) addExpr() (value, error) {
	left, err := p.mulExpr()
	if err != nil {
		return value{}, err
	}
	for {
		op := p.peek()
		if op != "+" && op != "-" {
			return left, nil
		}
		p.eat()
		right, err := p.mulExpr()
		if err != nil {
			return value{}, err
		}
		ai, ok1 := left.Int()
		bi, ok2 := right.Int()
		if !ok1 || !ok2 {
			return value{}, errors.New("non-numeric argument")
		}
		if op == "+" {
			left = valInt(ai + bi)
		} else {
			left = valInt(ai - bi)
		}
	}
}

func (p *parser) mulExpr() (value, error) {
	left, err := p.primary()
	if err != nil {
		return value{}, err
	}
	for {
		op := p.peek()
		if op != "*" && op != "/" && op != "%" {
			return left, nil
		}
		p.eat()
		right, err := p.primary()
		if err != nil {
			return value{}, err
		}
		ai, ok1 := left.Int()
		bi, ok2 := right.Int()
		if !ok1 || !ok2 {
			return value{}, errors.New("non-numeric argument")
		}
		switch op {
		case "*":
			left = valInt(ai * bi)
		case "/":
			if bi == 0 {
				return value{}, errors.New("division by zero")
			}
			left = valInt(ai / bi)
		case "%":
			if bi == 0 {
				return value{}, errors.New("division by zero")
			}
			left = valInt(ai % bi)
		}
	}
}

func (p *parser) primary() (value, error) {
	if p.eof() {
		return value{}, errors.New("syntax error: missing operand")
	}
	t := p.eat()
	switch t {
	case "(":
		v, err := p.expr()
		if err != nil {
			return value{}, err
		}
		if p.eat() != ")" {
			return value{}, errors.New("syntax error: missing ')'")
		}
		return v, nil
	case "length":
		if p.eof() {
			return value{}, errors.New("length: missing operand")
		}
		s := p.eat()
		return valInt(int64(len(s))), nil
	case "substr":
		if p.pos+2 >= len(p.tokens)+1 {
			return value{}, errors.New("substr: missing operand")
		}
		s := p.eat()
		posStr := p.eat()
		lenStr := p.eat()
		pos, err := strconv.Atoi(posStr)
		if err != nil {
			return value{}, fmt.Errorf("substr: invalid position: %s", posStr)
		}
		ln, err := strconv.Atoi(lenStr)
		if err != nil {
			return value{}, fmt.Errorf("substr: invalid length: %s", lenStr)
		}
		if pos < 1 || ln <= 0 || pos > len(s) {
			return valStr(""), nil
		}
		end := pos - 1 + ln
		if end > len(s) {
			end = len(s)
		}
		return valStr(s[pos-1 : end]), nil
	}
	return valStr(t), nil
}
