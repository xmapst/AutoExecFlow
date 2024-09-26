package eval

import (
	"fmt"
	"math/rand"
	"reflect"
	"regexp"

	"github.com/expr-lang/expr"
	"github.com/pkg/errors"
)

var exprMatcher = regexp.MustCompile(`{{\s*(.+?)\s*}}`)

func EvaluateExpr(ex string, c map[string]any) (any, error) {
	ex = sanitizeExpr(ex)
	env := map[string]any{
		"randomInt": randomInt,
		"sequence":  sequence,
		"sprintf":   fmt.Sprintf,
	}
	for k, v := range c {
		env[k] = v
	}
	program, err := expr.Compile(ex, expr.Env(env))
	if err != nil {
		return "", errors.Wrapf(err, "error compiling expression: %s", ex)
	}
	output, err := expr.Run(program, env)
	if err != nil {
		return "", errors.Wrapf(err, "error evaluating expression: %s", ex)
	}
	return output, nil
}

func sanitizeExpr(ex string) string {
	if matches := exprMatcher.FindStringSubmatch(ex); matches != nil {
		return matches[1]
	}
	return ex
}

func randomInt(args ...any) (int, error) {
	if len(args) == 1 {
		if args[0] == nil {
			return 0, errors.Errorf("not expecting nil argument")
		}
		v := reflect.ValueOf(args[0])
		if !v.CanInt() {
			return 0, errors.Errorf("invalid arg type %s", v.Type())
		}
		return rand.Intn(int(v.Int())), nil
	} else if len(args) == 0 {
		return rand.Int(), nil
	} else {
		return 0, errors.Errorf("invalid number of arguments for trim (expected 0 or 1, got %d)", len(args))
	}
}

func sequence(start, stop int) []int {
	if start > stop {
		return []int{}
	}
	result := make([]int, stop-start)
	for ix := range result {
		result[ix] = start
		start = start + 1
	}
	return result
}
