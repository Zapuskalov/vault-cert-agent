package internal

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

func VaultClient(cfg Config) (*api.Client, error) {
	vaultCfg := api.DefaultConfig()
	vaultCfg.Address = cfg.Vault.Addr

	// TLS auth (mTLS)
	if cfg.Vault.Auth.Method == "tls" {
		if err := vaultCfg.ConfigureTLS(&api.TLSConfig{
			ClientCert: cfg.Vault.Auth.TLS.Cert,
			ClientKey:  cfg.Vault.Auth.TLS.Key,
			CACert:     cfg.Vault.Auth.TLS.CA,
		}); err != nil {
			return nil, fmt.Errorf("[%s] failed to configure TLS: %w", cfg.Name, err)
		}
	}

	client, err := api.NewClient(vaultCfg)
	if err != nil {
		return nil, fmt.Errorf("[%s] failed to create Vault client: %w", cfg.Name, err)
	}

	switch cfg.Vault.Auth.Method {
	case "token":
		client.SetToken(cfg.Vault.Auth.Token)

	case "approle":
		sec, err := client.Logical().Write("auth/approle/login", map[string]interface{}{
			"role_id":   cfg.Vault.Auth.RoleID,
			"secret_id": cfg.Vault.Auth.SecretID,
		})
		if err != nil {
			return nil, fmt.Errorf("[%s] approle login failed: %w", cfg.Name, err)
		}
		if sec.Auth == nil {
			return nil, fmt.Errorf("[%s] approle login returned nil auth", cfg.Name)
		}
		client.SetToken(sec.Auth.ClientToken)

	case "tls":

	default:
		return nil, fmt.Errorf("[%s] unknown auth method: %s", cfg.Name, cfg.Vault.Auth.Method)
	}

	return client, nil
}
