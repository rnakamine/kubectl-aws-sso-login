package kubernetes

import (
	"encoding/json"
	"fmt"
	"time"
)

// ExecCredential represents the Kubernetes ExecCredential format
// This is what kubectl expects from an exec credential plugin
type ExecCredential struct {
	APIVersion string               `json:"apiVersion"`
	Kind       string               `json:"kind"`
	Status     ExecCredentialStatus `json:"status"`
}

type ExecCredentialStatus struct {
	Token               string `json:"token"`
	ExpirationTimestamp string `json:"expirationTimestamp,omitempty"`
}

func NewExecCredential(token string, expirationTime time.Time) *ExecCredential {
	return &ExecCredential{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Kind:       "ExecCredential",
		Status: ExecCredentialStatus{
			Token:               token,
			ExpirationTimestamp: expirationTime.Format(time.RFC3339),
		},
	}
}

func (e *ExecCredential) PrintJSON() error {
	jsonData, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal ExecCredential to JSON: %w", err)
	}
	fmt.Println(string(jsonData))
	return nil
}
