package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
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
	slug := acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
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
