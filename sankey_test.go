package iosankey

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/expr-lang/expr"
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

func TestSankeyTransformer_Map(t *testing.T) {
	type fields struct {
		expressions []string
		envs        map[string]any
		exprOptions []expr.Option
	}

	tests := []struct {
		fields fields
		want   func(map[string]any) bool
	}{
		{
			fields: fields{
				expressions: []string{},
				envs:        nil,
				exprOptions: nil,
			},
			want: func(m map[string]any) bool {
				return m["FloatValue"].(float64) == 3.14 && m["StringValue"].(string) == "Hello, World!"
			},
		},

		{
			fields: fields{
				expressions: []string{
					"reset()",
				},
				envs:        nil,
				exprOptions: nil,
			},
			want: func(m map[string]any) bool {
				return len(m) == 0
			},
		},
	}

	caseNo := 0
	for _, tt := range tests {
		t.Run("CASE-"+strconv.Itoa(caseNo), func(t *testing.T) {
			s := &SankeyTransformer{
				expressions: tt.fields.expressions,
				envs:        tt.fields.envs,
				exprOptions: tt.fields.exprOptions,
			}
			got, err := s.Map(mixedData)
			if err != nil {
				t.Errorf("Map() error = %v", err)
				return
			}
			if !tt.want(got) {
				t.Errorf("Map() got = %v mismatch", got)
			}
		})

		caseNo++
	}
}

func Test_Readme(t *testing.T) {

	var srcData = MixedDataTypes{
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

	st := NewSankeyTransformer(
		WithExpressions(
			"reset()",
			`set("name", "Tom")`,
			`set("friend.last", "Anderson")`,
			`drop("name")`,
			`set("keyFromSrc", $src.StructValue)`,
			`set("keyFromExternalEnv", $externalKey)`,
			`set($externalKey, externalKey)`,
			`myAccessLog()`,
			`set("KeyBuiltinFunc", uuidv4())`,
			`set("KeyCustomFunc", myToInt(paramStringInt))`,
		),
		WithEnvs(map[string]any{
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
			expr.Function(
				"myAccessLog",
				func(params ...any) (any, error) {
					fmt.Println("access at: " + time.Now().String())
					return nil, nil
				},
				new(func() any),
			),
		),
	)

	_, err := st.Map(srcData)
	if err != nil {
		t.Errorf("error occur. %w", err)
	}

}
