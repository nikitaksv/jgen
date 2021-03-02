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

package validation

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	pb "github.com/nikitaksv/gendata/proto"
)

func ValidateGenerateRequest(r *pb.GenerateRequest) error {
	return validation.ValidateStruct(r,
		validation.Field(
			&r.Template,
			validation.Required,
			validation.By(func(value interface{}) error {
				return ValidateTemplate(value.(*pb.Template))
			}),
		),
		validation.Field(
			&r.Schema,
			validation.Required,
			validation.By(func(value interface{}) error {
				return ValidateSchema(value.(*pb.Schema))
			}),
		),
	)
}

func ValidateSchema(schema *pb.Schema) error {
	return validation.ValidateStruct(schema,
		validation.Field(&schema.Content, validation.Required),
	)
}

func ValidateTemplate(tmpl *pb.Template) error {
	return validation.ValidateStruct(tmpl,
		validation.Field(&tmpl.Content, validation.Required),
	)
}
