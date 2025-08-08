package x509info

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/certfield"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

// This value is stored in the private state to indicate an optional, computed
// value in state is computed. An optional, computed object is one that will be
// set to a default by the server if not set by the client, e.g. if the x509
// object is not set in a request it will be set in the reply with a default
// common name and sans. When the terraform client sets the object and then
// later removes it in a subsequent request, the object will be unknown. We
// can use the value from state, but only if it's the computed default value
// set by the server, which is what this flag tracks. This has to be valid json
// even though it's never parsed.
var Computed = []byte(`{"computed": true}`)

const X509PrivateKey = "x509"

func NewResourceSchema() resourceschema.SingleNestedAttribute {
	certField := certfield.NewResourceSchema()
	certFieldList := certfield.NewListResourceSchema()

	return resourceschema.SingleNestedAttribute{
		MarkdownDescription: "", // TODO: add description
		Optional:            true,
		Computed:            true,
		PlanModifiers: []planmodifier.Object{
			utils.MaybeUseStateForUnknown(X509PrivateKey, Computed),
			utils.NullWhen(path.Root("credential").AtName("certificate_info").AtName("ssh"), basetypes.NewObjectNull(Attributes)),
		},
		Validators: []validator.Object{
			objectvalidator.ConflictsWith(
				path.MatchRoot("credential").AtName("certificate_info").AtName("ssh"),
			),
		},
		Attributes: map[string]resourceschema.Attribute{
			"common_name":         certField,
			"sans":                certFieldList,
			"organization":        certFieldList,
			"organizational_unit": certFieldList,
			"locality":            certFieldList,
			"country":             certFieldList,
			"province":            certFieldList,
			"street_address":      certFieldList,
			"postal_code":         certFieldList,
		},
	}
}

func NewDataSourceSchema() datasourceschema.SingleNestedAttribute {
	certField := certfield.NewDataSourceSchema()
	certFieldList := certfield.NewListDataSourceSchema()

	return datasourceschema.SingleNestedAttribute{
		MarkdownDescription: "", // TODO: add description
		Optional:            true,
		Computed:            true,
		Attributes: map[string]datasourceschema.Attribute{
			"common_name":         certField,
			"sans":                certFieldList,
			"organization":        certFieldList,
			"organizational_unit": certFieldList,
			"locality":            certFieldList,
			"country":             certFieldList,
			"province":            certFieldList,
			"street_address":      certFieldList,
			"postal_code":         certFieldList,
		},
	}
}
