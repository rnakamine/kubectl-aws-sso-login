package aws

import (
	"fmt"
	"os"
	"os/exec"
)

// SSOLogin performs AWS SSO login for the specified profile
func SSOLogin(profile string) error {
	fmt.Fprintf(os.Stderr, "Starting AWS SSO login...\n")
	args := []string{"sso", "login"}
	if profile != "" {
		args = append(args, "--profile", profile)
	}

	cmd := exec.Command("aws", args...)
	// Redirect AWS CLI output to stderr to avoid mixing with our JSON output
	// AWS SSO login outputs progress messages to stdout, but we need to keep
	// our stdout clean for the ExecCredential JSON that kubectl expects
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS SSO login failed: %w", err)
	}

	return nil
}
