/*
Copyright © 2025 Ryo Nakamine <rnakamine8080@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/rnakamine/kubectl-aws-sso-auth/internal/aws"
	"github.com/rnakamine/kubectl-aws-sso-auth/internal/kubernetes"
	"github.com/rnakamine/kubectl-aws-sso-auth/internal/sso"
	"github.com/spf13/cobra"
)

var (
	clusterName string
	region      string
)

var getTokenCmd = &cobra.Command{
	Use:   "get-token",
	Short: "Get EKS authentication token with AWS SSO",
	Long:  `Get EKS authentication token with AWS SSO`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if clusterName == "" {
			fmt.Fprintf(os.Stderr, "Error: --cluster-name flag is required\n")
			return fmt.Errorf("cluster-name is required")
		}
		if region == "" {
			fmt.Fprintf(os.Stderr, "Error: --region flag is required\n")
			return fmt.Errorf("region is required")
		}

		profile := os.Getenv("AWS_PROFILE")

		if err := aws.CheckAWSCLI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		if err := sso.FindValidSession(); err != nil {
			fmt.Fprintf(os.Stderr, "SSO session status: %v\n", err)

			if err := aws.SSOLogin(profile); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return err
			}

			if err := sso.FindValidSession(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: Failed to find valid session after login: %v\n", err)
				return err
			}
		}

		eksToken, err := aws.GetToken(clusterName, region, profile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return err
		}

		credential := kubernetes.NewExecCredential(
			eksToken.Status.Token,
			eksToken.Status.ExpirationTimestamp,
		)

		if err := credential.PrintJSON(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: Failed to output credential: %v\n", err)
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getTokenCmd)
	getTokenCmd.Flags().StringVar(&clusterName, "cluster-name", "", "EKS cluster name (required)")
	getTokenCmd.Flags().StringVar(&region, "region", "", "AWS region (required)")
}
