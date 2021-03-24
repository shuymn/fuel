package fuel

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
)

var itemDecodeOnlyTests = []struct {
	name  string
	given map[string]types.AttributeValue
	want  interface{}
}{
	{
		// unexported embedded pointers should be ignored
		name: "embedded unexported pointer",
		given: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberBOOL{Value: true},
		},
		want: struct {
			*embedded
		}{},
	},
	{
		// unexported fields should be ignored
		name: "unexported fields",
		given: map[string]types.AttributeValue{
			"a": &types.AttributeValueMemberBOOL{Value: true},
		},
		want: struct {
			a bool
		}{},
	},
	{
		// embedded pointers shouldn't clobber existing fields
		name: "exported pointer embedded struct clobber",
		given: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberS{Value: "OK"},
		},
		want: struct {
			Embedded string
			*ExportedEmbedded
		}{
			Embedded:         "OK",
			ExportedEmbedded: &ExportedEmbedded{},
		},
	},
}

func TestUnmarshalAsymmetric(t *testing.T) {
	for _, tc := range itemDecodeOnlyTests {
		rv := reflect.New(reflect.TypeOf(tc.want))
		got := rv.Interface()
		if err := UnmarshalItem(tc.given, got); err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		opt := cmp.AllowUnexported(tc.want)
		if diff := cmp.Diff(tc.want, rv.Elem().Interface(), opt); diff != "" {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, diff)
		}
	}
}

func TestUnmarshalAppend(t *testing.T) {
	var results []struct {
		User  int `dynamodb:"UserID"`
		Page  int
		Limit uint
		Null  interface{}
	}
	id := "12345"
	page := "5"
	limit := "20"
	null := true
	item := map[string]types.AttributeValue{
		"UserID": &types.AttributeValueMemberN{Value: id},
		"Page":   &types.AttributeValueMemberN{Value: page},
		"Limit":  &types.AttributeValueMemberN{Value: limit},
		"Null":   &types.AttributeValueMemberNULL{Value: null},
	}

	for range [15]struct{}{} {
		if err := unmarshalAppend(item, &results); err != nil {
			t.Fatal(err)
		}
	}

	for _, h := range results {
		if h.User != 12345 || h.Page != 5 || h.Limit != 20 || h.Null != nil {
			t.Error("invalid hit", h)
		}
	}

	var mapResults []map[string]interface{}

	for range [15]struct{}{} {
		err := unmarshalAppend(item, &mapResults)
		if err != nil {
			t.Fatal(err)
		}
	}

	for _, h := range mapResults {
		if h["UserID"] != 12345.0 || h["Page"] != 5.0 || h["Limit"] != 20.0 || h["Null"] != nil {
			t.Error("invalid interface{} hit", h)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for _, tc := range encodingTests {
		rv := reflect.New(reflect.TypeOf(tc.in))
		if err := unmarshalReflect(tc.out, rv.Elem()); err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if want, got := tc.in, rv.Elem().Interface(); !cmp.Equal(want, got) {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, cmp.Diff(want, got))
		}
	}
}

func TestUnmarshalItem(t *testing.T) {
	for _, tc := range itemEncodingTests {
		rv := reflect.New(reflect.TypeOf(tc.in))
		if err := unmarshalItem(tc.out, rv.Interface()); err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		var opt cmp.Option
		if reflect.TypeOf(tc.in).Kind() == reflect.Struct {
			opt = cmp.AllowUnexported(tc.in)
		}
		if want, got := tc.in, rv.Elem().Interface(); !cmp.Equal(want, got, opt) {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, cmp.Diff(want, got, opt))
		}
	}
}

func TestUnmarshalNULL(t *testing.T) {
	tru := true
	arbitrary := "hello world"
	double := new(*int)
	item := map[string]types.AttributeValue{
		"String":    &types.AttributeValueMemberNULL{Value: tru},
		"Slice":     &types.AttributeValueMemberNULL{Value: tru},
		"Array":     &types.AttributeValueMemberNULL{Value: tru},
		"StringPtr": &types.AttributeValueMemberNULL{Value: tru},
		"DoublePtr": &types.AttributeValueMemberNULL{Value: tru},
		"Map":       &types.AttributeValueMemberNULL{Value: tru},
		"Interface": &types.AttributeValueMemberNULL{Value: tru},
	}

	type resultType struct {
		String    string
		Slice     []string
		Array     [2]byte
		StringPtr *string
		DoublePtr **int
		Map       map[string]int
		Interface interface{}
	}

	// dirty result, we want this to be reset
	result := resultType{
		String:    "ABC",
		Slice:     []string{"A", "B"},
		Array:     [2]byte{'A', 'B'},
		StringPtr: &arbitrary,
		DoublePtr: double,
		Map: map[string]int{
			"A": 1,
		},
		Interface: "interface{}",
	}

	if err := UnmarshalItem(item, &result); err != nil {
		t.Error(err)
		return
	}

	if diff := cmp.Diff(resultType{}, result); diff != "" {
		t.Errorf("unmarshal null: missmatch (-want, +got):\n%s", diff)
	}
}
