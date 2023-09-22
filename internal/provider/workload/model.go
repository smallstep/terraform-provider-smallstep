package workload

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/endpoint_configuration"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const typeName = "smallstep_workload"

type Model struct {
	ID                   types.String                                 `tfsdk:"id"`
	WorkloadType         types.String                                 `tfsdk:"workload_type"`
	DisplayName          types.String                                 `tfsdk:"display_name"`
	Slug                 types.String                                 `tfsdk:"slug"`
	DeviceCollectionSlug types.String                                 `tfsdk:"device_collection_slug"`
	CertificateInfo      *endpoint_configuration.CertificateInfoModel `tfsdk:"certificate_info"`
	KeyInfo              *endpoint_configuration.KeyInfoModel         `tfsdk:"key_info"`
	// ReloadInfo and Hooks are optional. Need to use Object type to support
	// the "unknown" state.
	ReloadInfo            types.Object `tfsdk:"reload_info"`
	Hooks                 types.Object `tfsdk:"hooks"`
	AdminEmails           types.Set    `tfsdk:"admin_emails"`
	DeviceMetadataKeySANs types.Set    `tfsdk:"device_metadata_key_sans"`
	StaticSANs            types.Set    `tfsdk:"static_sans"`
}

func fromAPI(ctx context.Context, ec *v20230301.EndpointConfiguration, state utils.AttributeGetter) (*Model, diag.Diagnostics) {
	var diags diag.Diagnostics

	ciDuration, d := utils.ToEqualString(ctx, ec.CertificateInfo.Duration, state, path.Root("certificate_info").AtName("duration"), utils.IsDurationEqual)
	diags = append(diags, d...)

	ciCrtFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.CrtFile, state, path.Root("certificate_info").AtName("crt_file"))
	diags = append(diags, d...)

	ciKeyFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.KeyFile, state, path.Root("certificate_info").AtName("key_file"))
	diags = append(diags, d...)

	ciRootFile, d := utils.ToOptionalString(ctx, ec.CertificateInfo.RootFile, state, path.Root("certificate_info").AtName("root_file"))
	diags = append(diags, d...)

	ciUID, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Uid, state, path.Root("certificate_info").AtName("uid"))
	diags = append(diags, d...)

	ciGID, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Gid, state, path.Root("certificate_info").AtName("gid"))
	diags = append(diags, d...)

	ciMode, d := utils.ToOptionalInt(ctx, ec.CertificateInfo.Mode, state, path.Root("certificate_info").AtName("mode"))
	diags = append(diags, d...)

	model := &Model{
		ID:          types.StringValue(utils.Deref(ec.Id)),
		DisplayName: types.StringValue(ec.Name),
		CertificateInfo: &endpoint_configuration.CertificateInfoModel{
			Type:     types.StringValue(string(ec.CertificateInfo.Type)),
			Duration: ciDuration,
			CrtFile:  ciCrtFile,
			KeyFile:  ciKeyFile,
			RootFile: ciRootFile,
			UID:      ciUID,
			GID:      ciGID,
			Mode:     ciMode,
		},
	}

	if ec.Hooks != nil {
		sign, d := endpoint_configuration.HookFromAPI(ctx, ec.Hooks.Sign, path.Root("hooks").AtName("sign"), state)
		diags.Append(d...)

		renew, d := endpoint_configuration.HookFromAPI(ctx, ec.Hooks.Renew, path.Root("hooks").AtName("renew"), state)
		diags.Append(d...)

		hooksObj, d := basetypes.NewObjectValue(endpoint_configuration.HooksObjectType, map[string]attr.Value{
			"sign":  sign,
			"renew": renew,
		})
		diags.Append(d...)

		model.Hooks = hooksObj
	} else {
		model.Hooks = basetypes.NewObjectNull(endpoint_configuration.HooksObjectType)
	}

	if ec.KeyInfo != nil {
		format, d := utils.ToOptionalString(ctx, ec.KeyInfo.Format, state, path.Root("key_info").AtName("format"))
		diags = append(diags, d...)

		pubFile, d := utils.ToOptionalString(ctx, ec.KeyInfo.PubFile, state, path.Root("key_info").AtName("pub_file"))
		diags = append(diags, d...)

		typ, d := utils.ToOptionalString(ctx, ec.KeyInfo.Type, state, path.Root("key_info").AtName("type"))
		diags = append(diags, d...)

		model.KeyInfo = &endpoint_configuration.KeyInfoModel{
			Format:  format,
			PubFile: pubFile,
			Type:    typ,
		}
	}

	if ec.ReloadInfo != nil {
		pidFile, d := utils.ToOptionalString(ctx, ec.ReloadInfo.PidFile, state, path.Root("reload_info").AtName("method"))
		diags.Append(d...)

		signal, d := utils.ToOptionalInt(ctx, ec.ReloadInfo.Signal, state, path.Root("reload_info").AtName("signal"))
		diags.Append(d...)

		reloadInfoObject, d := basetypes.NewObjectValue(endpoint_configuration.ReloadInfoType, map[string]attr.Value{
			"method":   types.StringValue(string(ec.ReloadInfo.Method)),
			"pid_file": pidFile,
			"signal":   signal,
		})
		diags.Append(d...)
		model.ReloadInfo = reloadInfoObject
	} else {
		model.ReloadInfo = basetypes.NewObjectNull(endpoint_configuration.ReloadInfoType)
	}

	return model, diags
}

func toAPI(ctx context.Context, model *Model) (*v20230301.Workload, diag.Diagnostics) {
	hooksModel := &endpoint_configuration.HooksModel{}
	diags := model.Hooks.As(ctx, &hooksModel, basetypes.ObjectAsOptions{
		UnhandledUnknownAsEmpty: true,
	})
	hooks, d := hooksModel.ToAPI(ctx)
	diags.Append(d...)

	reloadInfo := &endpoint_configuration.ReloadInfoModel{}
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
		Id:                    model.ID.ValueStringPointer(),
		DisplayName:           model.DisplayName.ValueString(),
		WorkloadType:          model.WorkloadType.ValueString(),
		Slug:                  model.Slug.ValueString(),
		CertificateInfo:       &ci,
		KeyInfo:               model.KeyInfo.ToAPI(),
		Hooks:                 hooks,
		ReloadInfo:            reloadInfo.ToAPI(),
		AdminEmails:           &adminEmails,
		DeviceMetadataKeySANs: &deviceMetadataKeySANs,
		StaticSANs:            staticSANs,
	}, diags
}
