package device_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

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

var _ resource.Resource = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20230301.Client
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

	httpResp, err := r.client.GetCollection(ctx, state.Slug.ValueString(), &v20230301.GetCollectionParams{})
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
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d reading collection %s: %s", httpResp.StatusCode, state.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20230301.Collection{}
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), remote.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_count"), remote.InstanceCount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), remote.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), remote.CreatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("updated_at"), remote.UpdatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), remote.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_uri"), remote.SchemaURI)...)
	// Response is a collection. Use plan for everything else.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_type"), state.DeviceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws_vm"), state.AWSDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), state.AdminEmails)...)
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	component, props, err := utils.Describe("newDeviceCollection")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	aws, awsProps, err := utils.Describe("awsVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	gcp, _, err := utils.Describe("gcpVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	azure, _, err := utils.Describe("azureVM")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	tpm, _, err := utils.Describe("tpm")
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
			"id": schema.StringAttribute{
				MarkdownDescription: "Internal use only",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: props["slug"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"schema_uri": schema.StringAttribute{
				MarkdownDescription: props["schemaURI"],
				Computed:            true,
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
			"admin_emails": schema.SetAttribute{
				MarkdownDescription: props["adminEmails"],
				ElementType:         types.StringType,
				Required:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
					setplanmodifier.RequiresReplace(),
				},
			},
			"device_type": schema.StringAttribute{
				MarkdownDescription: props["deviceType"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aws_vm": schema.SingleNestedAttribute{
				MarkdownDescription: aws,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"accounts": schema.SetAttribute{
						MarkdownDescription: awsProps["accounts"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableCustomSANs"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"azure_vm": schema.SingleNestedAttribute{
				MarkdownDescription: azure,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{},
			},
			"gcp_vm": schema.SingleNestedAttribute{
				MarkdownDescription: gcp,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{},
			},
			"tpm": schema.SingleNestedAttribute{
				MarkdownDescription: tpm,
				Optional:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
				Attributes: map[string]schema.Attribute{},
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

	reqBody, d := toAPI(ctx, &plan)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PostDeviceCollections(ctx, &v20230301.PostDeviceCollectionsParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create device collection %q: %v", plan.Slug.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d creating device collection %q: %s", httpResp.StatusCode, plan.Slug.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	collection := &v20230301.Collection{}
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

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("slug"), state.Slug)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("instance_count"), state.InstanceCount)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("display_name"), state.DisplayName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("created_at"), state.CreatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("updated_at"), state.UpdatedAt)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), state.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("schema_uri"), state.SchemaURI)...)
	// Response is a collection. Use plan for everything else.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("device_type"), plan.DeviceType)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aws_vm"), plan.AWSDevice)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("admin_emails"), plan.AdminEmails)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Device collections must be destroyed and re-created.")
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Not implemented but cannot report an error because this is called in acceptance tests
	resp.Diagnostics.AddWarning("Delete not supported", "Coming soon.")
}
