package utils

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/netip"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.step.sm/crypto/jose"
	"go.step.sm/crypto/minica"
	"go.step.sm/crypto/pemutil"
	"go.step.sm/crypto/randutil"
)

var UUIDRegexp = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
var CARegexp = regexp.MustCompile(`-----BEGIN CERTIFICATE-----`)
var IPv4Regexp = regexp.MustCompile(`^\d+\.\d+\.\d+\.\d+$`)

func SmallstepAPIClientFromEnv() (*v20250101.Client, error) {
	token := os.Getenv("SMALLSTEP_API_TOKEN")
	if token == "" {
		return nil, errors.New("missing environment variable SMALLSTEP_API_TOKEN")
	}
	server := os.Getenv("SMALLSTEP_API_URL")
	if server == "" {
		return nil, errors.New("missing environment variable SMALLSTEP_API_URL")
	}

	client, err := v20250101.NewClient(server, v20250101.WithRequestEditorFn(v20250101.RequestEditorFn(func(ctx context.Context, r *http.Request) error {
		r.Header.Set("X-Smallstep-Api-Version", "2025-01-01")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	})))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewAuthority(t *testing.T) *v20250101.Authority {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	slug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	reqBody := v20250101.PostAuthoritiesJSONRequestBody{
		Name:        slug + " Authority",
		AdminEmails: []string{"eng@smallstep.com"},
		Subdomain:   slug,
		Type:        "devops",
	}
	resp, err := client.PostAuthorities(context.Background(), &v20250101.PostAuthoritiesParams{}, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create authority: %d: %s", resp.StatusCode, body)
	}

	authority := &v20250101.Authority{}
	err = json.NewDecoder(resp.Body).Decode(authority)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		resp, err := client.DeleteAuthority(context.Background(), authority.Id, &v20250101.DeleteAuthorityParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return authority
}

func NewOIDCProvisioner(t *testing.T, authorityID string) (*v20250101.Provisioner, *v20250101.OidcProvisioner) {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	req := v20250101.Provisioner{
		Name: acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Type: "OIDC",
	}
	oidc := v20250101.OidcProvisioner{
		ClientID:              acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		ClientSecret:          acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		ConfigurationEndpoint: "https://accounts.google.com/.well-known/openid-configuration",
	}
	if err := req.FromOidcProvisioner(oidc); err != nil {
		t.Fatal(err)
	}
	resp, err := client.PostAuthorityProvisioners(context.Background(), authorityID, &v20250101.PostAuthorityProvisionersParams{}, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create provisioner: %d: %s", resp.StatusCode, body)
	}

	provisioner := &v20250101.Provisioner{}
	if err := json.NewDecoder(resp.Body).Decode(&provisioner); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		resp, err := client.DeleteProvisioner(context.Background(), authorityID, *provisioner.Id, &v20250101.DeleteProvisionerParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return provisioner, &oidc
}

func NewJWK(t *testing.T, pass string) (string, string) {
	jwk, jwe, err := jose.GenerateDefaultKeyPair([]byte(pass))
	require.NoError(t, err)

	pubJSON, err := jwk.MarshalJSON()
	require.NoError(t, err)

	priv, err := jwe.CompactSerialize()
	require.NoError(t, err)

	return string(pubJSON), priv
}

// CACerts returns a new root and intermediate pem-encoded certificate
func CACerts(t *testing.T) (string, string) {
	ca, err := minica.New()
	require.NoError(t, err)

	root, err := pemutil.Serialize(ca.Root)
	require.NoError(t, err)

	intermediate, err := pemutil.Serialize(ca.Intermediate)
	require.NoError(t, err)

	return string(pem.EncodeToMemory(root)), string(pem.EncodeToMemory(intermediate))
}

func NewWebhook(t *testing.T, provisionerID, authorityID string) *v20250101.ProvisionerWebhook {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	url := "https://example.com/hook"
	name, err := randutil.Alphanumeric(8)
	require.NoError(t, err)

	req := v20250101.ProvisionerWebhook{
		Name:       name,
		Url:        &url,
		Kind:       "ENRICHING",
		CertType:   "ALL",
		ServerType: "EXTERNAL",
	}

	resp, err := client.PostWebhooks(context.Background(), authorityID, provisionerID, &v20250101.PostWebhooksParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode, "got %d: %s", resp.StatusCode, body)

	wh := &v20250101.ProvisionerWebhook{}
	require.NoError(t, json.Unmarshal(body, wh))
	wh.Secret = nil

	return wh
}

func Slug(t *testing.T) string {
	slug, err := randutil.String(10, "abcdefghijklmnopqrstuvwxyz0123456789")
	require.NoError(t, err)
	return "tfprovider" + slug
}

func IP(t *testing.T) string {
	size := 4
	if rand.Int31n(2) == 0 {
		size = 16
	}
	ip, err := randutil.Bytes(size)
	require.NoError(t, err)
	addr, ok := netip.AddrFromSlice(ip)
	require.True(t, ok)
	return addr.String()
}

func NewDevice(t *testing.T) *v20250101.Device {
	t.Helper()

	deviceName, err := randutil.Alphanumeric(12)
	require.NoError(t, err)
	permanentID, err := randutil.Alphanumeric(12)
	require.NoError(t, err)
	displayID, err := randutil.Alphanumeric(12)
	require.NoError(t, err)
	serial, err := randutil.Alphanumeric(12)
	require.NoError(t, err)

	req := v20250101.DeviceRequest{
		PermanentIdentifier: permanentID,
		DisplayId:           Ref(displayID),
		DisplayName:         Ref(deviceName),
		Metadata: &v20250101.DeviceMetadata{
			"k1": "v1",
		},
		Tags:      Ref([]string{"ubuntu"}),
		Os:        Ref(v20250101.Linux),
		Ownership: Ref(v20250101.User),
		Serial:    Ref(serial),
		User: &v20250101.DeviceUser{
			Email: "employee@example.com",
		},
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PostDevices(context.Background(), &v20250101.PostDevicesParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 201, resp.StatusCode, string(body))

	device := &v20250101.Device{}
	err = json.Unmarshal(body, device)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteDevice(context.Background(), device.Id, &v20250101.DeleteDeviceParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return device
}

func NewAccount(t *testing.T) *v20250101.Account {
	t.Helper()

	req := v20250101.AccountRequest{
		Name: "An Account",
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PostAccounts(context.Background(), &v20250101.PostAccountsParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 201, resp.StatusCode, string(body))

	account := &v20250101.Account{}
	err = json.Unmarshal(body, account)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteAccount(context.Background(), account.Id, &v20250101.DeleteAccountParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return account
}

func NewManagedRADIUS(t *testing.T) *v20250101.ManagedRadius {
	t.Helper()
	ca, _ := CACerts(t)

	reqBody := v20250101.ManagedRadius{
		NasIPs:   []string{IP(t)},
		Name:     Slug(t),
		ClientCA: ca,
		ReplyAttributes: &[]v20250101.ReplyAttribute{
			{
				Name:  "Tunnel-Type",
				Value: Ref("13"),
			},
			{
				Name:                 "Tunnel-Private-Group-ID",
				ValueFromCertificate: Ref("2.5.4.11"),
			},
		},
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PostManagedRadius(t.Context(), &v20250101.PostManagedRadiusParams{}, reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 201, resp.StatusCode, string(body))

	radius := &v20250101.ManagedRadius{}
	err = json.Unmarshal(body, radius)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteManagedRadius(context.Background(), *radius.Id, &v20250101.DeleteManagedRadiusParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	resp, err = client.GetManagedRadius(t.Context(), *radius.Id, &v20250101.GetManagedRadiusParams{
		Secret: Ref(true),
	})
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode, string(body))

	radius = &v20250101.ManagedRadius{}
	err = json.Unmarshal(body, radius)
	require.NoError(t, err)

	return radius
}

func NewIdentityProvider(t *testing.T) *v20250101.IdentityProvider {
	t.Helper()
	ca, _ := CACerts(t)

	reqBody := v20250101.IdentityProvider{
		TrustRoots: ca,
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PutIdentityProvider(t.Context(), &v20250101.PutIdentityProviderParams{}, reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 200, resp.StatusCode, string(body))

	idp := &v20250101.IdentityProvider{}
	err = json.Unmarshal(body, idp)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteIdentityProvider(context.Background(), &v20250101.DeleteIdentityProviderParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return idp
}

func NewIdentityProviderClient(t *testing.T) *v20250101.IdpClient {
	t.Helper()

	reqBody := v20250101.IdpClient{
		RedirectURI: "https://example.com/" + Slug(t),
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PostIdpClients(t.Context(), &v20250101.PostIdpClientsParams{}, reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 201, resp.StatusCode, string(body))

	c := &v20250101.IdpClient{}
	err = json.Unmarshal(body, c)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteIdpClient(context.Background(), *c.Id, &v20250101.DeleteIdpClientParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return c
}

func NewCredential(t *testing.T) *v20250101.Credential {
	t.Helper()

	reqBody := v20250101.Credential{
		Slug: "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Certificate: v20250101.CredentialCertificate{
			Duration: "1h",
			Type:     "X509",
		},
		Key: v20250101.CredentialKey{
			Type:       Ref(v20250101.CredentialKeyType("ECDSA_P256")),
			Protection: Ref(v20250101.CredentialKeyProtection("NONE")),
		},
	}

	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.PostCredentials(t.Context(), &v20250101.PostCredentialsParams{}, reqBody)
	require.NoError(t, err)
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, 201, resp.StatusCode, string(body))

	c := &v20250101.Credential{}
	err = json.Unmarshal(body, c)
	require.NoError(t, err)

	t.Cleanup(func() {
		resp, err := client.DeleteCredential(context.Background(), *c.Id, &v20250101.DeleteCredentialParams{})
		require.NoError(t, err)
		assert.Equal(t, 204, resp.StatusCode)
	})

	return c
}
