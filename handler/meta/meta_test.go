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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMeta_UnmarshalFromJSON(t *testing.T) {
	jsonData := `{"a1":"one","a2":-10,"a3":"ahead","a4":false,"a5":17,"a6":"wet","a7":[{"b1":"one","c2":"two","c3":false},{"b1":"one","b4":4,"b5":["five"],"b6":true}],"a8":null}`
	want := `{"key":"","type":"","properties":[{"key":"a1","type":"string","nest":null},{"key":"a2","type":"int","nest":null},{"key":"a3","type":"string","nest":null},{"key":"a4","type":"bool","nest":null},{"key":"a5","type":"int","nest":null},{"key":"a6","type":"string","nest":null},{"key":"a7","type":"arrayObject","nest":{"key":"a7","type":"arrayObject","properties":[{"key":"b1","type":"string","nest":null},{"key":"c2","type":"string","nest":null},{"key":"c3","type":"bool","nest":null},{"key":"b4","type":"int","nest":null},{"key":"b5","type":"arrayString","nest":null},{"key":"b6","type":"bool","nest":null}]}},{"key":"a8","type":"null","nest":null}]}`

	m := &Meta{Type: &Type{
		origin: "",
		alias:  "",
	}}
	err := json.Unmarshal([]byte(jsonData), &m)
	if err != nil {
		t.Error(err)
	}

	metaJson, _ := json.Marshal(m)
	assert.JSONEq(t, want, string(metaJson))
}
