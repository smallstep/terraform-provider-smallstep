package utils

import (
	"encoding/json"
	"fmt"
	"io"

	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	apiserver "github.com/smallstep/terraform-provider-smallstep/internal/apiserver/v20230301"
)

// APIErrorMsg attempts to parse .Message from the API JSON response.
// Otherwise it returns the full response body.
func APIErrorMsg(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil {
		return "Failed to read Smallstep API response"
	}
	e := &v20230301.Error{}
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
func Deref[T dereferencable](v *T) (r T) {
	if v != nil {
		r = *v
	}
	return
}

// Describe parses descriptions for a component from its schema in Smallstep's
// OpenAPI spec. This ensures the terraform attribute documentation is kept in
// sync the the API spec.
func Describe(component string) (string, map[string]string, error) {
	spec, err := apiserver.GetSwagger()

	if err != nil {
		return "", nil, err
	}

	componentSchema, ok := spec.Components.Schemas[component]
	if !ok || componentSchema.Value == nil {
		return "", nil, fmt.Errorf("no schema found for %q in OpenAPI spec", component)
	}

	description := componentSchema.Value.Description

	propertyDescriptions := make(map[string]string, len(componentSchema.Value.Properties))

	for prop, schema := range componentSchema.Value.Properties {
		propertyDescriptions[prop] = schema.Value.Description
	}

	return description, propertyDescriptions, nil
}
