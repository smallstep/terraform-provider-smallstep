package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
)

type createTokenReq struct {
	TeamID string   `json:"teamID"`
	Bundle [][]byte `json:"bundle"`
}

type createTokenResp struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

func apiClientWithClientCert(ctx context.Context, server, teamID, cert, key string) (*v20230301.Client, error) {
	// TODO move to validation
	if _, err := uuid.Parse(teamID); err != nil {
		return nil, fmt.Errorf("team-id argument must be a valid UUID")
	}

	clientCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}
	r := &createTokenReq{
		TeamID: teamID,
		Bundle: clientCert.Certificate,
	}
	err = json.NewEncoder(b).Encode(r)
	if err != nil {
		return nil, err
	}

	authURL, err := url.JoinPath(server, "auth")
	if err != nil {
		return nil, err
	}
	post, err := http.NewRequest("POST", authURL, b)
	if err != nil {
		return nil, err
	}
	post.Header.Set("Content-Type", "application/json")
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		GetClientCertificate: func(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
			return &clientCert, nil
		},
		MinVersion: tls.VersionTLS12,
	}
	client := http.Client{
		Transport: transport,
	}
	resp, err := client.Do(post)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody := &createTokenResp{}
	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return nil, err
	}
	if resp.StatusCode != 201 {
		if respBody.Message != "" {
			return nil, errors.New(respBody.Message)
		}
		return nil, fmt.Errorf("failed to create token: %d", resp.StatusCode)
	}

	apiClient, err := v20230301.NewClient(server, v20230301.WithRequestEditorFn(v20230301.RequestEditorFn(func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", respBody.Token))
		return nil
	})))
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}
