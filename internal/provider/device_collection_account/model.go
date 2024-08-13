package device_collection_account

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/workload"
)

const name = "smallstep_device_collection_account"

type Model struct {
	Slug                 types.String `tfsdk:"slug"`
	AccountID            types.String `tfsdk:"account_id"`
	DeviceCollectionSlug types.String `tfsdk:"device_collection_slug"`
	AuthorityID          types.String `tfsdk:"authority_id"`
	DisplayName          types.String `tfsdk:"display_name"`
	CertificateInfo      types.Object `tfsdk:"certificate_info"`
	KeyInfo              types.Object `tfsdk:"key_info"`
	ReloadInfo           types.Object `tfsdk:"reload_info"`
	CertificateData      types.Object `tfsdk:"certificate_data"`
}

func toAPI(ctx context.Context, model *Model) (*v20231101.DeviceCollectionAccount, diag.Diagnostics) {
	var diags diag.Diagnostics

	reloadInfo := &workload.ReloadInfoModel{}
	ds := model.ReloadInfo.As(ctx, &reloadInfo, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(ds...)

	certInfo := &workload.CertificateInfoModel{}
	ds = model.CertificateInfo.As(ctx, &certInfo, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	keyInfo := &workload.KeyInfoModel{}
	ds = model.KeyInfo.As(ctx, &keyInfo, basetypes.ObjectAsOptions{})
	diags.Append(ds...)

	dca := &v20231101.DeviceCollectionAccount{
		Slug:            model.Slug.ValueString(),
		AccountID:       model.AccountID.ValueString(),
		AuthorityID:     model.AuthorityID.ValueStringPointer(),
		DisplayName:     model.DisplayName.ValueString(),
		CertificateInfo: certInfo.ToAPI(),
		KeyInfo:         keyInfo.ToAPI(),
		ReloadInfo:      reloadInfo.ToAPI(),
	}

	certDataModel := &workload.CertificateDataModel{}
	diags = model.CertificateData.As(ctx, &certDataModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	x509Fields, ds := certDataModel.ToAPI(ctx)
	diags.Append(ds...)
	err := dca.FromX509Fields(x509Fields)
	if err != nil {
		diags.AddError("Device Collection Account Certificate Data", err.Error())
		return nil, diags
	}

	return dca, diags
}

func fromAPI(ctx context.Context, dca *v20231101.DeviceCollectionAccount, dcSlug string, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	model := &Model{
		Slug:                 types.StringValue(dca.Slug),
		AccountID:            types.StringValue(dca.AccountID),
		DeviceCollectionSlug: types.StringValue(dcSlug),
		DisplayName:          types.StringValue(dca.DisplayName),
		AuthorityID:          types.StringValue(utils.Deref(dca.AuthorityID)),
	}

	x509Fields, err := dca.AsX509Fields()
	if err != nil {
		diags.AddError("parse device collection account response", err.Error())
		return nil, diags
	}

	certInfo, ds := workload.CertInfoFromAPI(ctx, dca.CertificateInfo, state)
	diags.Append(ds...)
	model.CertificateInfo = certInfo

	certData, ds := workload.CertDataFromAPI(ctx, x509Fields, state)
	diags.Append(ds...)
	model.CertificateData = certData

	keyInfo, ds := workload.KeyInfoFromAPI(ctx, dca.KeyInfo, state)
	diags.Append(ds...)
	model.KeyInfo = keyInfo

	reloadInfo, ds := workload.ReloadInfoFromAPI(ctx, dca.ReloadInfo, state)
	diags.Append(ds...)
	model.ReloadInfo = reloadInfo

	return model, diags
}
