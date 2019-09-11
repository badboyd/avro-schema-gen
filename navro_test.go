// Copyright © 2018 Kelvin Vuong <kelvin@chotot.vn>
//
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

package navro

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type anotherType struct {
	list     []string
	bits     []byte
	fix      [3]int
	arraymap []map[string]int
}

type someMore struct {
	name string
}

type someType struct {
	num     int
	amap    map[string]int
	another anotherType
	someMore
	arraystr []someMore
}

type noInterface struct {
	in interface{}
}

type noPointer struct {
	in *int
}

type Pointer struct {
	in *int
}

type StructPointer struct {
	P   []*Pointer
	PND *Pointer
}

var r = strings.NewReplacer(
	"\n", "",
	"\t", "",
	" ", "",
)

var testCasesV2 = map[string]struct {
	in                  interface{}
	support             bool
	expectedRecordNames map[string]bool
	expected            string
}{
	"nointerface": {
		in:                  noInterface{},
		support:             false,
		expected:            "",
		expectedRecordNames: nil,
	},
	"Pointer": {
		in:      Pointer{},
		support: true,
		expected: `{
			"type":"record",
			"name":"Pointer",
			"fields":[
				{"name":"in",
				"default":null,
				"type":[
					"null",
					"long"
					]
				}
			]
		}`,
		expectedRecordNames: map[string]bool{"Pointer": true},
	},
	"StructPointer": {
		in:      StructPointer{},
		support: true,
		expected: `{
		  "type": "record",
		  "name": "StructPointer",
		  "fields": [
		    {
		      "name": "P",
		      "type": {
		        "type": "array",
		        "items": [
		          "null",
		          {
		            "type": "record",
		            "name": "Pointer",
		            "fields": [
		              {
		                "name": "in",
		                "default": null,
		                "type": [
		                  "null",
		                  "long"
		                ]
		              }
		            ]
		          }
		        ]
		      }
		    },
		    {
		      "name": "PND",
		      "default": null,
		      "type": [
		        "null",
		        "Pointer"
		      ]
		    }
		  ]
		}`,
		expectedRecordNames: map[string]bool{
			"Pointer":       true,
			"StructPointer": true,
		},
	},
	"string": {
		in:                  "some string",
		support:             true,
		expected:            `{"type": "string"}`,
		expectedRecordNames: map[string]bool{},
	},
	"int": {
		in:                  10,
		support:             true,
		expected:            `{"type": "long"}`,
		expectedRecordNames: map[string]bool{},
	},
	"double": {
		in:                  0.48,
		support:             true,
		expected:            `{"type": "double"}`,
		expectedRecordNames: map[string]bool{},
	},
	"array": {
		in:                  []int{1, 2, 3},
		support:             true,
		expected:            `{"type": "array", "items": "long"}`,
		expectedRecordNames: map[string]bool{},
	},
	"struct": {
		in:      anotherType{[]string{"test", "t"}, nil, [3]int{1, 2, 3}, []map[string]int{}},
		support: true,
		expected: `{
			"type": "record",
			"name": "anotherType",
			"fields":
			[
				{
					"name": "list",
					"type": {"type": "array", "items": "string"}
				},
				{
					"name": "bits",
					"type": "bytes"
				},
				{
					"name": "fix",
					"type": {"type": "array", "items": "long"}
				},
				{
					"name": "arraymap",
					"type": {"type": "array", "items": {
						"type": "map",
						"values": "long"
						}
					}
				}
			]
		}`,
		expectedRecordNames: map[string]bool{"anotherType": true},
	},
	"nested": {
		in:      someType{},
		support: true,
		expected: `{
	"type" : "record",
	"name" : "someType",
	"fields" : [
	{
		"name" : "num",
		"type" : "long"
	},
	{
		"name" : "amap",
		"type" : {
			"type": "map",
			"values": "long"
		}
	},
	{
		"name": "another",
		"type":
		{
			"type": "record",
			"name": "anotherType",
			"fields":
			[
				{
					"name": "list",
					"type": {"type": "array", "items": "string"}
				},
				{
					"name": "bits",
					"type": "bytes"
				},
				{
					"name": "fix",
					"type": {"type": "array", "items": "long"}
				},
				{
					"name": "arraymap",
					"type": {"type": "array", "items": {
						"type": "map",
						"values": "long"
						}
					}
				}
			]
		}
	},
	{
		"name": "someMore",
		"type": {
			"type": "record",
			"name": "someMore",
			"fields": [
			{
				"name": "name",
				"type": "string"
			}
			]
		}
	},
	{
		"name": "arraystr",
		"type": {"type": "array", "items": "someMore"
		}
	}
	]
}`,
		expectedRecordNames: map[string]bool{
			"anotherType": true,
			"someMore":    true,
			"someType":    true,
		},
	},
}

func TestGenerate(t *testing.T) {
	for i, test := range testCasesV2 {
		fmt.Printf("-- Test %s --\n", i)
		s, mT, err := Generate(test.in)
		if err != nil {
			_, ok := err.(NotSupported)
			if !ok {
				fmt.Print(err)
				t.Fail()
				continue
			}
			if ok && test.support {
				fmt.Print("Expect support")
				t.Fail()
				continue
			}
		}
		if err == nil {
			if strings.Compare(r.Replace(s), r.Replace(test.expected)) != 0 {
				fmt.Println("got: \n ", r.Replace(s))
				fmt.Println("expected: \n ", r.Replace(test.expected))
				t.Fail()
				continue
			}
			if !reflect.DeepEqual(mT, test.expectedRecordNames) {
				fmt.Println("got: \n ", mT)
				fmt.Println("expected: \n ", test.expectedRecordNames)
				t.Fail()
				continue
			}
		}
		fmt.Printf("--> passed\n")
	}
}
