package device_collection

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const collectionTypeName = "smallstep_device_collection"

type Model struct {
	Slug          types.String `tfsdk:"slug"`
	DisplayName   types.String `tfsdk:"display_name"`
	InstanceCount types.Int64  `tfsdk:"instance_count"`
	CreatedAt     types.String `tfsdk:"created_at"`
	UpdatedAt     types.String `tfsdk:"updated_at"`
	SchemaURI     types.String `tfsdk:"schema_uri"`
	AdminEmails   types.Set    `tfsdk:"admin_emails"`
	DeviceType    types.String `tfsdk:"device_type"`
	AWSDevice     *AWSDevice   `tfsdk:"aws_vm"`
	GCPDevice     *GCPDevice   `tfsdk:"gcp_vm"`
	AzureDevice   *AzureDevice `tfsdk:"azure_vm"`
	TPMDevice     *TPMDevice   `tfsdk:"tpm"`
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	ID types.String `tfsdk:"id"`
}

type AWSDevice struct {
	Accounts          types.Set  `tfsdk:"accounts"`
	DisableCustomSANs types.Bool `tfsdk:"disable_custom_sans"`
}

type GCPDevice struct {
}

type AzureDevice struct {
}

type TPMDevice struct {
}

func fromAPI(ctx context.Context, collection *v20230301.Collection, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		Slug:          types.StringValue(collection.Slug),
		InstanceCount: types.Int64Value(int64(collection.InstanceCount)),
		DisplayName:   types.StringValue(collection.DisplayName),
		CreatedAt:     types.StringValue(collection.CreatedAt.Format(time.RFC3339)),
		UpdatedAt:     types.StringValue(collection.UpdatedAt.Format(time.RFC3339)),
		ID:            types.StringValue(collection.Slug),
		SchemaURI:     types.StringPointerValue(collection.SchemaURI),
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.NewDeviceCollection, diag.Diagnostics) {
	var diags diag.Diagnostics

	var adminEmails []string
	diags.Append(model.AdminEmails.ElementsAs(ctx, &adminEmails, false)...)

	dc := &v20230301.NewDeviceCollection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueString(),
		AdminEmails: &adminEmails,
		DeviceType:  v20230301.NewDeviceCollectionDeviceType(model.DeviceType.ValueString()),
	}

	switch dc.DeviceType {
	case v20230301.NewDeviceCollectionDeviceTypeAwsVm:
		if model.AWSDevice == nil {
			diags.AddError("AWS Device", "aws_vm is required with device type aws-vm")
			return nil, diags
		}
		aws := v20230301.AwsVM{
			Accounts:          []string{},
			DisableCustomSANs: model.AWSDevice.DisableCustomSANs.ValueBoolPointer(),
		}
		d := model.AWSDevice.Accounts.ElementsAs(ctx, &aws.Accounts, false)
		diags.Append(d...)

		if err := dc.DeviceTypeConfiguration.FromAwsVM(aws); err != nil {
			diags.AddError("AWS VM", err.Error())
			return nil, diags
		}
	default:
		diags.AddError("Device Type", fmt.Sprintf("Unsupported device collection type %q", dc.DeviceType))
		return nil, diags
	}

	return dc, diags
}
