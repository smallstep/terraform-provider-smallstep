package authority

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
)

// type name for both resources and data sources
const authorityTypeName = "smallstep_authority"

type DataModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Type             types.String `tfsdk:"type"`
	Domain           types.String `tfsdk:"domain"`
	Fingerprint      types.String `tfsdk:"fingerprint"`
	Root             types.String `tfsdk:"root"`
	CreatedAt        types.String `tfsdk:"created_at"`
	ActiveRevocation types.Bool   `tfsdk:"active_revocation"`
	AdminEmails      types.Set    `tfsdk:"admin_emails"`
}

type ResourceModel struct {
	ID                 types.String     `tfsdk:"id"`
	Name               types.String     `tfsdk:"name"`
	Type               types.String     `tfsdk:"type"`
	Subdomain          types.String     `tfsdk:"subdomain"`
	Domain             types.String     `tfsdk:"domain"`
	Fingerprint        types.String     `tfsdk:"fingerprint"`
	Root               types.String     `tfsdk:"root"`
	CreatedAt          types.String     `tfsdk:"created_at"`
	ActiveRevocation   types.Bool       `tfsdk:"active_revocation"`
	AdminEmails        types.Set        `tfsdk:"admin_emails"`
	IntermediateIssuer *X509IssuerModel `tfsdk:"intermediate_issuer"`
	RootIssuer         *X509IssuerModel `tfsdk:"root_issuer"`
}

type X509IssuerModel struct {
	Name            types.String            `tfsdk:"name"`
	Duration        types.String            `tfsdk:"duration"`
	KeyVersion      types.String            `tfsdk:"key_version"`
	MaxPathLength   types.Int64             `tfsdk:"max_path_length"`
	NameConstraints *NameConstraintsModel   `tfsdk:"name_constraints"`
	Subject         *DistinguishedNameModel `tfsdk:"subject"`
}

func (issuer *X509IssuerModel) AsAPI(ctx context.Context) (*v20250101.X509Issuer, diag.Diagnostics) {
	if issuer == nil {
		return nil, diag.Diagnostics{}
	}

	nameConstraints, diagnostics := issuer.NameConstraints.AsAPI(ctx)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	maxPathLength := int(issuer.MaxPathLength.ValueInt64())

	return &v20250101.X509Issuer{
		Duration:        issuer.Duration.ValueStringPointer(),
		KeyVersion:      v20250101.X509IssuerKeyVersion(issuer.KeyVersion.ValueString()),
		MaxPathLength:   &maxPathLength,
		Name:            issuer.Name.ValueString(),
		NameConstraints: nameConstraints,
		Subject:         issuer.Subject.AsAPI(),
	}, diagnostics
}

type NameConstraintsModel struct {
	Critical                types.Bool `tfsdk:"critical"`
	ExcludedDNSDomains      types.Set  `tfsdk:"excluded_dns_domains"`
	ExcludedEmailAddresses  types.Set  `tfsdk:"excluded_email_addresses"`
	ExcludedIPRanges        types.Set  `tfsdk:"excluded_ip_ranges"`
	ExcludedURIDomains      types.Set  `tfsdk:"excluded_uri_domains"`
	PermittedDNSDomains     types.Set  `tfsdk:"permitted_dns_domains"`
	PermittedEmailAddresses types.Set  `tfsdk:"permitted_email_addresses"`
	PermittedIPRanges       types.Set  `tfsdk:"permitted_ip_ranges"`
	PermittedURIDomains     types.Set  `tfsdk:"permitted_uri_domains"`
}

func (nc *NameConstraintsModel) AsAPI(ctx context.Context) (*v20250101.NameConstraints, diag.Diagnostics) {
	var d diag.Diagnostics

	if nc == nil {
		return nil, d
	}

	var excludedDNSDomains *[]string
	var excludedEmailAddresses *[]string
	var excludedIPRanges *[]string
	var excludedURIDomains *[]string
	var permittedDNSDomains *[]string
	var permittedEmailAddresses *[]string
	var permittedIPRanges *[]string
	var permittedURIDomains *[]string

	d.Append(nc.ExcludedDNSDomains.ElementsAs(ctx, &excludedDNSDomains, false)...)
	d.Append(nc.ExcludedEmailAddresses.ElementsAs(ctx, &excludedEmailAddresses, false)...)
	d.Append(nc.ExcludedIPRanges.ElementsAs(ctx, &excludedIPRanges, false)...)
	d.Append(nc.ExcludedURIDomains.ElementsAs(ctx, &excludedURIDomains, false)...)
	d.Append(nc.PermittedDNSDomains.ElementsAs(ctx, &permittedDNSDomains, false)...)
	d.Append(nc.PermittedEmailAddresses.ElementsAs(ctx, &permittedEmailAddresses, false)...)
	d.Append(nc.PermittedIPRanges.ElementsAs(ctx, &permittedIPRanges, false)...)
	d.Append(nc.PermittedURIDomains.ElementsAs(ctx, &permittedURIDomains, false)...)
	if d.HasError() {
		return nil, d
	}

	return &v20250101.NameConstraints{
		Critical:                nc.Critical.ValueBoolPointer(),
		ExcludedDNSDomains:      excludedDNSDomains,
		ExcludedEmailAddresses:  excludedEmailAddresses,
		ExcludedIPRanges:        excludedIPRanges,
		ExcludedURIDomains:      excludedURIDomains,
		PermittedDNSDomains:     permittedDNSDomains,
		PermittedEmailAddresses: permittedEmailAddresses,
		PermittedIPRanges:       permittedIPRanges,
		PermittedURIDomains:     permittedURIDomains,
	}, d
}

type DistinguishedNameModel struct {
	CommonName         types.String `tfsdk:"common_name"`
	Country            types.String `tfsdk:"country"`
	EmailAddress       types.String `tfsdk:"email_address"`
	Locality           types.String `tfsdk:"locality"`
	Organization       types.String `tfsdk:"organization"`
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	PostalCode         types.String `tfsdk:"postal_code"`
	Province           types.String `tfsdk:"province"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	StreetAddress      types.String `tfsdk:"street_address"`
}

func (dn *DistinguishedNameModel) AsAPI() *v20250101.DistinguishedName {
	if dn == nil {
		return nil
	}

	return &v20250101.DistinguishedName{
		CommonName:         dn.CommonName.ValueStringPointer(),
		Country:            dn.Country.ValueStringPointer(),
		EmailAddress:       dn.EmailAddress.ValueStringPointer(),
		Locality:           dn.Locality.ValueStringPointer(),
		Organization:       dn.Organization.ValueStringPointer(),
		OrganizationalUnit: dn.OrganizationalUnit.ValueStringPointer(),
		PostalCode:         dn.PostalCode.ValueStringPointer(),
		Province:           dn.Province.ValueStringPointer(),
		SerialNumber:       dn.SerialNumber.ValueStringPointer(),
		StreetAddress:      dn.StreetAddress.ValueStringPointer(),
	}
}
