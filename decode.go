package fuel

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Unmarshaler is the interface implemented by objects that can unmarshal
// an AttributeValue into themselves
type Unmarshaler interface {
	UnmarshalDynamoDB(av types.AttributeValue) error
}

// ItemUnmarshaler is the interface implemented by objects that can unmarshal
// an Item (a map of strings to AttributeValues) into themselves
type ItemUnmarshaler interface {
	UnmarshalDynamoDBItem(item map[string]types.AttributeValue) error
}

func UnmarshalAppend(item map[string]types.AttributeValue, out interface{}) error {
	return unmarshalAppend(item, out)
}

func UnmarshalItem(item map[string]types.AttributeValue, out interface{}) error {
	return unmarshalItem(item, out)
}

func Unmarshal(av types.AttributeValue, out interface{}) error {
	rv := reflect.ValueOf(out)
	return unmarshalReflect(av, rv)
}

var (
	nilTum  encoding.TextUnmarshaler
	tumType = reflect.TypeOf(&nilTum).Elem()
)

// unmarshal one value
func unmarshalReflect(av types.AttributeValue, rv reflect.Value) error {
	// first try interface unmarshal stuff
	if rv.CanInterface() {
		var iface interface{}
		if rv.CanAddr() {
			iface = rv.Addr().Interface()
		} else {
			iface = rv.Interface()
		}

		if x, ok := iface.(*time.Time); ok {
			if avN, ok := av.(*types.AttributeValueMemberN); ok {
				// implicit unixtime
				// TODO: have unixtime unmarshal explicitly check struct tasgs
				ts, err := strconv.ParseInt(avN.Value, 10, 64)
				if err != nil {
					return err
				}

				*x = time.Unix(ts, 0).UTC()
				return nil
			}
		}

		switch x := iface.(type) {
		case *types.AttributeValueMemberBOOL:
			switch y := av.(type) {
			case *types.AttributeValueMemberBOOL:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberB:
			switch y := av.(type) {
			case *types.AttributeValueMemberB:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberBS:
			switch y := av.(type) {
			case *types.AttributeValueMemberBS:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberL:
			switch y := av.(type) {
			case *types.AttributeValueMemberL:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberM:
			switch y := av.(type) {
			case *types.AttributeValueMemberM:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberN:
			switch y := av.(type) {
			case *types.AttributeValueMemberN:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberNS:
			switch y := av.(type) {
			case *types.AttributeValueMemberNS:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberNULL:
			switch y := av.(type) {
			case *types.AttributeValueMemberNULL:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberS:
			switch y := av.(type) {
			case *types.AttributeValueMemberS:
				*x = *y
				return nil
			}
		case *types.AttributeValueMemberSS:
			switch y := av.(type) {
			case *types.AttributeValueMemberSS:
				*x = *y
				return nil
			}
		case Unmarshaler:
			return x.UnmarshalDynamoDB(av)
		case encoding.TextUnmarshaler:
			if avS, ok := av.(*types.AttributeValueMemberS); ok {
				return x.UnmarshalText([]byte(avS.Value))
			}
		}
	}

	if !rv.CanSet() {
		return nil
	}

	if _, ok := av.(*types.AttributeValueMemberNULL); ok {
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}

	switch rv.Kind() {
	case reflect.Ptr:
		pt := reflect.New(rv.Type().Elem())
		rv.Set(pt)
		if avNULL, ok := av.(*types.AttributeValueMemberNULL); !ok || !(avNULL.Value) {
			return unmarshalReflect(av, rv.Elem())
		}
		return nil
	case reflect.Bool:
		avBOOL, ok := av.(*types.AttributeValueMemberBOOL)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into bool", avTypeName(av))
		}
		rv.SetBool(avBOOL.Value)
		return nil
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		avN, ok := av.(*types.AttributeValueMemberN)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into int", avTypeName(av))
		}
		n, err := strconv.ParseInt(avN.Value, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(n)
		return nil
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		avN, ok := av.(*types.AttributeValueMemberN)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into uint", avTypeName(av))
		}
		n, err := strconv.ParseUint(avN.Value, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(n)
		return nil
	case reflect.Float64, reflect.Float32:
		avN, ok := av.(*types.AttributeValueMemberN)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into float", avTypeName(av))
		}
		n, err := strconv.ParseFloat(avN.Value, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(n)
		return nil
	case reflect.String:
		avS, ok := av.(*types.AttributeValueMemberS)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into string", avTypeName(av))
		}
		rv.SetString(avS.Value)
		return nil
	case reflect.Struct:
		avM, ok := av.(*types.AttributeValueMemberM)
		if !ok {
			return fmt.Errorf("dynamodb: cannot unmarshal %s data into struct", avTypeName(av))
		}
		if err := unmarshalItem(avM.Value, rv.Addr().Interface()); err != nil {
			return err
		}
		return nil

	case reflect.Map:
		if rv.IsNil() {
			// TODO: maybe always remake this?
			// I think the JSON library doesn't ...
			rv.Set(reflect.MakeMap(rv.Type()))
		}

		var truthy reflect.Value
		switch {
		case rv.Type().Elem().Kind() == reflect.Bool:
			truthy = reflect.ValueOf(true)
		case rv.Type().Elem() == emptyStructType:
			fallthrough
		case rv.Type().Elem().Kind() == reflect.Struct && rv.Type().Elem().NumField() == 0:
			truthy = reflect.ValueOf(struct{}{})
		default:
			if _, ok := av.(*types.AttributeValueMemberM); !ok {
				return fmt.Errorf("dynamodb: unmarshal map set: value type must be struct{} or bool, got %v", rv.Type())
			}
		}

		switch x := av.(type) {
		case *types.AttributeValueMemberM:
			// TODO: this is probably slow
			kp := reflect.New(rv.Type().Key())
			kv := kp.Elem()
			for k, v := range x.Value {
				innerRV := reflect.New(rv.Type().Elem())
				if err := unmarshalReflect(v, innerRV.Elem()); err != nil {
					return err
				}
				if kp.Type().Implements(tumType) {
					tm := kp.Interface().(encoding.TextUnmarshaler)
					if err := tm.UnmarshalText([]byte(k)); err != nil {
						return fmt.Errorf("dynamodb: unmarshal map: key error: %v", err)
					}
				} else {
					kv.SetString(k)
				}
				rv.SetMapIndex(kv, innerRV.Elem())
			}
			return nil
		case *types.AttributeValueMemberSS:
			kp := reflect.New(rv.Type().Key())
			kv := kp.Elem()
			for _, s := range x.Value {
				if kp.Type().Implements(tumType) {
					tm := kp.Interface().(encoding.TextUnmarshaler)
					if err := tm.UnmarshalText([]byte(s)); err != nil {
						return fmt.Errorf("dynamodb: unmarshal map (SS): key error: %v", err)
					}
				} else {
					kv.SetString(s)
				}
				rv.SetMapIndex(kv, truthy)
			}
			return nil
		case *types.AttributeValueMemberNS:
			kv := reflect.New(rv.Type().Key()).Elem()
			for _, n := range x.Value {
				if err := unmarshalReflect(&types.AttributeValueMemberN{Value: n}, kv); err != nil {
					return err
				}
				rv.SetMapIndex(kv, truthy)
			}
			return nil
		case *types.AttributeValueMemberBS:
			for _, bb := range x.Value {
				kv := reflect.New(rv.Type().Key()).Elem()
				reflect.Copy(kv, reflect.ValueOf(bb))
				rv.SetMapIndex(kv, truthy)
			}
			return nil
		default:
			return fmt.Errorf("dynamodb: cannot unmarshal %s vdata into map", avTypeName(av))
		}
	case reflect.Slice:
		return unmarshalSlice(av, rv)
	case reflect.Array:
		arr := reflect.New(rv.Type()).Elem()
		elemType := arr.Type().Elem()
		switch x := av.(type) {
		case *types.AttributeValueMemberB:
			if len(x.Value) > arr.Len() {
				return fmt.Errorf("dynamodb: cannot marshal %s into %s; too small (dst len: %d, src len: %d)", avTypeName(av), arr.Type().String(), arr.Len(), len(x.Value))
			}
			reflect.Copy(arr, reflect.ValueOf(x.Value))
			rv.Set(arr)
			return nil
		case *types.AttributeValueMemberL:
			if len(x.Value) > arr.Len() {
				return fmt.Errorf("dynamodb: cannot marshal %s into %s; too small (dst len: %d, src len: %d)", avTypeName(av), arr.Type().String(), arr.Len(), len(x.Value))
			}
			for i, innerAV := range x.Value {
				innerRV := reflect.New(elemType).Elem()
				if err := unmarshalReflect(innerAV, innerRV); err != nil {
					return nil
				}
				arr.Index(i).Set(innerRV)
			}
			rv.Set(arr)
			return nil
		}
	case reflect.Interface:
		if rv.NumMethod() == 0 {
			iface, err := av2iface(av)
			if err != nil {
				return err
			}
			if iface == nil {
				rv.Set(reflect.Zero(rv.Type()))
			} else {
				rv.Set(reflect.ValueOf(iface))
			}
			return nil
		}
	}

	iface := rv.Interface()
	return fmt.Errorf("dynamodb: cannot unmarshal to type: %T (%+v)", iface, iface)
}

func unmarshalSlice(av types.AttributeValue, rv reflect.Value) error {
	switch x := av.(type) {
	case *types.AttributeValueMemberB:
		rv.SetBytes(x.Value)
		return nil
	case *types.AttributeValueMemberL:
		slicev := reflect.MakeSlice(rv.Type(), 0, len(x.Value))
		for _, innerAV := range x.Value {
			innerRV := reflect.New(rv.Type().Elem()).Elem()
			if err := unmarshalReflect(innerAV, innerRV); err != nil {
				return err
			}
			slicev = reflect.Append(slicev, innerRV)
		}
		rv.Set(slicev)
		return nil

	// there's brobably a better way to do these
	case *types.AttributeValueMemberBS:
		slicev := reflect.MakeSlice(rv.Type(), 0, len(x.Value))
		for _, b := range x.Value {
			innerRV := reflect.New(rv.Type().Elem()).Elem()
			if err := unmarshalReflect(&types.AttributeValueMemberB{Value: b}, innerRV); err != nil {
				return err
			}
			slicev = reflect.Append(slicev, innerRV)
		}
		rv.Set(slicev)
		return nil
	case *types.AttributeValueMemberSS:
		slicev := reflect.MakeSlice(rv.Type(), 0, len(x.Value))
		for _, str := range x.Value {
			innerRV := reflect.New(rv.Type().Elem()).Elem()
			if err := unmarshalReflect(&types.AttributeValueMemberS{Value: str}, innerRV); err != nil {
				return err
			}
			slicev = reflect.Append(slicev, innerRV)
		}
		rv.Set(slicev)
		return nil
	case *types.AttributeValueMemberNS:
		slicev := reflect.MakeSlice(rv.Type(), 0, len(x.Value))
		for _, n := range x.Value {
			innerRV := reflect.New(rv.Type().Elem()).Elem()
			if err := unmarshalReflect(&types.AttributeValueMemberN{Value: n}, innerRV); err != nil {
				return nil
			}
			slicev = reflect.Append(slicev, innerRV)
		}
		rv.Set(slicev)
		return nil
	}
	return fmt.Errorf("dynamodb: cannot unmarshal %s data into slice", avTypeName(av))
}

func fieldsInStruct(rv reflect.Value) map[string]reflect.Value {
	if rv.Kind() == reflect.Ptr {
		return fieldsInStruct(rv.Elem())
	}

	fields := make(map[string]reflect.Value)
	for i := 0; i < rv.Type().NumField(); i++ {
		field := rv.Type().Field(i)
		fv := rv.Field(i)
		isPtr := fv.Type().Kind() == reflect.Ptr

		name, _ := fieldInfo(field)
		if name == "-" {
			// skip
			continue
		}

		// embed anonymous structs, they could be pointers so test that too
		if (fv.Type().Kind() == reflect.Struct || isPtr && fv.Type().Elem().Kind() == reflect.Struct) && field.Anonymous {
			if isPtr {
				// need to protect from setting unexported pointers because it will panic
				if !fv.CanSet() {
					continue
				}
				// set zero value for pointer
				zero := reflect.New(fv.Type().Elem())
				fv.Set(zero)
				fv = zero
			}

			innerFields := fieldsInStruct(fv)
			for k, v := range innerFields {
				// don't clobber top-level fields
				if _, ok := fields[k]; ok {
					continue
				}
				fields[k] = v
			}
			continue
		}
		fields[name] = fv
	}
	return fields
}

func unmarshalItem(item map[string]types.AttributeValue, out interface{}) error {
	switch x := out.(type) {
	case *map[string]types.AttributeValue:
		*x = item
		return nil
	case awsEncoder:
		return fmt.Errorf("dynamodb: unimplemented: aws encoder")
	case ItemUnmarshaler:
		return x.UnmarshalDynamoDBItem(item)
	}

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("dynamodb: unmarshal: not a pointer: %T", out)
	}

	switch rv.Elem().Kind() {
	case reflect.Ptr:
		rv.Elem().Set(reflect.New(rv.Elem().Type().Elem()))
		return unmarshalItem(item, rv.Elem().Interface())
	case reflect.Struct:
		var err error
		rv.Elem().Set(reflect.Zero(rv.Type().Elem()))
		fields := fieldsInStruct(rv.Elem())
		for name, fv := range fields {
			if av, ok := item[name]; ok {
				if innerErr := unmarshalReflect(av, fv); innerErr != nil {
					err = innerErr
				}
			}
		}
		return err
	case reflect.Map:
		mapv := rv.Elem()
		if mapv.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("dynamodb: unmarshal: map key must be a string: %T", mapv.Interface())
		}
		if mapv.IsNil() {
			mapv.Set(reflect.MakeMap(mapv.Type()))
		}

		for k, av := range item {
			innerRV := reflect.New(mapv.Type().Elem()).Elem()
			if err := unmarshalReflect(av, innerRV); err != nil {
				return err
			}
			mapv.SetMapIndex(reflect.ValueOf(k), innerRV)
		}
		return nil
	}
	return fmt.Errorf("dynamodb: unmarshal: unsupported type: %T", out)
}

func unmarshalAppend(item map[string]types.AttributeValue, out interface{}) error {
	if _, ok := out.(awsEncoder); ok {
		return fmt.Errorf("dynamodb: unimplemented: aws encoder")
	}

	rv := reflect.ValueOf(out)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dynamodb: unmarshal append: result argument must be a slice pointer")
	}

	slicev := rv.Elem()
	innerRV := reflect.New(slicev.Type().Elem())
	if err := unmarshalItem(item, innerRV.Interface()); err != nil {
		return err
	}
	slicev = reflect.Append(slicev, innerRV.Elem())

	rv.Elem().Set(slicev)
	return nil
}

// av2iface converts an AttributeValue into interface{}
func av2iface(av types.AttributeValue) (interface{}, error) {
	switch x := av.(type) {
	case *types.AttributeValueMemberB:
		return x.Value, nil
	case *types.AttributeValueMemberBS:
		return x.Value, nil
	case *types.AttributeValueMemberBOOL:
		return x.Value, nil
	case *types.AttributeValueMemberN:
		return strconv.ParseFloat(x.Value, 64)
	case *types.AttributeValueMemberS:
		return x.Value, nil
	case *types.AttributeValueMemberL:
		list := make([]interface{}, 0, len(x.Value))
		for _, item := range x.Value {
			iface, err := av2iface(item)
			if err != nil {
				return nil, err
			}
			list = append(list, iface)
		}
		return list, nil
	case *types.AttributeValueMemberNS:
		set := make([]float64, 0, len(x.Value))
		for _, n := range x.Value {
			f, err := strconv.ParseFloat(n, 64)
			if err != nil {
				return nil, err
			}
			set = append(set, f)
		}
		return set, nil
	case *types.AttributeValueMemberSS:
		set := make([]string, 0, len(x.Value))
		set = append(set, x.Value...)
		return set, nil
	case *types.AttributeValueMemberM:
		m := make(map[string]interface{}, len(x.Value))
		for k, v := range x.Value {
			iface, err := av2iface(v)
			if err != nil {
				return nil, err
			}
			m[k] = iface
		}
		return m, nil
	case *types.AttributeValueMemberNULL:
		return nil, nil
	case *types.UnknownUnionMember:
		return nil, fmt.Errorf("dynamodb: unknown tag: %s", x.Tag)
	}
	return nil, fmt.Errorf("dynamodb: unsupported attribute value: %#v", av)
}

func avTypeName(av types.AttributeValue) string {
	switch av.(type) {
	case *types.AttributeValueMemberB:
		return "binary"
	case *types.AttributeValueMemberBOOL:
		return "boolean"
	case *types.AttributeValueMemberBS:
		return "binary set"
	case *types.AttributeValueMemberL:
		return "list"
	case *types.AttributeValueMemberM:
		return "map"
	case *types.AttributeValueMemberN:
		return "number"
	case *types.AttributeValueMemberNS:
		return "number set"
	case *types.AttributeValueMemberNULL:
		return "null"
	case *types.AttributeValueMemberS:
		return "string"
	case *types.AttributeValueMemberSS:
		return "string set"
	case *types.UnknownUnionMember:
		return "unknown"
	default:
		return "<empty>"
	}
}
