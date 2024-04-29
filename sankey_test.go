package iosankey

import (
	"fmt"
	"github.com/expr-lang/expr"
	"reflect"
	"strconv"
	"testing"
)

type MixedDataTypes struct {
	IntValue    int
	FloatValue  float64
	StringValue string
	BoolValue   bool
	ArrayValue  [3]int
	SliceValue  []int
	MapValue    map[string]int
	IntPtrValue *int
	StructValue struct {
		Name string
		Age  int
	}
}

var mixedData = MixedDataTypes{
	IntValue:    42,
	FloatValue:  3.14,
	StringValue: "Hello, World!",
	BoolValue:   true,
	ArrayValue:  [3]int{1, 2, 3},
	SliceValue:  make([]int, 0, 5),
	MapValue:    map[string]int{"one": 1, "two": 2},
	IntPtrValue: func() *int { var temp int = 100; return &temp }(),
	StructValue: struct {
		Name string
		Age  int
	}{"Alice", 30},
}

func Test_TransformOrMap(t *testing.T) {

	st := NewSankeyTransformer(
		WithExpressions(
			"reset()",
			`set("name", "Tom")`,
			`set("friend.last", "Anderson")`,
			`drop("name")`,
			`set("keyFromSrc", $src.StructValue)`,
			`set("keyFromExternalEnv", $externalKey)`,
			`set($externalKey, externalKey)`,
			`set("KeyBuiltinFunc", uuidv4())`,
			`set("KeyCustomFunc", myToInt(paramStringInt))`,
		),
		WithEnvs(map[string]interface{}{
			"$externalKey":   "eValue1",
			"externalKey":    "eValue2",
			"paramStringInt": "123",
		}),
		WithExprOptions(
			expr.Function(
				"myToInt",
				func(params ...any) (any, error) {
					return strconv.Atoi(params[0].(string))
				},
				new(func(string) int),
			),
		),
	)

	m, err := st.Map(mixedData)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(m)
}

func TestSankeyTransformer_Map(t *testing.T) {
	type fields struct {
		expressions []string
		envs        map[string]any
		exprOptions []expr.Option
	}
	type args struct {
		src any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string]any
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SankeyTransformer{
				expressions: tt.fields.expressions,
				envs:        tt.fields.envs,
				exprOptions: tt.fields.exprOptions,
			}
			got, err := s.Map(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Map() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSankeyTransformer_Transform(t *testing.T) {
	type fields struct {
		expressions []string
		envs        map[string]any
		exprOptions []expr.Option
	}
	type args struct {
		src any
		dst any
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SankeyTransformer{
				expressions: tt.fields.expressions,
				envs:        tt.fields.envs,
				exprOptions: tt.fields.exprOptions,
			}
			if err := s.Transform(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("Transform() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
