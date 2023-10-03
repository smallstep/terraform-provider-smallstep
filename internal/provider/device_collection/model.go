package device_collection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const collectionTypeName = "smallstep_device_collection"

type Model struct {
	Slug        types.String `tfsdk:"slug"`
	DisplayName types.String `tfsdk:"display_name"`
	AdminEmails types.Set    `tfsdk:"admin_emails"`
	DeviceType  types.String `tfsdk:"device_type"`
	AWSDevice   *AWSDevice   `tfsdk:"aws_vm"`
	GCPDevice   *GCPDevice   `tfsdk:"gcp_vm"`
	AzureDevice *AzureDevice `tfsdk:"azure_vm"`
	TPMDevice   *TPMDevice   `tfsdk:"tpm"`
	// https://developer.hashicorp.com/terraform/plugin/framework/acctests#implement-id-attribute
	ID types.String `tfsdk:"id"`
}

type AWSDevice struct {
	Accounts          types.Set  `tfsdk:"accounts"`
	DisableCustomSANs types.Bool `tfsdk:"disable_custom_sans"`
}

type GCPDevice struct {
	ServiceAccounts   types.Set  `tfsdk:"service_accounts"`
	ProjectIDs        types.Set  `tfsdk:"project_ids"`
	DisableCustomSANs types.Bool `tfsdk:"disable_custom_sans"`
}

type AzureDevice struct {
}

type TPMDevice struct {
}

func fromAPI(ctx context.Context, collection *v20230301.DeviceCollection, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		ID:          types.StringValue(collection.Slug),
		Slug:        types.StringValue(collection.Slug),
		DisplayName: types.StringValue(collection.DisplayName),
		DeviceType:  types.StringValue(string(collection.DeviceType)),
	}

	switch collection.DeviceType {
	case v20230301.DeviceCollectionDeviceTypeAwsVm:
		aws, err := collection.DeviceTypeConfiguration.AsAwsVM()
		if err != nil {
			diags.AddError("Read AWS Device Configuration", err.Error())
			return nil, diags
		}

		disableCustomSANs, d := utils.ToOptionalBool(ctx, aws.DisableCustomSANs, state, path.Root("aws_vm").AtName("disable_custom_sans"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var accounts []attr.Value
		for _, account := range aws.Accounts {
			accounts = append(accounts, types.StringValue(account))
		}
		accountsSet, diags := types.SetValue(types.StringType, accounts)
		if diags.HasError() {
			return nil, diags
		}

		model.AWSDevice = &AWSDevice{
			Accounts:          accountsSet,
			DisableCustomSANs: disableCustomSANs,
		}
	case v20230301.DeviceCollectionDeviceTypeGcpVm:
		gcp, err := collection.DeviceTypeConfiguration.AsGcpVM()
		if err != nil {
			diags.AddError("Read GCP Device Configuration", err.Error())
			return nil, diags
		}

		disableCustomSANs, d := utils.ToOptionalBool(ctx, gcp.DisableCustomSANs, state, path.Root("gcp_vm").AtName("disable_custom_sans"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		serviceAccounts, d := utils.ToOptionalSet(ctx, gcp.ServiceAccounts, state, path.Root("gcp_vm").AtName("service_accounts"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		projectIDs, d := utils.ToOptionalSet(ctx, gcp.ProjectIDs, state, path.Root("gcp_vm").AtName("project_ids"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		model.GCPDevice = &GCPDevice{
			DisableCustomSANs: disableCustomSANs,
			ServiceAccounts:   serviceAccounts,
			ProjectIDs:        projectIDs,
		}
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.DeviceCollection, diag.Diagnostics) {
	var diags diag.Diagnostics

	var adminEmails []string
	diags.Append(model.AdminEmails.ElementsAs(ctx, &adminEmails, false)...)

	dc := &v20230301.DeviceCollection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueString(),
		AdminEmails: &adminEmails,
		DeviceType:  v20230301.DeviceCollectionDeviceType(model.DeviceType.ValueString()),
	}

	switch dc.DeviceType {
	case v20230301.DeviceCollectionDeviceTypeAwsVm:
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
	case v20230301.DeviceCollectionDeviceTypeGcpVm:
		if model.GCPDevice == nil {
			diags.AddError("GCP Device", "gcp_vm is required with device type gcp-vm")
			return nil, diags
		}
		gcp := v20230301.GcpVM{
			DisableCustomSANs: model.GCPDevice.DisableCustomSANs.ValueBoolPointer(),
			ProjectIDs:        nil,
			ServiceAccounts:   nil,
		}
		if !model.GCPDevice.ServiceAccounts.IsNull() {
			d := model.GCPDevice.ServiceAccounts.ElementsAs(ctx, &gcp.ServiceAccounts, false)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
		}
		if !model.GCPDevice.ProjectIDs.IsNull() {
			d := model.GCPDevice.ProjectIDs.ElementsAs(ctx, &gcp.ProjectIDs, false)
			diags.Append(d...)
			if diags.HasError() {
				return nil, diags
			}
		}
		if err := dc.DeviceTypeConfiguration.FromGcpVM(gcp); err != nil {
			diags.AddError("GCP VM", err.Error())
			return nil, diags
		}
	default:
		diags.AddError("Device Type", fmt.Sprintf("Unsupported device collection type %q", dc.DeviceType))
		return nil, diags
	}

	return dc, diags
}
