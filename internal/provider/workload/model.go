package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_workload"

type Model struct {
	ID                   types.String          `tfsdk:"id"`
	WorkloadType         types.String          `tfsdk:"workload_type"`
	DisplayName          types.String          `tfsdk:"display_name"`
	Slug                 types.String          `tfsdk:"slug"`
	DeviceCollectionSlug types.String          `tfsdk:"device_collection_slug"`
	CertificateInfo      *CertificateInfoModel `tfsdk:"certificate_info"`
	KeyInfo              *KeyInfoModel         `tfsdk:"key_info"`
	// ReloadInfo and Hooks are optional. Need to use Object type to support
	// the "unknown" state.
	ReloadInfo            types.Object `tfsdk:"reload_info"`
	Hooks                 types.Object `tfsdk:"hooks"`
	AdminEmails           types.Set    `tfsdk:"admin_emails"`
	DeviceMetadataKeySANs types.Set    `tfsdk:"device_metadata_key_sans"`
	StaticSANs            types.List   `tfsdk:"static_sans"`
}

func fromAPI(ctx context.Context, workload *v20230301.Workload, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	ciDuration, d := utils.ToEqualString(ctx, workload.CertificateInfo.Duration, state, path.Root("certificate_info").AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	ciCrtFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.CrtFile, state, path.Root("certificate_info").AtName("crt_file"))
	diags = append(diags, d...)

	ciKeyFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.KeyFile, state, path.Root("certificate_info").AtName("key_file"))
	diags = append(diags, d...)

	ciRootFile, d := utils.ToOptionalString(ctx, workload.CertificateInfo.RootFile, state, path.Root("certificate_info").AtName("root_file"))
	diags = append(diags, d...)

	ciUID, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Uid, state, path.Root("certificate_info").AtName("uid"))
	diags = append(diags, d...)

	ciGID, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Gid, state, path.Root("certificate_info").AtName("gid"))
	diags = append(diags, d...)

	ciMode, d := utils.ToOptionalInt(ctx, workload.CertificateInfo.Mode, state, path.Root("certificate_info").AtName("mode"))
	diags = append(diags, d...)

	model := &Model{
		ID:           types.StringValue(workload.Slug),
		Slug:         types.StringValue(workload.Slug),
		WorkloadType: types.StringValue(workload.WorkloadType),
		DisplayName:  types.StringValue(workload.DisplayName),
		CertificateInfo: &CertificateInfoModel{
			Type:     types.StringValue(string(workload.CertificateInfo.Type)),
			Duration: ciDuration,
			CrtFile:  ciCrtFile,
			KeyFile:  ciKeyFile,
			RootFile: ciRootFile,
			UID:      ciUID,
			GID:      ciGID,
			Mode:     ciMode,
		},
	}

	model.StaticSANs, d = utils.ToOptionalList(ctx, workload.StaticSANs, state, path.Root("static_sans"))
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	model.DeviceMetadataKeySANs, d = utils.ToOptionalSet(ctx, workload.DeviceMetadataKeySANs, state, path.Root("device_metadata_key_sans"))
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	if workload.Hooks != nil {
		sign, d := HookFromAPI(ctx, workload.Hooks.Sign, path.Root("hooks").AtName("sign"), state)
		diags.Append(d...)

		renew, d := HookFromAPI(ctx, workload.Hooks.Renew, path.Root("hooks").AtName("renew"), state)
		diags.Append(d...)

		hooksObj, d := basetypes.NewObjectValue(HooksObjectType, map[string]attr.Value{
			"sign":  sign,
			"renew": renew,
		})
		diags.Append(d...)

		model.Hooks = hooksObj
	} else {
		model.Hooks = basetypes.NewObjectNull(HooksObjectType)
	}

	if workload.KeyInfo != nil {
		format, d := utils.ToOptionalString(ctx, workload.KeyInfo.Format, state, path.Root("key_info").AtName("format"))
		diags = append(diags, d...)

		pubFile, d := utils.ToOptionalString(ctx, workload.KeyInfo.PubFile, state, path.Root("key_info").AtName("pub_file"))
		diags = append(diags, d...)

		typ, d := utils.ToOptionalString(ctx, workload.KeyInfo.Type, state, path.Root("key_info").AtName("type"))
		diags = append(diags, d...)

		model.KeyInfo = &KeyInfoModel{
			Format:  format,
			PubFile: pubFile,
			Type:    typ,
		}
	}

	if workload.ReloadInfo != nil {
		pidFile, d := utils.ToOptionalString(ctx, workload.ReloadInfo.PidFile, state, path.Root("reload_info").AtName("method"))
		diags.Append(d...)

		signal, d := utils.ToOptionalInt(ctx, workload.ReloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
		diags.Append(d...)

		unitName, d := utils.ToOptionalString(ctx, workload.ReloadInfo.UnitName, state, path.Root("reload_info").AtName("unit_name"))
		diags.Append(d...)

		reloadInfoObject, d := basetypes.NewObjectValue(ReloadInfoType, map[string]attr.Value{
			"method":    types.StringValue(string(workload.ReloadInfo.Method)),
			"pid_file":  pidFile,
			"signal":    signal,
			"unit_name": unitName,
		})
		diags.Append(d...)
		model.ReloadInfo = reloadInfoObject
	} else {
		model.ReloadInfo = basetypes.NewObjectNull(ReloadInfoType)
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.Workload, diag.Diagnostics) {
	hooksModel := &HooksModel{}
	diags := model.Hooks.As(ctx, &hooksModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	hooks, d := hooksModel.ToAPI(ctx)
	diags.Append(d...)

	reloadInfo := &ReloadInfoModel{}
	d = model.ReloadInfo.As(ctx, &reloadInfo, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	diags.Append(d...)

	var adminEmails []string
	diags.Append(model.AdminEmails.ElementsAs(ctx, &adminEmails, false)...)

	var deviceMetadataKeySANs []string
	diags.Append(model.DeviceMetadataKeySANs.ElementsAs(ctx, &deviceMetadataKeySANs, false)...)

	var staticSANs []string
	diags.Append(model.StaticSANs.ElementsAs(ctx, &staticSANs, false)...)

	ci := model.CertificateInfo.ToAPI()

	return &v20230301.Workload{
		DisplayName:           model.DisplayName.ValueString(),
		WorkloadType:          model.WorkloadType.ValueString(),
		Slug:                  model.Slug.ValueString(),
		CertificateInfo:       &ci,
		KeyInfo:               model.KeyInfo.ToAPI(),
		Hooks:                 hooks,
		ReloadInfo:            reloadInfo.ToAPI(),
		AdminEmails:           &adminEmails,
		DeviceMetadataKeySANs: &deviceMetadataKeySANs,
		StaticSANs:            &staticSANs,
	}, diags
}
