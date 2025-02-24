package utils

import (
	"encoding/json"
	"fmt"
	"io"

	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
)

// APIErrorMsg attempts to parse .Message from the API JSON response.
// Otherwise it returns the full response body.
func APIErrorMsg(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil {
		return "Failed to read Smallstep API response"
	}
	e := &v20250101.Error{}
	if err := json.Unmarshal(body, e); err != nil {
		return string(body)
	}
	return e.Message
}

type dereferencable interface {
	string |
		bool |
		int |
		[]string
}

// Deref gets the default value for a pointer type. This makes it easier to work
// with the generated API client code, which uses pointers for optional fields.
func Deref[T any](v *T) (r T) {
	if v != nil {
		r = *v
	}
	return
}

func Ref[T any](v T) *T {
	return &v
}

func ToIntPointer(i64 *int64) *int {
	if i64 == nil {
		return nil
	}
	i := int(*i64)
	return &i
}

func ToStringPointer[Out ~string](str *string) *Out {
	if str == nil {
		return nil
	}
	s := Out(*str)
	return &s
}

// Describe parses descriptions for a component from its schema in Smallstep's
// OpenAPI spec. This ensures the terraform attribute documentation is kept in
// sync with the API spec.
func Describe(component string) (string, map[string]string, error) {
	spec, err := v20250101.GetSwagger()
	if err != nil {
		return "", nil, err
	}

	componentSchema, ok := spec.Components.Schemas[component]
	if !ok || componentSchema.Value == nil {
		return "", nil, fmt.Errorf("no schema found for %q in OpenAPI spec", component)
	}

	description := componentSchema.Value.Description

	props := componentSchema.Value.Properties
	// provisioner schema uses AllOf
	for _, s := range componentSchema.Value.AllOf {
		if s == nil || s.Value == nil {
			continue
		}
		if props == nil {
			props = s.Value.Properties
			continue
		}
		for k, p := range s.Value.Properties {
			props[k] = p
		}
	}

	propertyDescriptions := make(map[string]string, len(props))

	for prop, schema := range props {
		d := schema.Value.Description
		if len(schema.Value.Enum) > 0 {
			d += " Allowed values:"
			for _, enum := range schema.Value.Enum {
				d += fmt.Sprintf(" `%s`", enum)
			}
		} else if schema.Value.Items != nil && schema.Value.Items.Value != nil && len(schema.Value.Items.Value.Enum) > 0 {
			d += " Allowed values:"
			for _, enum := range schema.Value.Items.Value.Enum {
				d += fmt.Sprintf(" `%s`", enum)
			}
		}
		propertyDescriptions[prop] = d
	}

	return description, propertyDescriptions, nil
}
