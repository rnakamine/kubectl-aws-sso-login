# kubectl-aws-sso-login

This is a client-go credential (exec) plugin that automates AWS SSO authentication for EKS clusters.

## Installation

```bash
go install github.com/rnakamine/kubectl-aws-sso-login@latest
```

Or download the binary from the [releases page](https://github.com/rnakamine/kubectl-aws-sso-login/releases).

## Usage

### Setup

Configure your kubeconfig to use this credential plugin:

1. Run aws eks update-kubeconfig first, then modify the generated configuration:

```bash
aws eks update-kubeconfig --name <your-cluster-name> --region <your-region> --profile <your-profile>
```

2. Edit `~/.kube/config` manually and change the exec command:

```yaml
users:
- name: arn:aws:eks:<your-region>:<your-account>:cluster/<your-cluster-name>
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: kubectl-aws-sso-login
      args:
      - get-token
      - --cluster-name
      - <your-cluster-name>
      - --region
      - <your-region>
      env:
      - name: AWS_PROFILE
        value: <your-profile>
```

### Running

```bash
kubectl get pods
```

When kubectl needs to authenticate, it will call this plugin which automatically:

- Checks if your AWS SSO session is valid
- Prompts for SSO login if needed
- Returns the authentication token to kubectl

## Requirements

- AWS CLI v2
- kubectl
- AWS SSO profile configured (`aws configure sso`)

## License

MIT
