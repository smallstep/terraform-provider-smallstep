package collection_instance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
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
	resp.TypeName = instanceTypeName
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20230301.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20230301.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.CollectionSlug.ValueString()
	id := state.ID.ValueString()

	httpResp, err := r.client.GetCollectionInstance(ctx, slug, id, &v20230301.GetCollectionInstanceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read collection instance %s/%s: %v", slug, id, err),
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
			fmt.Sprintf("Received status %d reading collection instance %s/%s: %s", httpResp.StatusCode, slug, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	instance := &v20230301.CollectionInstance{}
	if err := json.NewDecoder(httpResp.Body).Decode(instance); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection instance %s/%s: %v", slug, id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, slug, instance, req.State)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read collection resource %s/%s", slug, id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("collectionInstance")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	collectionWarning := `
The collection will be created implicitly if it does not exist.
If creating this collection with a smallstep_collection resource in the same config you MUST use depends_on to avoid race conditions.`

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"collection_slug": schema.StringAttribute{
				MarkdownDescription: props["collectionSlug"] + collectionWarning,
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"data": schema.StringAttribute{
				MarkdownDescription: props["data"],
				Required:            true,
			},
			"out_data": schema.StringAttribute{
				MarkdownDescription: "The instance data stored after any modifications made server-side. If the instance belongs to a device collection a host ID attribute will be added to the data.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: props["createdAt"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: props["updatedAt"],
				Computed:            true,
			},
		},
	}
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, diags := toAPI(&plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	b, _ := json.Marshal(reqBody)
	tflog.Trace(ctx, string(b))

	id := plan.ID.ValueString()
	slug := plan.CollectionSlug.ValueString()

	httpResp, err := a.client.PutCollectionInstance(ctx, slug, id, &v20230301.PutCollectionInstanceParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create collection instance %s/%s: %v", slug, id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	instance := &v20230301.CollectionInstance{}
	if err := json.NewDecoder(httpResp.Body).Decode(instance); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection instance %s/%s: %v", slug, id, err),
		)
		return
	}

	state, diags := fromAPI(ctx, slug, instance, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create collection instance %s/%s resource", slug, id))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &Model{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, diags := toAPI(plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()
	slug := plan.CollectionSlug.ValueString()

	httpResp, err := r.client.PutCollectionInstance(ctx, slug, id, &v20230301.PutCollectionInstanceParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to update collection instance %s/%s: %v", slug, id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	instance := &v20230301.CollectionInstance{}
	if err := json.NewDecoder(httpResp.Body).Decode(instance); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection instance %s/%s: %v", slug, id, err),
		)
		return
	}

	state, diags := fromAPI(ctx, slug, instance, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("update collection instance %s/%s resource", slug, id))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slug := state.CollectionSlug.ValueString()
	id := state.ID.ValueString()

	httpResp, err := a.client.DeleteCollectionInstance(ctx, slug, id, &v20230301.DeleteCollectionInstanceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete collection instance %s/%s: %v", slug, id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d deleting collection instance %s/%s: %s", httpResp.StatusCode, slug, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Import ID must be "<collection_slug>/<id>"`,
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("collection_slug"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
