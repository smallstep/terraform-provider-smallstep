package provider

import (
	"encoding/json"
	"fmt"
	"io"

	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	apiserver "github.com/smallstep/terraform-provider-smallstep/internal/apiserver/v20230301"
)

func apiErrorMsg(r io.Reader) string {
	body, err := io.ReadAll(r)
	if err != nil {
		return "Failed to read response body"
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

func deref[T dereferencable](v *T) (r T) {
	if v != nil {
		r = *v
	}
	return
}

func describe(component string) (string, map[string]string, error) {
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
