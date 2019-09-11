// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//Package navro walk a native struct value and generate usable avro schemas
package navro

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	defaultTag = "avro"
)

//NotSupported indicate unsupportive error type
type NotSupported interface {
	isNotSupport()
}

type errorSupport struct {
	msg string
}

func (e errorSupport) Error() string {
	return e.msg
}

func (e errorSupport) isNotSupport() {
}

// Generate creates new schema for whatever interface i is
// works same as Generate but less code, easier to read
// hope it works fine
func Generate(i interface{}) (string, map[string]bool, error) {
	recordSchemas := make(map[string]*RecordSchema)
	schema, err := schemaByType(reflect.TypeOf(i), recordSchemas)
	if err != nil {
		return "", nil, err
	}
	recordNames := make(map[string]bool)
	for recordName := range recordSchemas {
		recordNames[recordName] = true
	}
	return schema.String(), recordNames, nil
}

func schemaByType(t reflect.Type, schemas map[string]*RecordSchema) (Schema, error) {
	switch t.Kind() {
	case reflect.Invalid:
		return new(NullSchema), nil
	case reflect.Bool:
		return new(BooleanSchema), nil
	case reflect.Uint8, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint16: // <= 32 bits will be mapped to Int avro
		return new(IntSchema), nil
	case reflect.Int, reflect.Uint32, reflect.Int64, reflect.Uint64: // cause in Go, int could be 32 or 64 bit
		return new(LongSchema), nil
	case reflect.String:
		return new(StringSchema), nil
	case reflect.Float32:
		return new(FloatSchema), nil
	case reflect.Float64:
		return new(DoubleSchema), nil
	case reflect.Array, reflect.Slice:
		// special case for []byte => bytes avro schema
		if t.Elem().Kind() == reflect.Uint8 {
			return new(BytesSchema), nil
		} // end of hack

		elemSchema, err := schemaByType(t.Elem(), schemas)
		if err != nil {
			return nil, err
		}
		schema := &ArraySchema{
			Items: elemSchema,
		}
		return schema, nil
	case reflect.Map: // key must be string
		if t.Key().Kind() != reflect.String {
			return nil, errorSupport{"Do not support map with non-string key"}
		}
		elemSchema, err := schemaByType(t.Elem(), schemas)
		if err != nil {
			return nil, err
		}
		schema := &MapSchema{
			Values: elemSchema,
		}
		return schema, nil
	case reflect.Struct:
		if schemas[t.Name()] != nil {
			return &RecursiveSchema{Actual: schemas[t.Name()]}, nil
		}

		schema := &RecordSchema{
			Name:   t.Name(),
			Fields: make([]*SchemaField, 0, t.NumField()),
		}
		for i := 0; i < t.NumField(); i++ {
			structField := t.Field(i)
			fieldTypeSchema, err := schemaByType(structField.Type, schemas)
			if err != nil {
				return nil, err
			}
			fieldSchema := &SchemaField{
				Name: getNameByJSONTag(structField),
				Type: fieldTypeSchema,
			}
			schema.Fields = append(schema.Fields, fieldSchema)
		}
		schemas[t.Name()] = schema
		return schema, nil
	case reflect.Ptr:
		elemSchema, err := schemaByType(t.Elem(), schemas)
		if err != nil {
			return nil, err
		}
		schema := &UnionSchema{
			Types: []Schema{new(NullSchema), elemSchema},
		}
		return schema, nil
	default:
		return nil, errorSupport{fmt.Sprintf("%s will not be supported", t.Kind().String())}
	}
}

func getNameByJSONTag(st reflect.StructField) string {
	tagValue, exists := st.Tag.Lookup(defaultTag)
	if !exists {
		return st.Name
	}

	return strings.Split(tagValue, ",")[0]
}
