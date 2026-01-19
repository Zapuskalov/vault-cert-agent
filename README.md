# Vault Cert Agent

Vault Cert Agent is a tool for issuing and automatically renewing certificates from HashiCorp Vault. It supports writing certificates, private keys, and CA chains to disk, running hooks after certificate issuance, and authenticating via Token, AppRole, or mTLS.

---

## Features

- Issue new certificates from Vault PKI
- Auto-renew certificates before expiration
- Write certificates, private keys, and CA chains to disk
- Run custom hooks (e.g., reload Nginx/Haproxy) after issuing a certificate
- Support authentication via Token, AppRole, or mTLS
- Support multiple configurations via a config directory

---

## Installation

### Build from source

```bash
# Build the binary in the current directory
go build -o vault-cert-agent

# Or install to $GOPATH/bin or $HOME/go/bin
go install
```

Make sure \$HOME/go/bin is in your PATH if you use `go install`:

```bash
export PATH=\$PATH:\$HOME/go/bin
```

---

## Usage

```bash
# Issue a certificate once using a single config file
vault-cert-agent issue --config ./config.yaml

# Run as a daemon to auto-renew certificates
vault-cert-agent daemon --config ./config.yaml

# Support multiple configurations using a directory
vault-cert-agent daemon --config-dir /etc/vault-cert-agent/

# Generate shell completion for bash
vault-cert-agent completion bash > /etc/bash_completion.d/vault-cert-agent
```

---

## Configuration

Configuration can be provided via:

- YAML or JSON config file (`--config` or `--config-dir`)
- Environment variables for Vault connection and auth (`VAULT_*`)
- Defaults

**Example config (YAML):**

```yaml
vault:
  addr: https://pki.example.com
  auth:
    method: token      # "token", "approle", "tls"
    token: "<vault-token>"   # required if method is token
#    role_id: "<role-id>"     # required if method is approle
#    secret_id: "<secret-id>" # required if method is approle
#    tls:
#      cert: /path/to/client.crt   # required if method is tls
#      key: /path/to/client.key
#      ca: /path/to/ca.crt

pki:
  path: intermediate
  role: my-role
  common_name: myhost.example.com
  alt_names: ["alt1.example.com","alt2.example.com"]
  ttl: 30d

output:
  dir: /path/to/output/dir

daemon:
  check_interval: 30m
  renew_before: 48h

hooks:
  post_issue:
    - systemctl reload nginx
```

---

## Environment Variables

Only Vault connection/auth can be configured via ENV:

- `VAULT_ADDR`          → Vault server address
- `VAULT_AUTH_METHOD`   → "token", "approle", or "tls"
- `VAULT_AUTH_TOKEN`    → Token for "token" auth
- `VAULT_AUTH_ROLE_ID`  → RoleID for "approle" auth
- `VAULT_AUTH_SECRET_ID`→ SecretID for "approle" auth
- `VAULT_AUTH_TLS_CERT` → Path to client certificate for mTLS
- `VAULT_AUTH_TLS_KEY`  → Path to private key for mTLS
- `VAULT_AUTH_TLS_CA`   → Path to CA certificate for mTLS

---

## Running as a systemd Service

You can run `vault-cert-agent` as a daemon managed by systemd. This allows the agent to automatically start on boot and restart on failure.

Create a systemd unit file `/etc/systemd/system/vault-cert-agent.service`:

```ini
[Unit]
Description=Vault PKI Certificate Agent
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=vault-agent
Group=vault-agent
ExecStart=/usr/local/bin/vault-cert-agent daemon --config-dir /etc/vault-cert-agent/
Restart=on-failure
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

**Notes:**
- Replace `/usr/local/bin/vault-cert-agent` with the actual path to the binary if installed elsewhere.
- Ensure the `vault-agent` user exists and has access to the output directories and configuration files.
- Environment variables set in the unit file (e.g., `VAULT_AUTH_TOKEN`) will override any defaults and are only used for Vault authentication.

### Managing the service:

```bash
# Reload systemd to recognize the new service
sudo systemctl daemon-reload

# Start the service
sudo systemctl start vault-cert-agent

# Enable autostart on boot
sudo systemctl enable vault-cert-agent

# Check status
sudo systemctl status vault-cert-agent

# Follow logs
sudo journalctl -u vault-cert-agent -f
```
