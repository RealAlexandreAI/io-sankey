package iosankey

import (
	"fmt"

	"github.com/RealAlexandreAI/json-repair"
	"github.com/expr-lang/expr"
	"github.com/google/uuid"
)

var (
	uuidv4Func = expr.Function("uuidv4",
		func(params ...any) (any, error) {
			return uuid.NewString(), nil
		},
		new(func() string),
	)

	jsonrepairFunc = expr.Function("jsonrepair",
		func(params ...any) (any, error) {
			var param1 string
			if p1, ok := params[0].(string); !ok {
				return nil, fmt.Errorf("param1 must be a string")
			} else {
				param1 = p1
			}
			return jsonrepair.RepairJSON(param1)
		},
		new(func(string) string),
	)

	//---------------

	exprOptions = []expr.Option{
		expr.AllowUndefinedVariables(),
		expr.Optimize(true),
	}

	builtinOptions = append(exprOptions, []expr.Option{
		uuidv4Func,
		jsonrepairFunc,
	}...)
)
