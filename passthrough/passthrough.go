// Package passthrough implements a socket-based credential helper.
package passthrough

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/docker/docker-credential-helpers/credentials"
)

// Passthrough only supports retrieving the credentials for a specific server
// URL, or listing all stored credentials.
type Passthrough struct{}

// Add adds new credentials to the keychain.
func (p Passthrough) Add(_ *credentials.Credentials) error {
	return errors.New("adding credentials is not supported")
}

// Delete removes credentials from the store.
func (p Passthrough) Delete(_ string) error {
	return errors.New("deleting credentials is not supported")
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/run/creds.sock")
			},
		},
	}
}

// Get returns the username and secret to use for a given registry server URL.
func (p Passthrough) Get(serverURL string) (string, string, error) {
	req := bytes.NewBufferString(fmt.Sprintf(`{"ServerURL": "%s"}`, serverURL))
	resp, err := newHTTPClient().Post("http://unix/get", "application/json", req)
	if err != nil {
		return "", "", fmt.Errorf("failed to query credentials for %s: %w", serverURL, err)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response body: %w", err)
	}

	var creds credentials.Credentials
	if err := json.Unmarshal(payload, &creds); err != nil {
		return "", "", fmt.Errorf("failed to decode credentials for %s: %w", serverURL, err)
	}

	return creds.Username, creds.Secret, nil
}

// List returns the stored URLs and corresponding usernames for a given credentials label
func (p Passthrough) List() (map[string]string, error) {
	resp, err := newHTTPClient().Post("http://unix/list", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}

	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var creds map[string]string
	if err := json.Unmarshal(payload, &creds); err != nil {
		return nil, fmt.Errorf("failed to decode the list of credentials: %w", err)
	}

	return creds, nil
}
