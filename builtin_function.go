package iosankey

import (
	"bytes"
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/RealAlexandreAI/json-repair"
	"github.com/expr-lang/expr"
	"github.com/google/uuid"
	"github.com/itchyny/gojq"
	"github.com/samber/lo"
	"github.com/spyzhov/ajson"
	"text/template"
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

	jsonpathFunc = expr.Function("jsonpath",
		func(params ...any) (any, error) {
			this, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("param0 type check (string)")
			}
			path, ok := params[1].(string)
			if !ok {
				return nil, fmt.Errorf("param1 type check (string)")
			}

			root, _ := ajson.Unmarshal([]byte(this))
			nodes, err := root.JSONPath(path)
			if err != nil {
				return nil, fmt.Errorf("jsonpath engine eval error %w", err)
			}

			var rst []string

			for _, node := range nodes {
				rst = append(rst, node.String())
			}

			return rst, nil
		},
		new(func(string, string) []string),
	)

	sprigFunc = expr.Function("sprig",
		func(params ...any) (any, error) {
			tpl, ok := params[0].(string)
			if !ok {
				return nil, fmt.Errorf("param0 type check (string)")
			}

			fmap := sprig.TxtFuncMap()
			t := template.Must(template.New("anonymous").Funcs(fmap).Parse(tpl))

			var b bytes.Buffer
			err := t.Execute(&b, params[1])
			if err != nil {
				return "", err
			}
			return b.String(), nil
		},
		new(func(string, any) string),
	)

	jqFunc = expr.Function("jq",
		func(params ...any) (any, error) {
			var input any
			switch t := params[0].(type) {
			case map[string]any, []any:
				input = t
			case []map[string]any:
				input = lo.Map(t, func(item map[string]any, index int) any {
					return item
				})
			default:
				return nil, fmt.Errorf("param0 type check (string)")
			}

			path, ok := params[1].(string)
			if !ok {
				return nil, fmt.Errorf("param1 type check (string)")
			}

			query, err := gojq.Parse(path)
			if err != nil {
				return nil, fmt.Errorf("gojq Parse error %w", err)
			}

			iter := query.Run(input)

			var rst []any

			for {
				v, ok := iter.Next()
				if !ok {
					break
				}
				if err, ok := v.(error); ok {
					return nil, fmt.Errorf("gojq Run error %w", err)
				} else {
					rst = append(rst, v)
				}
			}

			return rst, nil
		},
		new(func(any, string) []any),
	)

	//---------------

	exprOptions = []expr.Option{
		expr.AllowUndefinedVariables(),
		expr.Optimize(true),
	}

	builtinOptions = append(exprOptions, []expr.Option{
		uuidv4Func,
		jsonrepairFunc,
		jqFunc,
		sprigFunc,
		jsonpathFunc,
	}...)
)
