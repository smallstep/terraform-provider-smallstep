package utils

import (
	"context"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/stretchr/testify/require"
	"go.step.sm/crypto/jose"
	"go.step.sm/crypto/minica"
	"go.step.sm/crypto/pemutil"
	"go.step.sm/crypto/randutil"
)

func SmallstepAPIClientFromEnv() (*v20230301.Client, error) {
	token := os.Getenv("SMALLSTEP_API_TOKEN")
	if token == "" {
		return nil, errors.New("Missing environment variable SMALLSTEP_API_TOKEN")
	}
	server := os.Getenv("SMALLSTEP_API_URL")
	if server == "" {
		return nil, errors.New("Missing environment variable SMALLSTEP_API_URL")
	}

	client, err := v20230301.NewClient(server, v20230301.WithRequestEditorFn(v20230301.RequestEditorFn(func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		return nil
	})))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func NewAuthority(t *testing.T) *v20230301.Authority {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	slug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	reqBody := v20230301.PostAuthoritiesJSONRequestBody{
		Name:        slug + " Authority",
		AdminEmails: []string{"eng@smallstep.com"},
		Subdomain:   slug,
		Type:        "devops",
	}
	resp, err := client.PostAuthorities(context.Background(), &v20230301.PostAuthoritiesParams{}, reqBody)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create authority: %d: %s", resp.StatusCode, body)
	}

	authority := &v20230301.Authority{}
	err = json.NewDecoder(resp.Body).Decode(authority)
	if err != nil {
		t.Fatal(err)
	}

	return authority
}

func NewOIDCProvisioner(t *testing.T, authorityID string) (*v20230301.Provisioner, *v20230301.OidcProvisioner) {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	req := v20230301.Provisioner{
		Name: acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		Type: "OIDC",
	}
	oidc := v20230301.OidcProvisioner{
		ClientID:              acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		ClientSecret:          acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum),
		ConfigurationEndpoint: "https://accounts.google.com/.well-known/openid-configuration",
	}
	if err := req.FromOidcProvisioner(oidc); err != nil {
		t.Fatal(err)
	}
	resp, err := client.PostAuthorityProvisioners(context.Background(), authorityID, &v20230301.PostAuthorityProvisionersParams{}, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Failed to create provisioner: %d: %s", resp.StatusCode, body)
	}

	provisioner := &v20230301.Provisioner{}
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

func NewWebhook(t *testing.T, provisionerID, authorityID string) *v20230301.ProvisionerWebhook {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	url := "https://example.com/hook"
	name, err := randutil.Alphanumeric(8)
	require.NoError(t, err)

	req := v20230301.ProvisionerWebhook{
		Name:       name,
		Url:        &url,
		Kind:       "ENRICHING",
		CertType:   "ALL",
		ServerType: "EXTERNAL",
	}

	resp, err := client.PostWebhooks(context.Background(), authorityID, provisionerID, &v20230301.PostWebhooksParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode, "got %d: %s", resp.StatusCode, body)

	wh := &v20230301.ProvisionerWebhook{}
	require.NoError(t, json.Unmarshal(body, wh))
	wh.Secret = nil

	return wh
}

func NewCollection(t *testing.T) *v20230301.Collection {
	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	slug := Slug(t)
	displayName := "Collection " + slug

	req := v20230301.NewCollection{
		Slug:        slug,
		DisplayName: &displayName,
	}

	resp, err := client.PostCollections(context.Background(), &v20230301.PostCollectionsParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode, "got %d: %s", resp.StatusCode, body)

	collection := &v20230301.Collection{}
	err = json.Unmarshal(body, collection)
	require.NoError(t, err)

	return collection
}

func NewInstance(t *testing.T, slug string) *v20230301.CollectionInstance {
	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	id, err := randutil.Alphanumeric(8)
	require.NoError(t, err)

	req := v20230301.PutCollectionInstanceJSONRequestBody{
		Data: map[string]string{"id": id},
	}
	resp, err := client.PutCollectionInstance(context.Background(), slug, id, &v20230301.PutCollectionInstanceParams{}, req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, "got %d: %s", resp.StatusCode, body)

	instance := &v20230301.CollectionInstance{}
	err = json.Unmarshal(body, instance)
	require.NoError(t, err)

	return instance
}

func Slug(t *testing.T) string {
	slug, err := randutil.String(10, "abcdefghijklmnopqrstuvwxyz0123456789")
	require.NoError(t, err)
	return "tfprovider" + slug
}

// There can only be 1 per team - don't try to create a new one if one exists
func FixAttestationAuthority(t *testing.T, catalog string) *v20230301.AttestationAuthority {
	client, err := SmallstepAPIClientFromEnv()
	require.NoError(t, err)

	resp, err := client.GetAttestationAuthorities(context.Background(), &v20230301.GetAttestationAuthoritiesParams{})
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusOK, resp.StatusCode, "list attestation authorities got %d: %s", resp.StatusCode, body)

	list := []*v20230301.AttestationAuthority{}
	err = json.Unmarshal(body, &list)
	require.NoError(t, err)

	if len(list) > 0 {
		return list[0]
	}

	root, intermediate := CACerts(t)

	req := v20230301.AttestationAuthority{
		Name:                  "tfprovider",
		AttestorRoots:         root,
		AttestorIntermediates: &intermediate,
		Catalog:               catalog,
	}

	resp, err = client.PostAttestationAuthorities(context.Background(), &v20230301.PostAttestationAuthoritiesParams{}, req)
	require.NoError(t, err)
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode, "create attestation authority got %d: %s", resp.StatusCode, body)

	aa := &v20230301.AttestationAuthority{}
	err = json.Unmarshal(body, aa)
	require.NoError(t, err)

	return aa
}

func SweepAttestationAuthorities() error {
	client, err := SmallstepAPIClientFromEnv()
	if err != nil {
		return err
	}

	resp, err := client.GetAttestationAuthorities(context.Background(), &v20230301.GetAttestationAuthoritiesParams{})
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("list attestation authorities got %d: %s", resp.StatusCode, body)
	}

	list := []*v20230301.AttestationAuthority{}
	if err := json.Unmarshal(body, &list); err != nil {
		return err
	}

	for _, aa := range list {
		if !strings.HasPrefix(aa.Name, "tfprovider") {
			continue
		}
		resp, err := client.DeleteAttestationAuthority(context.Background(), *aa.Id, &v20230301.DeleteAttestationAuthorityParams{})
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusNoContent {
			return fmt.Errorf("delete attestation authority %q got %d: %s", *aa.Id, resp.StatusCode, body)
		}
		log.Printf("Successfully swept attestation authority %s\n", *aa.Id)
	}

	return nil
}
