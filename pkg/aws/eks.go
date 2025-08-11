package aws

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// EKSToken represents the token returned by aws eks get-token
type EKSToken struct {
	Kind       string   `json:"kind"`
	APIVersion string   `json:"apiVersion"`
	Spec       struct{} `json:"spec"`
	Status     struct {
		ExpirationTimestamp time.Time `json:"expirationTimestamp"`
		Token               string    `json:"token"`
	} `json:"status"`
}

// GetToken retrieves an EKS authentication token
func GetToken(clusterName, region, profile string) (*EKSToken, error) {
	args := []string{
		"eks", "get-token",
		"--cluster-name", clusterName,
		"--region", region,
	}

	if profile != "" {
		args = append(args, "--profile", profile)
	}

	cmd := exec.Command("aws", args...)
	output, err := cmd.Output()
	if err != nil {
		// Try to get stderr for more detailed error message
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get EKS token: %w\nstderr: %s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to get EKS token: %w", err)
	}

	var token EKSToken
	if err := json.Unmarshal(output, &token); err != nil {
		return nil, fmt.Errorf("failed to parse EKS token response: %w", err)
	}

	return &token, nil
}
