package bsql

import (
	"errors"
	"strconv"
	"strings"
)

type CV map[string]interface{}

type SqlExpr struct {
	Expr   string
	Values []interface{}
	Result string
	Err    error
}

func Expr(expr string, args ...interface{}) (sqlExpr *SqlExpr) {
	sqlExpr = &SqlExpr{Expr: expr, Values: args}
	exprs := strings.Split(expr, "?")
	if len(exprs) == 1 {
		sqlExpr.Result = expr
		return
	}
	if len(exprs)-1 != len(args) {
		sqlExpr.Err = errors.New("inconsistent number of parameters")
		return
	}
	result := exprs[0]
	for _, arg := range args {
		switch s := arg.(type) {
		case string:
			result = result + s
		case int:
			result = result + strconv.FormatInt(int64(s), 10)
		case int8:
			result = result + strconv.FormatInt(int64(s), 10)
		case int16:
			result = result + strconv.FormatInt(int64(s), 10)
		case int32:
			result = result + strconv.FormatInt(int64(s), 10)
		case int64:
			result = result + strconv.FormatInt(int64(s), 10)
		case float32:
			result = result + strconv.FormatFloat(float64(s), 'f', -1, 64)
		case float64:
			result = result + strconv.FormatFloat(s, 'f', -1, 64)
		case bool:
			if s {
				result = result + "1"
			} else {
				result = result + "0"
			}
		default:
			sqlExpr.Err = errors.New("parameter rejected")
		}
	}
	sqlExpr.Result = result
	return
}
