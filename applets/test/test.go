// Package test implements the `test` applet (and its `[` alias):
// the POSIX conditional-expression evaluator.
package test

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/Real-Fruit-Snacks/topsail/internal/applet"
	"github.com/Real-Fruit-Snacks/topsail/internal/ioutil"
)

func init() {
	applet.Register(applet.Applet{
		Name:    "test",
		Aliases: []string{"["},
		Help:    "evaluate a conditional expression",
		Usage:   usage,
		Main:    Main,
	})
}

const usage = `Usage: test EXPRESSION
       [ EXPRESSION ]
Evaluate EXPRESSION as a conditional and exit 0 if true, 1 if false,
or 2 on a syntax error.

File tests:
  -e FILE   FILE exists
  -f FILE   FILE exists and is a regular file
  -d FILE   FILE exists and is a directory
  -r FILE   FILE is readable
  -w FILE   FILE is writable
  -x FILE   FILE is executable
  -s FILE   FILE has size > 0
  -L FILE   FILE is a symbolic link

String tests:
  -z STRING        length is zero
  -n STRING        length is non-zero
  STRING1 = STRING2
  STRING1 != STRING2

Integer tests (-eq -ne -lt -le -gt -ge):
  N1 -eq N2

Logical:
  ! EXPR        negate
  EXPR1 -a EXPR2  AND  (deprecated)
  EXPR1 -o EXPR2  OR   (deprecated)

When invoked as '[', the final argument MUST be ']'.
`

// Main is the applet entry point.
func Main(argv []string) int {
	args := argv[1:]
	if argv[0] == "[" {
		if len(args) == 0 || args[len(args)-1] != "]" {
			ioutil.Errf("[: missing ']'")
			return 2
		}
		args = args[:len(args)-1]
	}

	if len(args) == 0 {
		// `test` with no args is false.
		return 1
	}
	p := &parser{tokens: args}
	v, err := p.orExpr()
	if err != nil {
		ioutil.Errf("test: %v", err)
		return 2
	}
	if !p.eof() {
		ioutil.Errf("test: syntax error: unexpected %q", p.peek())
		return 2
	}
	if v {
		return 0
	}
	return 1
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

func (p *parser) eat() string {
	if p.pos >= len(p.tokens) {
		return ""
	}
	t := p.tokens[p.pos]
	p.pos++
	return t
}

func (p *parser) eof() bool { return p.pos >= len(p.tokens) }

func (p *parser) orExpr() (bool, error) {
	left, err := p.andExpr()
	if err != nil {
		return false, err
	}
	for p.peek() == "-o" {
		p.eat()
		right, err := p.andExpr()
		if err != nil {
			return false, err
		}
		left = left || right
	}
	return left, nil
}

func (p *parser) andExpr() (bool, error) {
	left, err := p.notExpr()
	if err != nil {
		return false, err
	}
	for p.peek() == "-a" {
		p.eat()
		right, err := p.notExpr()
		if err != nil {
			return false, err
		}
		left = left && right
	}
	return left, nil
}

func (p *parser) notExpr() (bool, error) {
	if p.peek() == "!" {
		p.eat()
		v, err := p.notExpr()
		if err != nil {
			return false, err
		}
		return !v, nil
	}
	return p.primary()
}

func (p *parser) primary() (bool, error) {
	if p.eof() {
		return false, errors.New("missing operand")
	}
	if p.peek() == "(" {
		p.eat()
		v, err := p.orExpr()
		if err != nil {
			return false, err
		}
		if p.eat() != ")" {
			return false, errors.New("syntax error: missing ')'")
		}
		return v, nil
	}

	// 3-arg forms (binary): X OP Y
	if p.pos+2 < len(p.tokens)+1 && p.pos+2 <= len(p.tokens) {
		// Peek the operator at pos+1.
		if p.pos+1 < len(p.tokens) {
			op := p.tokens[p.pos+1]
			if isBinaryOp(op) {
				lhs := p.eat()
				p.eat() // consume op
				rhs := p.eat()
				return binaryOp(lhs, op, rhs)
			}
		}
	}

	// 2-arg forms (unary): -OP X
	t := p.peek()
	if isUnaryOp(t) {
		p.eat()
		if p.eof() {
			return false, errors.New("missing operand")
		}
		arg := p.eat()
		return unaryOp(t, arg)
	}

	// 1-arg form: STRING (true if non-empty)
	v := p.eat()
	return v != "", nil
}

func isUnaryOp(s string) bool {
	switch s {
	case "-e", "-f", "-d", "-r", "-w", "-x", "-s", "-L", "-h", "-z", "-n":
		return true
	}
	return false
}

func isBinaryOp(s string) bool {
	switch s {
	case "=", "!=",
		"-eq", "-ne", "-lt", "-le", "-gt", "-ge":
		return true
	}
	return false
}

func unaryOp(op, arg string) (bool, error) {
	switch op {
	case "-z":
		return arg == "", nil
	case "-n":
		return arg != "", nil
	case "-e":
		_, err := os.Stat(arg)
		return err == nil, nil
	case "-f":
		info, err := os.Stat(arg)
		return err == nil && info.Mode().IsRegular(), nil
	case "-d":
		info, err := os.Stat(arg)
		return err == nil && info.IsDir(), nil
	case "-s":
		info, err := os.Stat(arg)
		return err == nil && info.Size() > 0, nil
	case "-L", "-h":
		info, err := os.Lstat(arg)
		return err == nil && info.Mode()&os.ModeSymlink != 0, nil
	case "-r":
		_, err := os.Stat(arg)
		if err != nil {
			return false, nil
		}
		f, err := os.Open(arg) //nolint:gosec // user-supplied path is the whole point
		if err != nil {
			return false, nil
		}
		_ = f.Close()
		return true, nil
	case "-w":
		info, err := os.Stat(arg)
		if err != nil {
			return false, nil
		}
		return info.Mode()&0o200 != 0, nil
	case "-x":
		info, err := os.Stat(arg)
		if err != nil {
			return false, nil
		}
		return info.Mode()&0o111 != 0, nil
	}
	return false, errors.New("unknown unary operator: " + op)
}

func binaryOp(lhs, op, rhs string) (bool, error) {
	switch op {
	case "=":
		return lhs == rhs, nil
	case "!=":
		return lhs != rhs, nil
	case "-eq", "-ne", "-lt", "-le", "-gt", "-ge":
		a, err1 := strconv.ParseInt(strings.TrimSpace(lhs), 10, 64)
		b, err2 := strconv.ParseInt(strings.TrimSpace(rhs), 10, 64)
		if err1 != nil || err2 != nil {
			return false, errors.New("integer expression expected: " + lhs + " " + op + " " + rhs)
		}
		switch op {
		case "-eq":
			return a == b, nil
		case "-ne":
			return a != b, nil
		case "-lt":
			return a < b, nil
		case "-le":
			return a <= b, nil
		case "-gt":
			return a > b, nil
		case "-ge":
			return a >= b, nil
		}
	}
	return false, errors.New("unknown binary operator: " + op)
}
