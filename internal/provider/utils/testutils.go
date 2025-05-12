package utils

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/stretchr/testify/require"
	"go.step.sm/crypto/jose"
	"go.step.sm/crypto/minica"
	"go.step.sm/crypto/pemutil"
	"go.step.sm/crypto/randutil"
)

var UUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

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

	return account
}
