package cmd

import (
	"fmt"
	"os"

	"vault-cert-agent/internal"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	cfgDir  string
)

var rootCmd = &cobra.Command{
	Use:   "vault-cert-agent",
	Short: "Vault PKI certificate management agent",
	Long: `Vault PKI agent for issuing and automatically renewing certificates from HashiCorp Vault.

This tool allows you to:
  • Issue new certificates from Vault PKI
  • Auto-renew certificates before expiration
  • Write certificates, private keys, and CA chains to disk
  • Run custom hooks (e.g., reload nginx/haproxy) after issuing a certificate
  • Support authentication via Token, AppRole, or mTLS
  • Handle multiple configuration files via --config-dir (each config is processed separately)

Configuration can be provided via:
  • YAML or JSON config file (--config)
  • Directory with multiple configs (--config-dir)
  • Environment variables (VAULT_* for Vault connection/auth)
  • Defaults

Supported configuration fields (example YAML):

vault:
  addr: https://pki.example.com
  auth:
    method: token      # "token", "approle", "tls"
    token: "<vault-token>"   # required if method is token
    role_id: "<role-id>"     # required if method is approle
    secret_id: "<secret-id>" # required if method is approle
    tls:
      cert: /path/to/client.crt   # required if method is tls
      key: /path/to/client.key
      ca: /path/to/ca.crt

pki:
  path: intermediate
  role: my-role
  common_name: myhost.example.com
  alt_names: ["alt1.example.com","alt2.example.com"] # optional
  ttl: 30d  # optional, default Vault TTL

output:
  dir: /path/to/output/dir

daemon:
  check_interval: 30m
  renew_before: 48h

hooks:
  post_issue:
    - systemctl reload nginx

Supported environment variables (for Vault connection and authentication):

  VAULT_ADDR          → Vault server address
  VAULT_AUTH_METHOD   → "token", "approle", or "tls"
  VAULT_AUTH_TOKEN    → Token for "token" auth
  VAULT_AUTH_ROLE_ID  → RoleID for "approle" auth
  VAULT_AUTH_SECRET_ID→ SecretID for "approle" auth
  VAULT_AUTH_TLS_CERT → Path to client certificate for mTLS
  VAULT_AUTH_TLS_KEY  → Path to private key for mTLS
  VAULT_AUTH_TLS_CA   → Path to CA certificate for mTLS

Notes:
  • Only Vault connection and authentication can be configured via ENV.
  • PKI path, role, CN, output directory, TTL, and hooks must be set in the config file(s).
  • If using --config-dir, each YAML/JSON file is treated as a separate configuration.
`,
	Example: `  # Issue a certificate once using a single config file
  vault-cert-agent issue --config ./config.yaml

  # Run as a daemon that checks and auto-renews certificates every interval
  vault-cert-agent daemon --config ./config.yaml

  # Use multiple configs from a directory
  vault-cert-agent daemon --config-dir ./configs

  # Generate shell completion for bash
  vault-cert-agent completion bash > /etc/bash_completion.d/vault-cert-agent
`,
}

var issueCmd = &cobra.Command{
	Use:   "issue",
	Short: "Issue certificate(s) once",
	Long:  "Issue certificate(s) once for all loaded configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, cfg := range internal.Cfgs {
			if err := internal.RunOnce(cfg); err != nil {
				return err
			}
		}
		return nil
	},
}

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run as daemon and auto-renew certificates",
	Long:  "Run a supervisor that manages certificate workers for all loaded configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		return internal.RunDaemon()
	},
}

func Execute() {
	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"Path to config file (yaml or json)",
	)

	rootCmd.PersistentFlags().StringVar(
		&cfgDir,
		"config-dir",
		"",
		"Path to directory with config files (yaml/json)",
	)

	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(daemonCmd)

	cobra.OnInitialize(func() {
		internal.InitConfig(&cfgFile, &cfgDir)
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
