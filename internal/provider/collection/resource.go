package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20231101.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = collectionTypeName
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20231101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20231101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	httpResp, err := r.client.GetCollection(ctx, state.Slug.ValueString(), &v20231101.GetCollectionParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read collection %s: %v", state.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading collection %s: %s", reqID, httpResp.StatusCode, state.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.Collection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %s: %v", state.Slug.String(), err),
		)
		return
	}

	remote, d := fromAPI(ctx, collection, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read collection %q resource", collection.Slug))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("collection")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["displayName"],
				Optional:            true,
			},
			"schema_uri": schema.StringAttribute{
				MarkdownDescription: props["schemaURI"],
				Optional:            true,
			},
			"instance_count": schema.Int64Attribute{
				MarkdownDescription: props["instanceCount"],
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

	reqBody := toAPINew(&plan)

	b, _ := json.Marshal(reqBody)
	tflog.Trace(ctx, string(b))

	httpResp, err := a.client.PostCollections(ctx, &v20231101.PostCollectionsParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating collection %q: %s", reqID, httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.Collection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %q: %v", plan.Slug.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, collection, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("create collection %q resource", plan.Slug.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &Model{}

	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := toAPI(plan)

	httpResp, err := r.client.PutCollection(ctx, plan.Slug.ValueString(), &v20231101.PutCollectionParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to update collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating collection %q: %s", reqID, httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20231101.Collection{}
	if err := json.NewDecoder(httpResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %q: %v", plan.Slug.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, collection, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("update collection %q resource", plan.Slug.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Lookup the collection to ensure it's empty before deleting
	getResp, err := a.client.GetCollection(ctx, state.Slug.ValueString(), &v20231101.GetCollectionParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read collection %s: %v", state.Slug.String(), err),
		)
		return
	}
	defer getResp.Body.Close()
	if getResp.StatusCode == http.StatusNotFound {
		return
	}
	collection := &v20231101.Collection{}
	if err := json.NewDecoder(getResp.Body).Decode(collection); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection %s: %v", state.Slug.String(), err),
		)
		return
	}
	if collection.InstanceCount != 0 {
		resp.Diagnostics.AddError(
			"Delete Not Allowed",
			fmt.Sprintf("The collection cannot be deleted because it is not empty"),
		)
		return

	}

	httpResp, err := a.client.DeleteCollection(ctx, state.Slug.ValueString(), &v20231101.DeleteCollectionParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete collection %s: %v", state.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting collection %s: %s", reqID, httpResp.StatusCode, state.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("slug"), req, resp)
}
