package fuel

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/go-cmp/cmp"
)

var itemEncodeOnlyTests = []struct {
	name string
	in   interface{}
	out  map[string]types.AttributeValue
}{
	{
		name: "omitemptyelem",
		in: struct {
			L     []*string         `dynamodb:",omitemptyelem"`
			SS    []string          `dynamodb:",omitemptyelem,set"`
			M     map[string]string `dynamodb:",omitemptyelem"`
			Other bool
		}{
			L:     []*string{nil, aws.String("")},
			SS:    []string{""},
			M:     map[string]string{"test": ""},
			Other: true,
		},
		out: map[string]types.AttributeValue{
			"L":     &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
			"M":     &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
			"Other": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "omitemptyelem + omitempty",
		in: struct {
			L     []*string         `dynamodb:",omitemptyelem,omitempty"`
			M     map[string]string `dynamodb:",omitemptyelem,omitempty"`
			Other bool
		}{
			L:     []*string{nil, aws.String("")},
			M:     map[string]string{"test": ""},
			Other: true,
		},
		out: map[string]types.AttributeValue{
			"Other": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "unexported field",
		in: struct {
			Public   int
			private  int
			private2 *int
		}{
			Public:   555,
			private:  1337,
			private2: new(int),
		},
		out: map[string]types.AttributeValue{
			"Public": &types.AttributeValueMemberN{Value: "555"},
		},
	},
}

func TestMarshal(t *testing.T) {
	for _, tc := range encodingTests {
		got, err := marshal(tc.in, flagNone)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if diff := cmp.Diff(tc.out, got); diff != "" {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, diff)
		}
	}
}

func TestMarshalItem(t *testing.T) {
	for _, tc := range itemEncodingTests {
		got, err := marshalItem(tc.in)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if diff := cmp.Diff(tc.out, got); diff != "" {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, diff)
		}
	}
}

func TestMarshalItemAsymmetric(t *testing.T) {
	for _, tc := range itemEncodeOnlyTests {
		got, err := marshalItem(tc.in)
		if err != nil {
			t.Errorf("%s: unexpected error: %v", tc.name, err)
			continue
		}

		if diff := cmp.Diff(tc.out, got); diff != "" {
			t.Errorf("%s: missmatch (-want, +got):\n%s", tc.name, diff)
		}
	}
}
