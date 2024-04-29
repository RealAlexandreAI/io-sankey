package iosankey

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/patcher"
	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/samber/lo"
	"github.com/smallnest/deepcopy"
	"github.com/tidwall/sjson"
)

var (
	validate    *validator.Validate
	optsInPlace = &sjson.Options{
		Optimistic:     true,
		ReplaceInPlace: true,
	}
)

func init() {
	validate = validator.New(validator.WithRequiredStructEnabled())
}

// SankeyTransformer
// @Description:
type SankeyTransformer struct {
	expressions []string
	envs        map[string]any
	exprOptions []expr.Option
}

// NewSankeyTransformer
//
//	@Description:
//	@param opts
//	@return *SankeyTransformer
func NewSankeyTransformer(opts ...SankeyOption) *SankeyTransformer {
	transformer := &SankeyTransformer{
		expressions: make([]string, 0),
		envs:        make(map[string]any),
		exprOptions: make([]expr.Option, 0),
	}
	for _, opt := range opts {
		opt(transformer)
	}
	return transformer
}

type SankeyOption func(*SankeyTransformer)

// WithExprOptions
//
//	@Description:
//	@param exprOptions
//	@return SankeyOption
func WithExprOptions(exprOptions ...expr.Option) SankeyOption {
	return func(st *SankeyTransformer) {
		st.exprOptions = append(st.exprOptions, exprOptions...)
	}
}

// WithExpressions
//
//	@Description:
//	@param expressions
//	@return SankeyOption
func WithExpressions(expressions ...string) SankeyOption {
	return func(st *SankeyTransformer) {
		st.expressions = append(st.expressions, expressions...)
	}
}

// WithEnvs
//
//	@Description:
//	@param envs
//	@return SankeyOption
func WithEnvs(envs map[string]interface{}) SankeyOption {
	return func(st *SankeyTransformer) {
		for k, v := range envs {
			st.envs[k] = v
		}
	}
}

// Map
//
//	@Description:
//	@receiver s
//	@param src
//	@return map[string]any
//	@return error
func (s *SankeyTransformer) Map(src any) (map[string]any, error) {
	atSrc := make(map[string]any)
	if err := mapstructure.Decode(src, &atSrc); err != nil {
		return nil, fmt.Errorf("failed to decode source: %w", err)
	}

	atDst := deepcopy.Copy(atSrc)
	atDstStr, err := sonic.MarshalString(atDst)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dst string: %w", err)
	}

	mergedOptions := append(builtinOptions, s.exprOptions...)

	ctx := context.WithValue(context.TODO(), "dstCtxKey", atDstStr)

	mergedEnvs := lo.Assign[string, any](
		s.envs,
		map[string]any{
			"$src": src,
			"$dst": ctx,
			"reset": func(ctx context.Context) string {
				return "{}"
			},
			"set": func(ctx context.Context, k string, v any) string {
				dstStr := ctx.Value("dstCtxKey").(string)
				dstStr, _ = sjson.SetOptions(dstStr, k, v, optsInPlace)
				return dstStr
			},
			"drop": func(ctx context.Context, k string) string {
				dstStr := ctx.Value("dstCtxKey").(string)
				dstStr, _ = sjson.Delete(dstStr, k)
				return dstStr
			},
		},
	)

	evalOptions := append(mergedOptions, expr.Env(mergedEnvs))
	evalOptions = append(evalOptions, expr.Patch(patcher.WithContext{Name: "$dst"}))

	for _, expression := range s.expressions {

		program, compileErr := expr.Compile(expression, evalOptions...)
		if compileErr != nil {
			return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, compileErr)
		}

		output, runErr := expr.Run(program, mergedEnvs)
		if runErr != nil {
			return nil, fmt.Errorf("failed to run program with expression '%s': %w", expression, runErr)
		}

		ctx = context.WithValue(ctx, "dstCtxKey", output)
		mergedEnvs["$dst"] = ctx
	}

	ctx = mergedEnvs["$dst"].(context.Context)
	var dstMap map[string]any
	if err := sonic.UnmarshalString(ctx.Value("dstCtxKey").(string), &dstMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal destination string to map: %w", err)
	}

	return dstMap, nil
}

// Transform
//
//	@Description:
//	@receiver s
//	@param src
//	@param dst
//	@return error
func (s *SankeyTransformer) Transform(src any, dst any) error {
	var err error

	dstMap, err := s.Map(src)
	if err != nil {
		return fmt.Errorf("error mapping source to destination: %w", err)
	}

	if err = mapstructure.Decode(dstMap, dst); err != nil {
		return fmt.Errorf("error decoding to destination: %w", err)
	}

	if err = validate.Struct(dst); err != nil {
		return fmt.Errorf("error validating destination structure: %w", err)
	}

	return err
}
