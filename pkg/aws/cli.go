package aws

import (
	"fmt"
	"os/exec"
)

// CheckAWSCLI checks if AWS CLI is installed and available
func CheckAWSCLI() error {
	cmd := exec.Command("aws", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("AWS CLI is not installed or not in PATH: %w", err)
	}

	return nil
}
