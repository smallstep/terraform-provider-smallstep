package device_collection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const collectionTypeName = "smallstep_device_collection"

type Model struct {
	Slug        types.String `tfsdk:"slug"`
	AuthorityID types.String `tfsdk:"authority_id"`
	DisplayName types.String `tfsdk:"display_name"`
	DeviceType  types.String `tfsdk:"device_type"`
	AWSDevice   *AWSDevice   `tfsdk:"aws_vm"`
	GCPDevice   *GCPDevice   `tfsdk:"gcp_vm"`
	AzureDevice *AzureDevice `tfsdk:"azure_vm"`
	TPMDevice   *TPMDevice   `tfsdk:"tpm"`
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
	TenantID          types.String `tfsdk:"tenant_id"`
	ResourceGroups    types.Set    `tfsdk:"resource_groups"`
	DisableCustomSANs types.Bool   `tfsdk:"disable_custom_sans"`
	Audience          types.String `tfsdk:"audience"`
}

type TPMDevice struct {
	AttestorIntermediates types.String `tfsdk:"attestor_intermediates"`
	AttestorRoots         types.String `tfsdk:"attestor_roots"`
	ForceCN               types.Bool   `tfsdk:"force_cn"`
	RequireEAB            types.Bool   `tfsdk:"require_eab"`
}

func fromAPI(ctx context.Context, collection *v20231101.DeviceCollection, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		Slug:        types.StringValue(collection.Slug),
		DisplayName: types.StringValue(collection.DisplayName),
		DeviceType:  types.StringValue(string(collection.DeviceType)),
		AuthorityID: types.StringValue(collection.AuthorityID),
	}

	switch collection.DeviceType {
	case v20231101.DeviceCollectionDeviceTypeAwsVm:
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
	case v20231101.DeviceCollectionDeviceTypeAzureVm:
		azure, err := collection.DeviceTypeConfiguration.AsAzureVM()
		if err != nil {
			diags.AddError("Read Azure Device Configuration", err.Error())
			return nil, diags
		}

		disableCustomSANs, d := utils.ToOptionalBool(ctx, azure.DisableCustomSANs, state, path.Root("azure_vm").AtName("disable_custom_sans"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		var resourceGroups []attr.Value
		for _, rg := range azure.ResourceGroups {
			resourceGroups = append(resourceGroups, types.StringValue(rg))
		}
		resourceGroupsSet, diags := types.SetValue(types.StringType, resourceGroups)
		if diags.HasError() {
			return nil, diags
		}

		audience, d := utils.ToOptionalString(ctx, azure.Audience, state, path.Root("azure_vm").AtName("audience"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		model.AzureDevice = &AzureDevice{
			TenantID:          types.StringValue(azure.TenantID),
			ResourceGroups:    resourceGroupsSet,
			DisableCustomSANs: disableCustomSANs,
			Audience:          audience,
		}
	case v20231101.DeviceCollectionDeviceTypeGcpVm:
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
	case v20231101.DeviceCollectionDeviceTypeTpm:
		tpm, err := collection.DeviceTypeConfiguration.AsTpm()
		if err != nil {
			diags.AddError("Read TPM Device Configuration", err.Error())
			return nil, diags
		}

		forceCN, d := utils.ToOptionalBool(ctx, tpm.ForceCN, state, path.Root("tpm").AtName("force_cn"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		requireEAB, d := utils.ToOptionalBool(ctx, tpm.RequireEAB, state, path.Root("tpm").AtName("require_eab"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		attestorRoots, d := utils.ToOptionalString(ctx, tpm.AttestorRoots, state, path.Root("tpm").AtName("attestor_roots"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		attestorIntermediates, d := utils.ToOptionalString(ctx, tpm.AttestorIntermediates, state, path.Root("tpm").AtName("attestor_intermediates"))
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		model.TPMDevice = &TPMDevice{
			AttestorRoots:         attestorRoots,
			AttestorIntermediates: attestorIntermediates,
			RequireEAB:            requireEAB,
			ForceCN:               forceCN,
		}
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20231101.DeviceCollection, diag.Diagnostics) {
	var diags diag.Diagnostics

	dc := &v20231101.DeviceCollection{
		Slug:        model.Slug.ValueString(),
		DisplayName: model.DisplayName.ValueString(),
		DeviceType:  v20231101.DeviceCollectionDeviceType(model.DeviceType.ValueString()),
		AuthorityID: model.AuthorityID.ValueString(),
	}

	switch dc.DeviceType {
	case v20231101.DeviceCollectionDeviceTypeAwsVm:
		if model.AWSDevice == nil {
			diags.AddError("AWS Device", "aws_vm is required with device type aws-vm")
			return nil, diags
		}
		aws := v20231101.AwsVM{
			Accounts:          []string{},
			DisableCustomSANs: model.AWSDevice.DisableCustomSANs.ValueBoolPointer(),
		}
		d := model.AWSDevice.Accounts.ElementsAs(ctx, &aws.Accounts, false)
		diags.Append(d...)

		if err := dc.DeviceTypeConfiguration.FromAwsVM(aws); err != nil {
			diags.AddError("AWS VM", err.Error())
			return nil, diags
		}
	case v20231101.DeviceCollectionDeviceTypeAzureVm:
		if model.AzureDevice == nil {
			diags.AddError("Azure Device", "azure_vm is required with device type azure-vm")
			return nil, diags
		}
		azure := v20231101.AzureVM{
			TenantID:          model.AzureDevice.TenantID.ValueString(),
			ResourceGroups:    []string{},
			DisableCustomSANs: model.AzureDevice.DisableCustomSANs.ValueBoolPointer(),
			Audience:          model.AzureDevice.Audience.ValueStringPointer(),
		}
		d := model.AzureDevice.ResourceGroups.ElementsAs(ctx, &azure.ResourceGroups, false)
		diags.Append(d...)

		if err := dc.DeviceTypeConfiguration.FromAzureVM(azure); err != nil {
			diags.AddError("Azure VM", err.Error())
			return nil, diags
		}
	case v20231101.DeviceCollectionDeviceTypeGcpVm:
		if model.GCPDevice == nil {
			diags.AddError("GCP Device", "gcp_vm is required with device type gcp-vm")
			return nil, diags
		}
		gcp := v20231101.GcpVM{
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
	case v20231101.DeviceCollectionDeviceTypeTpm:
		if model.TPMDevice == nil {
			diags.AddError("TPM Device", "tpm block is required with device type tpm")
		}
		tpm := v20231101.Tpm{
			AttestorRoots:         model.TPMDevice.AttestorRoots.ValueStringPointer(),
			AttestorIntermediates: model.TPMDevice.AttestorIntermediates.ValueStringPointer(),
			ForceCN:               model.TPMDevice.ForceCN.ValueBoolPointer(),
			RequireEAB:            model.TPMDevice.RequireEAB.ValueBoolPointer(),
		}
		if err := dc.DeviceTypeConfiguration.FromTpm(tpm); err != nil {
			diags.AddError("TPM", err.Error())
			return nil, diags
		}
	default:
		diags.AddError("Device Type", fmt.Sprintf("Unsupported device collection type %q", dc.DeviceType))
		return nil, diags
	}

	return dc, diags
}
