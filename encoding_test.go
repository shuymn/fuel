package fuel

import (
	"encoding"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	maxUint = ^uint(0)
	maxInt  = int(maxUint >> 1)
)

var (
	maxIntStr  = strconv.Itoa(maxInt)
	maxUintStr = strconv.FormatUint(uint64(maxUint), 10)
)

type (
	customString string
	customEmpty  struct{}
)

var encodingTests = []struct {
	name string
	in   interface{}
	out  types.AttributeValue
}{
	{
		name: "strings",
		in:   "hello",
		out:  &types.AttributeValueMemberS{Value: "hello"},
	},
	{
		name: "bools",
		in:   true,
		out:  &types.AttributeValueMemberBOOL{Value: true},
	},
	{
		name: "ints",
		in:   123,
		out:  &types.AttributeValueMemberN{Value: "123"},
	},
	{
		name: "uints",
		in:   uint(123),
		out:  &types.AttributeValueMemberN{Value: "123"},
	},
	{
		name: "floats",
		in:   1.2,
		out:  &types.AttributeValueMemberN{Value: "1.2"},
	},
	{
		name: "pointer (int)",
		in:   new(int),
		out:  &types.AttributeValueMemberN{Value: "0"},
	},
	{
		name: "maps",
		in: map[string]bool{
			"OK": true,
		},
		out: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"OK": &types.AttributeValueMemberBOOL{Value: true},
		}},
	},
	{
		name: "empty maps",
		in: struct {
			Empty map[string]bool // don't omit
			Null  map[string]bool // omit
		}{
			Empty: map[string]bool{},
		},
		out: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"Empty": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{}},
		}},
	},
	{
		name: "textMarshaler maps",
		in: struct {
			M1 map[textMarshaler]bool // don't omit
		}{
			M1: map[textMarshaler]bool{textMarshaler(true): true},
		},
		out: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"M1": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"true": &types.AttributeValueMemberBOOL{Value: true},
			}},
		}},
	},
	{
		name: "struct",
		in: struct {
			OK bool
		}{OK: true},
		out: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
			"OK": &types.AttributeValueMemberBOOL{Value: true},
		}},
	},
	{
		name: "[]byte",
		in:   []byte{'O', 'K'},
		out:  &types.AttributeValueMemberB{Value: []byte{'O', 'K'}},
	},
	{
		name: "slice",
		in:   []int{1, 2, 3},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberN{Value: "1"},
			&types.AttributeValueMemberN{Value: "2"},
			&types.AttributeValueMemberN{Value: "3"},
		}},
	},
	{
		name: "array",
		in:   [3]int{1, 2, 3},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberN{Value: "1"},
			&types.AttributeValueMemberN{Value: "2"},
			&types.AttributeValueMemberN{Value: "3"},
		}},
	},
	{
		name: "byte array",
		in:   [4]byte{'a', 'b', 'c', 'd'},
		out:  &types.AttributeValueMemberB{Value: []byte{'a', 'b', 'c', 'd'}},
	},
	{
		name: "dynamodb.Marshaler",
		in:   customMarshaler(1),
		out:  &types.AttributeValueMemberBOOL{Value: true},
	},
	{
		name: "encoding.TextMarshaler",
		in:   textMarshaler(true),
		out:  &types.AttributeValueMemberS{Value: "true"},
	},
	{
		name: "dynamodb.AttributeValue",
		in: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberN{Value: "1"},
			&types.AttributeValueMemberN{Value: "2"},
			&types.AttributeValueMemberN{Value: "3"},
		}},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberN{Value: "1"},
			&types.AttributeValueMemberN{Value: "2"},
			&types.AttributeValueMemberN{Value: "3"},
		}},
	},
	{
		name: "slice with nil",
		in:   []*int64{nil, aws.Int64(0), nil, aws.Int64(1337), nil},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberNULL{Value: true},
			&types.AttributeValueMemberN{Value: "0"},
			&types.AttributeValueMemberNULL{Value: true},
			&types.AttributeValueMemberN{Value: "1337"},
			&types.AttributeValueMemberNULL{Value: true},
		}},
	},
	{
		name: "array with nil",
		in:   [...]*int64{nil, aws.Int64(0), nil, aws.Int64(1337), nil},
		out: &types.AttributeValueMemberL{
			Value: []types.AttributeValue{
				&types.AttributeValueMemberNULL{Value: true},
				&types.AttributeValueMemberN{Value: "0"},
				&types.AttributeValueMemberNULL{Value: true},
				&types.AttributeValueMemberN{Value: "1337"},
				&types.AttributeValueMemberNULL{Value: true},
			},
		},
	},
	{
		name: "slice with empty string",
		in:   []string{"", "hello", "", "world", ""},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: ""},
			&types.AttributeValueMemberS{Value: "hello"},
			&types.AttributeValueMemberS{Value: ""},
			&types.AttributeValueMemberS{Value: "world"},
			&types.AttributeValueMemberS{Value: ""},
		}},
	},
	{
		name: "array with empty string",
		in:   [...]string{"", "hello", "", "world", ""},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberS{Value: ""},
			&types.AttributeValueMemberS{Value: "hello"},
			&types.AttributeValueMemberS{Value: ""},
			&types.AttributeValueMemberS{Value: "world"},
			&types.AttributeValueMemberS{Value: ""},
		}},
	},
	{
		name: "slice of string pointers",
		in:   []*string{nil, aws.String("hello"), aws.String(""), aws.String("world"), nil},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberNULL{Value: true},
			&types.AttributeValueMemberS{Value: "hello"},
			&types.AttributeValueMemberS{Value: ""},
			&types.AttributeValueMemberS{Value: "world"},
			&types.AttributeValueMemberNULL{Value: true},
		}},
	},
	{
		name: "slice with empty binary",
		in:   [][]byte{{}, []byte("hello"), {}, []byte("world"), {}},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberB{Value: []byte{}},
			&types.AttributeValueMemberB{Value: []byte{'h', 'e', 'l', 'l', 'o'}},
			&types.AttributeValueMemberB{Value: []byte{}},
			&types.AttributeValueMemberB{Value: []byte{'w', 'o', 'r', 'l', 'd'}},
			&types.AttributeValueMemberB{Value: []byte{}},
		}},
	},
	{
		name: "array with empty binary",
		in:   [...][]byte{{}, []byte("hello"), {}, []byte("world"), {}},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberB{Value: []byte{}},
			&types.AttributeValueMemberB{Value: []byte{'h', 'e', 'l', 'l', 'o'}},
			&types.AttributeValueMemberB{Value: []byte{}},
			&types.AttributeValueMemberB{Value: []byte{'w', 'o', 'r', 'l', 'd'}},
			&types.AttributeValueMemberB{Value: []byte{}},
		}},
	},
	{
		name: "array with empty binary ptrs",
		in:   [...]*[]byte{byteSlicePtr([]byte{}), byteSlicePtr([]byte("hello")), nil, byteSlicePtr([]byte("world")), byteSlicePtr([]byte{})},
		out: &types.AttributeValueMemberL{Value: []types.AttributeValue{
			&types.AttributeValueMemberB{Value: []byte{}},
			&types.AttributeValueMemberB{Value: []byte{'h', 'e', 'l', 'l', 'o'}},
			&types.AttributeValueMemberNULL{Value: true},
			&types.AttributeValueMemberB{Value: []byte{'w', 'o', 'r', 'l', 'd'}},
			&types.AttributeValueMemberB{Value: []byte{}},
		}},
	},
}

var itemEncodingTests = []struct {
	name string
	in   interface{}
	out  map[string]types.AttributeValue
}{
	{
		name: "strings",
		in: struct {
			A string
		}{
			A: "hello",
		},
		out: map[string]types.AttributeValue{
			"A": &types.AttributeValueMemberS{Value: "hello"},
		},
	},
	{
		name: "pointer (string)",
		in: &struct {
			A string
		}{
			A: "hello",
		},
		out: map[string]types.AttributeValue{
			"A": &types.AttributeValueMemberS{Value: "hello"},
		},
	},
	{
		name: "pointer (value receiver TextMarshaler)",
		in: &struct {
			A *textMarshaler
		}{
			A: new(textMarshaler),
		},
		out: map[string]types.AttributeValue{
			"A": &types.AttributeValueMemberS{Value: "false"},
		},
	},
	{
		name: "rename",
		in: struct {
			A string `dynamodb:"renamed"`
		}{A: "hello"},
		out: map[string]types.AttributeValue{
			"renamed": &types.AttributeValueMemberS{Value: "hello"},
		},
	},
	{
		name: "skip",
		in: struct {
			A     string `dynamodb:"-"`
			Other bool
		}{
			A:     "",
			Other: true,
		},
		out: map[string]types.AttributeValue{
			"Other": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "omitempty",
		in: struct {
			A       bool       `dynamodb:",omitempty"`
			B       *bool      `dynamodb:",omitempty"`
			NilTime *time.Time `dynamodb:",omitempty"`
			L       []string   `dynamodb:",omitempty"`
			Other   bool
		}{Other: true},
		out: map[string]types.AttributeValue{
			"Other": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "automatic omitempty",
		in: struct {
			OK        string
			EmptyStr  string
			EmptyStr2 customString
			EmptyB    []byte
			EmptyL    []int
			EmptyM    map[string]bool
			EmptyPtr  *int
			NilTime   *time.Time
			NilCustom *customMarshaler
			NilText   *textMarshaler
		}{
			OK:     "OK",
			EmptyL: []int{},
		},
		out: map[string]types.AttributeValue{
			"OK":     &types.AttributeValueMemberS{Value: "OK"},
			"EmptyL": &types.AttributeValueMemberL{Value: []types.AttributeValue{}},
		},
	},
	{
		name: "allowempty flag",
		in: struct {
			S string `dynamodb:",allowempty"`
			B []byte `dynamodb:",allowempty"`
		}{B: []byte{}},
		out: map[string]types.AttributeValue{
			"S": &types.AttributeValueMemberS{Value: ""},
			"B": &types.AttributeValueMemberB{Value: []byte{}},
		},
	},
	{
		name: "allowemptyelem flag",
		in: struct {
			M map[string]*string `dynamodb:",allowemptyelem"`
		}{
			M: map[string]*string{"null": nil, "empty": aws.String(""), "normal": aws.String("hello")},
		},
		out: map[string]types.AttributeValue{
			"M": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"null":   &types.AttributeValueMemberNULL{Value: true},
				"empty":  &types.AttributeValueMemberS{Value: ""},
				"normal": &types.AttributeValueMemberS{Value: "hello"},
			}},
		},
	},
	{
		name: "null flag",
		in: struct {
			S       string             `dynamodb:",null"`
			B       []byte             `dynamodb:",null"`
			NilTime *time.Time         `dynamodb:",null"`
			M       map[string]*string `dynamodb:",null"`
			SS      []string           `dynamodb:",null,set"`
		}{},
		out: map[string]types.AttributeValue{
			"S":       &types.AttributeValueMemberNULL{Value: true},
			"B":       &types.AttributeValueMemberNULL{Value: true},
			"NilTime": &types.AttributeValueMemberNULL{Value: true},
			"M":       &types.AttributeValueMemberNULL{Value: true},
			"SS":      &types.AttributeValueMemberNULL{Value: true},
		},
	},
	{
		name: "embedded struct",
		in: struct {
			embedded
		}{
			embedded: embedded{
				Embedded: true,
			},
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "exported embedded struct",
		in: struct {
			ExportedEmbedded
		}{
			ExportedEmbedded: ExportedEmbedded{
				Embedded: true,
			},
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "exported pointer embedded struct",
		in: struct {
			*ExportedEmbedded
		}{
			ExportedEmbedded: &ExportedEmbedded{
				Embedded: true,
			},
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberBOOL{Value: true},
		},
	},
	{
		name: "embedded struct clobber",
		in: struct {
			Embedded string
			embedded
		}{
			Embedded: "OK",
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberS{Value: "OK"},
		},
	},
	{
		name: "pointer embedded struct clobber",
		in: struct {
			Embedded string
			*embedded
		}{
			Embedded: "OK",
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberS{Value: "OK"},
		},
	},
	{
		name: "exported embedded struct clobber",
		in: struct {
			Embedded string
			ExportedEmbedded
		}{
			Embedded: "OK",
		},
		out: map[string]types.AttributeValue{
			"Embedded": &types.AttributeValueMemberS{Value: "OK"},
		},
	},
	{
		name: "sets",
		in: struct {
			SS1  []string                   `dynamodb:",set"`
			SS2  []textMarshaler            `dynamodb:",set"`
			SS3  map[string]struct{}        `dynamodb:",set"`
			SS4  map[string]bool            `dynamodb:",set"`
			SS5  map[customString]struct{}  `dynamodb:",set"`
			SS6  []customString             `dynamodb:",set"`
			SS7  map[textMarshaler]struct{} `dynamodb:",set"`
			SS8  map[textMarshaler]bool     `dynamodb:",set"`
			SS9  []string                   `dynamodb:",set"`
			SS10 map[string]customEmpty     `dynamodb:",set"`
			BS1  [][]byte                   `dynamodb:",set"`
			BS2  map[[1]byte]struct{}       `dynamodb:",set"`
			BS3  map[[1]byte]bool           `dynamodb:",set"`
			BS4  [][]byte                   `dynamodb:",set"`
			NS1  []int                      `dynamodb:",set"`
			NS2  []float64                  `dynamodb:",set"`
			NS3  []uint                     `dynamodb:",set"`
			NS4  map[int]struct{}           `dynamodb:",set"`
			NS5  map[uint]bool              `dynamodb:",set"`
		}{
			SS1:  []string{"A", "B"},
			SS2:  []textMarshaler{textMarshaler(true), textMarshaler(false)},
			SS3:  map[string]struct{}{"A": {}},
			SS4:  map[string]bool{"A": true},
			SS5:  map[customString]struct{}{"A": {}},
			SS6:  []customString{"A", "B"},
			SS7:  map[textMarshaler]struct{}{textMarshaler(true): {}},
			SS8:  map[textMarshaler]bool{textMarshaler(false): true},
			SS9:  []string{"A", "B", ""},
			SS10: map[string]customEmpty{"A": {}},
			BS1:  [][]byte{{'A'}, {'B'}},
			BS2:  map[[1]byte]struct{}{{'A'}: {}},
			BS3:  map[[1]byte]bool{{'A'}: true},
			BS4:  [][]byte{{'A'}, {'B'}, {}},
			NS1:  []int{1, 2},
			NS2:  []float64{1, 2},
			NS3:  []uint{1, 2},
			NS4:  map[int]struct{}{maxInt: {}},
			NS5:  map[uint]bool{maxUint: true},
		},
		out: map[string]types.AttributeValue{
			"SS1":  &types.AttributeValueMemberSS{Value: []string{"A", "B"}},
			"SS2":  &types.AttributeValueMemberSS{Value: []string{"true", "false"}},
			"SS3":  &types.AttributeValueMemberSS{Value: []string{"A"}},
			"SS4":  &types.AttributeValueMemberSS{Value: []string{"A"}},
			"SS5":  &types.AttributeValueMemberSS{Value: []string{"A"}},
			"SS6":  &types.AttributeValueMemberSS{Value: []string{"A", "B"}},
			"SS7":  &types.AttributeValueMemberSS{Value: []string{"true"}},
			"SS8":  &types.AttributeValueMemberSS{Value: []string{"false"}},
			"SS9":  &types.AttributeValueMemberSS{Value: []string{"A", "B", ""}},
			"SS10": &types.AttributeValueMemberSS{Value: []string{"A"}},
			"BS1":  &types.AttributeValueMemberBS{Value: [][]byte{{'A'}, {'B'}}},
			"BS2":  &types.AttributeValueMemberBS{Value: [][]byte{{'A'}}},
			"BS3":  &types.AttributeValueMemberBS{Value: [][]byte{{'A'}}},
			"BS4":  &types.AttributeValueMemberBS{Value: [][]byte{{'A'}, {'B'}, {}}},
			"NS1":  &types.AttributeValueMemberNS{Value: []string{"1", "2"}},
			"NS2":  &types.AttributeValueMemberNS{Value: []string{"1", "2"}},
			"NS3":  &types.AttributeValueMemberNS{Value: []string{"1", "2"}},
			"NS4":  &types.AttributeValueMemberNS{Value: []string{maxIntStr}},
			"NS5":  &types.AttributeValueMemberNS{Value: []string{maxUintStr}},
		},
	},
	{
		name: "map as item",
		in: map[string]interface{}{
			"S": "Hello",
			"B": []byte{'A', 'B'},
			"N": float64(1.2),
			"L": []interface{}{"A", "B", 1.2},
			"M": map[string]interface{}{"OK": true},
		},
		out: map[string]types.AttributeValue{
			"S": &types.AttributeValueMemberS{Value: "Hello"},
			"B": &types.AttributeValueMemberB{Value: []byte{'A', 'B'}},
			"N": &types.AttributeValueMemberN{Value: "1.2"},
			"L": &types.AttributeValueMemberL{
				Value: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: "A"},
					&types.AttributeValueMemberS{Value: "B"},
					&types.AttributeValueMemberN{Value: "1.2"},
				},
			},
			"M": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"OK": &types.AttributeValueMemberBOOL{Value: true},
			}},
		},
	},
	{
		name: "map as key",
		in: struct {
			M map[string]interface{}
		}{
			M: map[string]interface{}{"Hello": "world"},
		},
		out: map[string]types.AttributeValue{
			"M": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"Hello": &types.AttributeValueMemberS{Value: "world"},
			}},
		},
	},
	{
		name: "map string attributevalue",
		in: map[string]types.AttributeValue{
			"M": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"Hello": &types.AttributeValueMemberS{Value: "world"},
			}},
		},
		out: map[string]types.AttributeValue{
			"M": &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"Hello": &types.AttributeValueMemberS{Value: "world"},
			}},
		},
	},
	{
		name: "time.Time (regular encoding)",
		in: struct {
			TTL time.Time
		}{
			TTL: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		out: map[string]types.AttributeValue{
			"TTL": &types.AttributeValueMemberS{Value: "2019-01-01T00:00:00Z"},
		},
	},
	{
		name: "time.Time (unixtime encoding)",
		in: struct {
			TTL time.Time `dynamodb:",unixtime"`
		}{
			TTL: time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		out: map[string]types.AttributeValue{
			"TTL": &types.AttributeValueMemberN{Value: "1546300800"},
		},
	},
	{
		name: "time.Time (zero unixtime encoding)",
		in: struct {
			TTL time.Time `dynamodb:",unixtime"`
		}{
			TTL: time.Time{},
		},
		out: map[string]types.AttributeValue{},
	},
	{
		name: "*time.Time (unixtime encoding)",
		in: struct {
			TTL *time.Time `dynamodb:",unixtime"`
		}{
			TTL: aws.Time(time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		out: map[string]types.AttributeValue{
			"TTL": &types.AttributeValueMemberN{Value: "1546300800"},
		},
	},
	{
		name: "*time.Time (zero unixtime encoding)",
		in: struct {
			TTL *time.Time `dynamodb:",unixtime"`
		}{
			TTL: nil,
		},
		out: map[string]types.AttributeValue{},
	},
	{
		name: "dynamodb.ItemUnmarshaler",
		in:   customItemMarshaler{Thing: 52},
		out: map[string]types.AttributeValue{
			"thing": &types.AttributeValueMemberN{Value: "52"},
		},
	},
	{
		name: "*dynamodb.ItemUnmarshaler",
		in:   &customItemMarshaler{Thing: 52},
		out: map[string]types.AttributeValue{
			"thing": &types.AttributeValueMemberN{Value: "52"},
		},
	},
}

type embedded struct {
	Embedded bool
}

type ExportedEmbedded struct {
	Embedded bool
}

type customMarshaler int

func (cm customMarshaler) MarshalDynamoDB() (types.AttributeValue, error) {
	return &types.AttributeValueMemberBOOL{
		Value: cm != 0,
	}, nil
}

func (cm *customMarshaler) UnmarshalDynamoDB(av types.AttributeValue) error {
	avBOOL, ok := av.(*types.AttributeValueMemberBOOL)
	if ok && avBOOL.Value == true {
		*cm = 1
	}
	return nil
}

type textMarshaler bool

func (tm textMarshaler) MarshalText() ([]byte, error) {
	if tm {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (tm *textMarshaler) UnmarshalText(text []byte) error {
	*tm = string(text) == "true"
	return nil
}

type textMarshalerPtr bool

func (tm *textMarshalerPtr) MarshalText() ([]byte, error) {
	if tm == nil {
		return []byte("null"), nil
	}
	if *tm {
		return []byte("true"), nil
	}
	return []byte("false"), nil
}

func (tm *textMarshalerPtr) UnmarshalText(text []byte) error {
	if string(text) == "null" {
		return nil
	}
	*tm = string(text) == "true"
	return nil
}

type customItemMarshaler struct {
	Thing interface{} `dynamodb:"thing"`
}

func (cim *customItemMarshaler) MarshalDynamoDBItem() (map[string]types.AttributeValue, error) {
	thing := strconv.Itoa(cim.Thing.(int))
	attrs := map[string]types.AttributeValue{
		"thing": &types.AttributeValueMemberN{
			Value: thing,
		},
	}
	return attrs, nil
}

func (cim *customItemMarshaler) UnmarshalDynamoDBItem(item map[string]types.AttributeValue) error {
	thingAttr := item["thing"]

	if thingAttr != nil {
		if thingAttrN, ok := thingAttr.(*types.AttributeValueMemberN); ok {
			thing, err := strconv.Atoi(thingAttrN.Value)
			if err != nil {
				return fmt.Errorf("Invalid number")
			}
			cim.Thing = thing
			return nil
		}
	}
	return fmt.Errorf("Missing or not a number")
}

func byteSlicePtr(a []byte) *[]byte {
	return &a
}

var (
	_ Marshaler                = new(customMarshaler)
	_ Unmarshaler              = new(customMarshaler)
	_ ItemMarshaler            = new(customItemMarshaler)
	_ ItemUnmarshaler          = new(customItemMarshaler)
	_ encoding.TextMarshaler   = new(textMarshaler)
	_ encoding.TextUnmarshaler = new(textMarshaler)
	_ encoding.TextMarshaler   = new(textMarshalerPtr)
	_ encoding.TextUnmarshaler = new(textMarshalerPtr)
)
