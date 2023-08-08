package authority

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20230301.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = authorityTypeName
}

func (r *Resource) x509IssuerSchema() (map[string]schema.Attribute, error) {
	_, properties, err := utils.Describe("x509Issuer")
	if err != nil {
		return nil, err
	}
	_, nameConstraints, err := utils.Describe("nameConstraints")
	if err != nil {
		return nil, err
	}
	_, subject, err := utils.Describe("distinguishedName")
	if err != nil {
		return nil, err
	}

	return map[string]schema.Attribute{
		"name": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: properties["name"],
		},
		"key_version": schema.StringAttribute{
			Required:            true,
			MarkdownDescription: properties["keyVersion"],
		},
		"duration": schema.StringAttribute{
			Optional:            true,
			MarkdownDescription: properties["duration"],
		},
		"max_path_length": schema.Int64Attribute{
			Optional:            true,
			MarkdownDescription: properties["maxPathLength"],
		},
		"name_constraints": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: properties["nameConstraints"],
			Attributes: map[string]schema.Attribute{
				"critical": schema.BoolAttribute{
					Optional:            true,
					MarkdownDescription: nameConstraints["critical"],
				},
				"permitted_dns_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["permittedDNSDomains"],
				},
				"excluded_dns_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["excludedDNSDomains"],
				},
				"permitted_ip_ranges": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["permittedIPRanges"],
				},
				"excluded_ip_ranges": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["excludedIPRanges"],
				},
				"permitted_email_addresses": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["permittedEmailAddresses"],
				},
				"excluded_email_addresses": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["excludedEmailAddresses"],
				},
				"permitted_uri_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["permittedURIDomains"],
				},
				"excluded_uri_domains": schema.SetAttribute{
					Optional:            true,
					ElementType:         types.StringType,
					MarkdownDescription: nameConstraints["excludedURIDomains"],
				},
			},
		},
		"subject": schema.SingleNestedAttribute{
			Optional:            true,
			MarkdownDescription: properties["subject"],
			Attributes: map[string]schema.Attribute{
				"common_name": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["commonName"],
				},
				"country": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["country"],
				},
				"organization": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["organization"],
				},
				"organizational_unit": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["organizationalUnit"],
				},
				"locality": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["locality"],
				},
				"province": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["province"],
				},
				"street_address": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["streetAddress"],
				},
				"email_address": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["emailAddress"],
				},
				"postal_code": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["postalCode"],
				},
				"serial_number": schema.StringAttribute{
					Optional:            true,
					MarkdownDescription: subject["serialNumber"],
				},
			},
		},
	}, nil
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, properties, err := utils.Describe("authority")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI authority schema",
			err.Error(),
		)
		return
	}
	x509Issuer, err := r.x509IssuerSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI x509-issuer schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: properties["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: properties["name"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: properties["type"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"subdomain": schema.StringAttribute{
				MarkdownDescription: properties["domain"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain": schema.StringAttribute{
				MarkdownDescription: properties["domain"],
				Computed:            true,
			},
			"fingerprint": schema.StringAttribute{
				MarkdownDescription: properties["fingerprint"],
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: properties["createdAt"],
				Computed:            true,
			},
			"active_revocation": schema.BoolAttribute{
				MarkdownDescription: properties["activeRevocation"],
				Optional:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
					boolplanmodifier.RequiresReplace(),
				},
			},
			"admin_emails": schema.SetAttribute{
				ElementType: types.StringType,
				Required:    true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"intermediate_issuer": schema.SingleNestedAttribute{
				MarkdownDescription: properties["intermediateIssuer"],
				Optional:            true,
				Attributes:          x509Issuer,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
			},
			"root_issuer": schema.SingleNestedAttribute{
				MarkdownDescription: properties["rootIssuer"],
				Optional:            true,
				Attributes:          x509Issuer,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
					objectplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20230301.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *v20230301.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var adminEmails []string
	resp.Diagnostics.Append(data.AdminEmails.ElementsAs(ctx, &adminEmails, false)...)

	intermediate, intermediateDiagnostics := data.IntermediateIssuer.AsAPI(ctx)
	resp.Diagnostics.Append(intermediateDiagnostics...)

	root, rootDiagnostics := data.RootIssuer.AsAPI(ctx)
	resp.Diagnostics.Append(rootDiagnostics...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := v20230301.PostAuthoritiesJSONRequestBody{
		Name:               data.Name.ValueString(),
		ActiveRevocation:   data.ActiveRevocation.ValueBoolPointer(),
		AdminEmails:        adminEmails,
		IntermediateIssuer: intermediate,
		RootIssuer:         root,
		Subdomain:          data.Subdomain.ValueString(),
		Type:               v20230301.NewAuthorityType(data.Type.ValueString()),
	}
	b, _ := json.Marshal(reqBody)
	tflog.Debug(ctx, string(b))
	httpResp, err := a.client.PostAuthorities(ctx, &v20230301.PostAuthoritiesParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create authority %q: %v", data.Name.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	authority := &v20230301.Authority{}
	if err := json.NewDecoder(httpResp.Body).Decode(authority); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal authority %q: %v", data.Name.ValueString(), err),
		)
		return
	}

	data.ID = types.StringValue(authority.Id)
	data.Domain = types.StringValue(authority.Domain)
	data.Fingerprint = types.StringValue(utils.Deref(authority.Fingerprint))
	data.CreatedAt = types.StringValue(authority.CreatedAt.Format(time.RFC3339))

	tflog.Trace(ctx, fmt.Sprintf("create authority %q resource", data.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (a *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.GetAuthority(ctx, data.ID.ValueString(), &v20230301.GetAuthorityParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read authority %q: %v", data.ID, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	authority := &v20230301.Authority{}
	if err := json.NewDecoder(httpResp.Body).Decode(authority); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal authority %q: %v", data.ID.ValueString(), err),
		)
		return
	}

	data.Name = types.StringValue(authority.Name)
	data.Type = types.StringValue(string(authority.Type))
	data.Domain = types.StringValue(authority.Domain)
	data.Fingerprint = types.StringValue(utils.Deref(authority.Fingerprint))
	data.CreatedAt = types.StringValue(authority.CreatedAt.Format(time.RFC3339))

	activeRevocation, diags := utils.ToOptionalBool(ctx, authority.ActiveRevocation, req.State, path.Root("active_revocation"))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ActiveRevocation = activeRevocation

	adminEmails, diags := utils.ToOptionalSet(ctx, authority.AdminEmails, req.State, path.Root("admin_emails"))
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.AdminEmails = adminEmails

	// Subdomain will be missing if this was an import but is required
	if data.Subdomain.IsNull() {
		parts := strings.Split(data.Domain.ValueString(), ".")
		data.Subdomain = types.StringValue(parts[0])
	}

	tflog.Trace(ctx, fmt.Sprintf("read authority %q resource", data.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update not supported. All changes require replacement.
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.DeleteAuthority(ctx, data.ID.ValueString(), &v20230301.DeleteAuthorityParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete authority %s: %v", data.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d deleting authority %s: %s", httpResp.StatusCode, data.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
