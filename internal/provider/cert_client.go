package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

type createTokenReq struct {
	TeamID string   `json:"teamID"`
	Bundle [][]byte `json:"bundle"`
}

type createTokenResp struct {
	Token   string `json:"token"`
	Message string `json:"message"`
}

// Uses a client cert to get an API token and returns a client using that token.
// Renews the token just before its 1 hour expiry in case of long running
// terraform applies.
func apiClientWithClientCert(ctx context.Context, server, teamID, cert, key string) (*v20250101.Client, error) {
	if _, err := uuid.Parse(teamID); err != nil {
		return nil, fmt.Errorf("team-id argument must be a valid UUID")
	}

	clientCert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		return nil, err
	}

	authURL, err := url.JoinPath(server, "auth")
	if err != nil {
		return nil, err
	}

	r := &createTokenReq{
		TeamID: teamID,
		Bundle: clientCert.Certificate,
	}

	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	var tkn string
	var m sync.RWMutex
	getTkn := func() error {
		post, err := http.NewRequest("POST", authURL, bytes.NewBuffer(b))
		if err != nil {
			return err
		}
		post.Header.Set("X-Smallstep-Api-Version", "2025-01-01")
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
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			msg := utils.APIErrorMsg(resp.Body)
			return fmt.Errorf("failed to create Smallstep API token with provided client certificate - the certificate may be expired or invalid. Response: %d. Details: %s", resp.StatusCode, msg)
		}

		respBody := &createTokenResp{}
		if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
			return err
		}

		m.Lock()
		tkn = respBody.Token
		m.Unlock()

		tflog.Info(ctx, "Created new Smallstep API token with client certificate")

		return nil
	}

	if err := getTkn(); err != nil {
		return nil, err
	}
	go func() {
		ticker := time.NewTicker(time.Minute * 59)

		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if err := getTkn(); err != nil {
				tflog.Error(ctx, "Failed to get Smallstep API token with client certificate")
			}
		}
	}()

	apiClient, err := v20250101.NewClient(server, v20250101.WithRequestEditorFn(v20250101.RequestEditorFn(func(ctx context.Context, r *http.Request) error {
		m.RLock()
		r.Header.Set("X-Smallstep-Api-Version", "2025-01-01")
		r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tkn))
		m.RUnlock()
		return nil
	})))
	if err != nil {
		return nil, err
	}

	return apiClient, nil
}
