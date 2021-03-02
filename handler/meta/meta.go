/*
 * Copyright (c) 2021 Nikita Krasnikov
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"

	"github.com/nikitaksv/dynjson"
	"github.com/nikitaksv/strcase"
)

const (
	TypeNull        = "null"
	TypeInt         = "int"
	TypeString      = "string"
	TypeBool        = "bool"
	TypeFloat       = "float"
	TypeObject      = "object"
	TypeArray       = "array"
	TypeArrayObject = "arrayObject"
	TypeArrayInt    = "arrayInt"
	TypeArrayString = "arrayString"
	TypeArrayBool   = "arrayBool"
	TypeArrayFloat  = "arrayFloat"
)

type TypeAliases map[string]string

func (a TypeAliases) Apply(t string) string {
	if nT, ok := a[t]; ok && nT != "" {
		return nT
	}
	return t
}
func (a TypeAliases) Add(k, v string) TypeAliases {
	a[k] = v
	return a
}

type Meta struct {
	index           int
	PrefixObjectKey string
	SortProps       bool
	Key             Key         `json:"key"`
	Type            *Type       `json:"type"`
	TypeAliases     TypeAliases `json:"-"`
	Properties      []*Property `json:"properties"`
}

func (m *Meta) UnmarshalJSON(data []byte) error {
	j := &dynjson.Json{}
	err := j.UnmarshalJSON(data)
	if err != nil {
		return err
	}
	if jObj, ok := j.Value.(*dynjson.Object); ok {
		parseMap(m, jObj)
	} else if jArr, ok := j.Value.(*dynjson.Array); ok {
		mergedArray := mergeArray(jArr)
		if len(mergedArray.Elements) > 0 {
			if valObj, ok := mergedArray.Elements[0].(*dynjson.Object); ok {
				parseMap(m, valObj)
			}
		}
	}
	return nil
}

func (m *Meta) getProperty(key Key) (*Property, bool) {
	for _, p := range m.Properties {
		if p.Key == key {
			return p, true
		}
	}
	return nil, false
}

func (m *Meta) Sort() {
	if m.SortProps {
		sortProperties(m.Properties)
	}
}

func (m *Meta) SortKeys() {
	if m.SortProps {
		sortPropertiesByKeys(m.Properties)
	}
}

func parseMap(obj *Meta, aMap *dynjson.Object) {
	for k, v := range aMap.Properties {
		prop := &Property{
			index: k,
			Key:   Key(v.Key),
			Type:  TypeOf(v.Value, obj.TypeAliases),
			Nest:  nil,
		}

		switch v.Value.(type) {
		case *dynjson.Object:
			nestObj := v.Value.(*dynjson.Object)
			newObj := &Meta{
				PrefixObjectKey: obj.PrefixObjectKey,
				SortProps:       obj.SortProps,
				Key:             Key(obj.PrefixObjectKey + prop.Key.String()),
				Type:            TypeOf(v.Value, obj.TypeAliases),
				TypeAliases:     obj.TypeAliases,
				Properties:      make([]*Property, 0, len(nestObj.Properties)),
			}
			parseMap(newObj, nestObj)
			prop.Nest = newObj
		case *dynjson.Array:
			nestedObj := &Meta{
				PrefixObjectKey: obj.PrefixObjectKey,
				SortProps:       obj.SortProps,
				Key:             prop.Key,
				Type:            TypeOf(v.Value, obj.TypeAliases),
				TypeAliases:     obj.TypeAliases,
				Properties:      nil,
			}
			mergedArray := mergeArray(v.Value.(*dynjson.Array))
			if len(mergedArray.Elements) > 0 {
				if valObj, ok := mergedArray.Elements[0].(*dynjson.Object); ok {
					parseMap(nestedObj, valObj)
				}
			}
			prop.Nest = nestedObj
		}

		obj.Properties = append(obj.Properties, prop)
		obj.SortKeys()
	}
}

// Merge Array, например ["string", {"a1": 1}] и [false, 20, {"b2": 2}] объединятся в ["string",false,20,{"a1": 1, "b2": 2}]
func mergeArray(arr *dynjson.Array) *dynjson.Array {
	res := &dynjson.Array{Elements: make([]interface{}, 0, len(arr.Elements))}

	m := &dynjson.Object{
		Key:        "",
		Properties: nil,
	}

	for _, v := range arr.Elements {
		switch v.(type) {
		case *dynjson.Object:
			m = mergeObjects(m, v.(*dynjson.Object))
		case *dynjson.Array:
			mergedArray := mergeArray(v.(*dynjson.Array))
			if len(mergedArray.Elements) > 0 {
				if mergedObj, ok := mergedArray.Elements[0].(*dynjson.Object); ok {
					m = mergeObjects(m, mergedObj)
					continue
				}
			}
			res.Elements = append(res.Elements, mergedArray.Elements...)
		default:
			res.Elements = append(res.Elements, v)
		}
	}

	if len(m.Properties) > 0 {
		res.Elements = append(res.Elements, m)
	}

	return res
}

func mergeObjects(maps ...*dynjson.Object) *dynjson.Object {
	result := &dynjson.Object{
		Properties: []*dynjson.Property{},
	}
	for _, m := range maps {
		for _, v := range m.Properties {
			existsV, exists := result.GetProperty(v.Key)
			switch v.Value.(type) {
			case *dynjson.Array:
				mergedArray := mergeArray(v.Value.(*dynjson.Array))
				if len(mergedArray.Elements) > 0 {
					if mergedObj, ok := mergedArray.Elements[0].(*dynjson.Object); ok {
						result.Properties = append(result.Properties, mergeObjects(v.Value.(*dynjson.Object), mergedObj).Properties...)
						continue
					}
				}
				v.Value = mergedArray
				result.Properties = append(result.Properties, v)
			case *dynjson.Object:
				if exists {
					result.Properties = append(result.Properties, mergeObjects(existsV.Value.(*dynjson.Object), v.Value.(*dynjson.Object)).Properties...)
				} else {
					result.Properties = append(result.Properties, v)
				}
			default:
				if !exists {
					result.Properties = append(result.Properties, v)
				}
			}
		}
	}
	return result
}

type Key string

func (k Key) String() string {
	return string(k)
}

// CamelCase ex. camelCase
func (k Key) CamelCase() Key {
	return Key(strcase.ToCamelCase(k.String()))
}

// PascalCase ex. PascalCase
func (k Key) PascalCase() Key {
	return Key(strcase.ToPascalCase(k.String()))
}

// SnakeCase ex. snake_case
func (k Key) SnakeCase() Key {
	return Key(strcase.ToSnakeCase(k.String()))
}

// KebabCase ex. kebab-case
func (k Key) KebabCase() Key {
	return Key(strcase.ToKebabCase(k.String()))
}

// DotCase ex. dot.case
func (k Key) DotCase() Key {
	return Key(strcase.ToDotCase(k.String()))
}

type Type struct {
	origin string
	alias  string
}

func (t Type) MarshalJSON() ([]byte, error) {
	bs, err := json.Marshal(t.String())
	return bs, err
}

func NewType(t string, a string) *Type {
	return &Type{
		origin: t,
		alias:  a,
	}
}

func (t Type) String() string {
	if t.alias != "" {
		return t.alias
	}
	return t.origin
}

func (t Type) Long() Type {
	return t
}
func (t Type) Short() Type {
	return t
}
func (t Type) IsNull() bool {
	return t.origin == TypeNull
}
func (t Type) IsInt() bool {
	return t.origin == TypeInt
}
func (t Type) IsBool() bool {
	return t.origin == TypeBool
}
func (t Type) IsFloat() bool {
	return t.origin == TypeFloat
}
func (t Type) IsNumber() bool {
	return t.IsFloat() || t.IsInt()
}
func (t Type) IsString() bool {
	return t.origin == TypeString
}
func (t Type) IsArray() bool {
	return t.origin == TypeArray ||
		t.origin == TypeArrayBool ||
		t.origin == TypeArrayFloat ||
		t.origin == TypeArrayObject ||
		t.origin == TypeArrayString
}
func (t Type) IsArrayObject() bool {
	return t.origin == TypeArrayObject
}
func (t Type) IsArrayBool() bool {
	return t.origin == TypeArrayBool
}
func (t Type) IsArrayInt() bool {
	return t.origin == TypeArrayInt
}
func (t Type) IsArrayFloat() bool {
	return t.origin == TypeArrayFloat
}
func (t Type) IsArrayString() bool {
	return t.origin == TypeArrayString
}
func (t Type) IsObject() bool {
	return t.origin == TypeObject
}

// Returning meta-type data
func TypeOf(v interface{}, aliases TypeAliases) *Type {
	switch v.(type) {
	case []interface{}:
		nT := typeOfArray(v.([]interface{}))
		return NewType(nT, aliases.Apply(nT))
	case *dynjson.Array:
		nT := typeOfArray(v.(*dynjson.Array).Elements)
		return NewType(nT, aliases.Apply(nT))
	case map[string]interface{}, *dynjson.Object:
		return NewType(TypeObject, aliases.Apply(TypeObject))
	case bool:
		return NewType(TypeBool, aliases.Apply(TypeBool))
	case float32, float64:
		vFloat64 := v.(float64)
		if vFloat64 == math.Trunc(vFloat64) {
			return NewType(TypeInt, aliases.Apply(TypeInt))
		}

		return NewType(TypeFloat, aliases.Apply(TypeFloat))
	case int, int8, int16, int32, int64:
		return NewType(TypeInt, aliases.Apply(TypeInt))
	case string:
		return NewType(TypeString, aliases.Apply(TypeString))
	default:
		return NewType(TypeNull, aliases.Apply(TypeNull))
	}
}

// If json/xml array have mixed type data. This function detect most superior data type.
func typeOfArray(arr []interface{}) string {
	mx := map[string]int{
		TypeArrayBool:   0,
		TypeArrayFloat:  0,
		TypeArrayInt:    0,
		TypeArrayString: 0,
		TypeArrayObject: 0,
		TypeArray:       0,
	}

	for _, v := range arr {
		switch v.(type) {
		case map[string]interface{}, *dynjson.Object:
			mx[TypeArrayObject]++
		case []interface{}:
			mx[typeOfArray(v.([]interface{}))]++
		case *dynjson.Array:
			mx[typeOfArray(v.(*dynjson.Array).Elements)]++
		case int, int8, int16, int32, int64:
			mx[TypeArrayInt]++
		case float32, float64:
			vFloat64 := v.(float64)
			if vFloat64 == math.Trunc(vFloat64) {
				mx[TypeArrayInt]++
			} else {
				mx[TypeArrayInt] = 0
				mx[TypeArrayFloat]++
			}
		case bool:
			mx[TypeArrayInt] = 0
			mx[TypeArrayFloat] = 0
			mx[TypeArrayBool]++
		case string:
			vS := v.(string)
			if vFloat64, err := strconv.ParseFloat(vS, 64); err == nil {
				if vFloat64 == math.Trunc(vFloat64) {
					mx[TypeArrayInt]++
				} else {
					mx[TypeArrayInt] = 0
					mx[TypeArrayFloat]++
				}
			} else if _, err := strconv.ParseBool(vS); err == nil {
				mx[TypeArrayInt] = 0
				mx[TypeArrayFloat] = 0
				mx[TypeArrayBool]++
			} else {
				if mx[TypeArrayInt] > 0 || mx[TypeArrayFloat] > 0 || mx[TypeArrayBool] > 0 || mx[TypeArrayObject] > 0 {
					// Then array have a mixed types
					mx[TypeArray]++
				} else {
					mx[TypeArrayString]++
				}
			}
		default:
			mx[TypeArray]++
		}
	}

	if mx[TypeArray] > 0 {
		return TypeArray
	}

	max := 0
	for k, v := range mx {
		if v > max {
			max = v
			return k
		}
	}

	return ""
}

type Property struct {
	// for origin order sorting
	index int

	Key  Key   `json:"key"`
	Type *Type `json:"type"`
	Nest *Meta `json:"nest"`
}

func sortProperties(props []*Property) {
	sort.Slice(props, func(i, j int) bool { return props[i].index < props[j].index })
}

func sortPropertiesByKeys(props []*Property) {
	sort.Slice(props, func(i, j int) bool { return props[i].Key < props[j].Key })
}

func Split(m Meta) []Meta {
	var objects []Meta
	objects = append(objects, m)
	for _, prop := range m.Properties {
		if prop.Type.IsObject() || prop.Type.IsArrayObject() {
			if prop.Nest != nil {
				objects = append(objects, Split(*prop.Nest)...)
			}
		}
	}
	return objects
}
