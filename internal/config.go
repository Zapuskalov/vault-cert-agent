package internal

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type TLSConfig struct {
	Cert string `mapstructure:"cert"`
	Key  string `mapstructure:"key"`
	CA   string `mapstructure:"ca"`
}

type VaultConfig struct {
	Addr string `mapstructure:"addr"`
	Auth struct {
		Method   string    `mapstructure:"method"`
		Token    string    `mapstructure:"token"`
		RoleID   string    `mapstructure:"role_id"`
		SecretID string    `mapstructure:"secret_id"`
		TLS      TLSConfig `mapstructure:"tls"`
	} `mapstructure:"auth"`
}

type PKIConfig struct {
	Path       string   `mapstructure:"path"`
	Role       string   `mapstructure:"role"`
	CommonName string   `mapstructure:"common_name"`
	AltNames   []string `mapstructure:"alt_names"`
	TTL        string   `mapstructure:"ttl"`
}

type OutputConfig struct {
	Dir string `mapstructure:"dir"`
}

type DaemonConfig struct {
	CheckInterval string `mapstructure:"check_interval"`
	RenewBefore   string `mapstructure:"renew_before"`
}

type HooksConfig struct {
	PostIssue []string `mapstructure:"post_issue"`
}

type Config struct {
	Name   string       `mapstructure:"-"`
	Vault  VaultConfig  `mapstructure:"vault"`
	PKI    PKIConfig    `mapstructure:"pki"`
	Output OutputConfig `mapstructure:"output"`
	Daemon DaemonConfig `mapstructure:"daemon"`
	Hooks  HooksConfig  `mapstructure:"hooks"`
}

var Cfgs []Config

func InitConfig(cfgFile *string, cfgDir *string) {
	if cfgFile != nil && *cfgFile != "" && cfgDir != nil && *cfgDir != "" {
		log.Fatal("Only one of --config or --config-dir can be specified")
	}

	switch {
	case cfgFile != nil && *cfgFile != "":
		cfg, err := loadConfigFile(*cfgFile)
		if err != nil {
			log.Fatal(err)
		}
		Cfgs = append(Cfgs, cfg)

	case cfgDir != nil && *cfgDir != "":
		if err := loadConfigDir(*cfgDir); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatal("Either --config or --config-dir must be specified")
	}

	if len(Cfgs) == 0 {
		log.Fatal("No valid configuration files loaded")
	}

	log.Printf("Loaded %d configuration(s)", len(Cfgs))
}

func loadConfigDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("cannot read config dir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			continue
		}

		path := filepath.Join(dir, e.Name())
		cfg, err := loadConfigFile(path)
		if err != nil {
			return err
		}
		Cfgs = append(Cfgs, cfg)
	}

	return nil
}

func loadConfigFile(path string) (Config, error) {
	var cfg Config

	v := viper.New()
	v.SetConfigFile(path)

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	v.BindEnv("vault.addr", "VAULT_ADDR")
	v.BindEnv("vault.auth.method", "VAULT_AUTH_METHOD")
	v.BindEnv("vault.auth.token", "VAULT_AUTH_TOKEN")
	v.BindEnv("vault.auth.role_id", "VAULT_AUTH_ROLE_ID")
	v.BindEnv("vault.auth.secret_id", "VAULT_AUTH_SECRET_ID")
	v.BindEnv("vault.auth.tls.cert", "VAULT_AUTH_TLS_CERT")
	v.BindEnv("vault.auth.tls.key", "VAULT_AUTH_TLS_KEY")
	v.BindEnv("vault.auth.tls.ca", "VAULT_AUTH_TLS_CA")

	if err := v.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("failed to decode config %s: %w", path, err)
	}

	cfg.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	log.Println("Loaded config:", path)
	return cfg, nil
}
