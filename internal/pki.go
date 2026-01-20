package internal

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

type Meta struct {
	Expiration int64 `json:"expiration"`
}

func writeFile(dir, name, content string, perm os.FileMode) {
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		log.Printf("[%s] Error writing file %s: %v", dir, name, err)
	}
}

func IssueCert(c *api.Client, cfg Config) (bool, error) {
	data := map[string]interface{}{
		"common_name": cfg.PKI.CommonName,
	}

	if len(cfg.PKI.AltNames) > 0 {
		data["alt_names"] = strings.Join(cfg.PKI.AltNames, ",")
	}

	if cfg.PKI.TTL != "" {
		data["ttl"] = cfg.PKI.TTL
	}

	sec, err := c.Logical().Write(cfg.PKI.Path+"/issue/"+cfg.PKI.Role, data)
	if err != nil {
		return false, err
	}

	if err := os.MkdirAll(cfg.Output.Dir, 0755); err != nil {
		return false, err
	}

	serverCert := sec.Data["certificate"].(string)
	privateKey := sec.Data["private_key"].(string)

	caChain := ""
	if chain, ok := sec.Data["ca_chain"].([]interface{}); ok {
		for _, c := range chain {
			if s, ok := c.(string); ok {
				caChain += s + "\n"
			}
		}
	} else if issuingCA, ok := sec.Data["issuing_ca"].(string); ok {
		caChain = issuingCA
	}

	writeFile(cfg.Output.Dir, "cert.pem", serverCert, 0644)
	writeFile(cfg.Output.Dir, "key.pem", privateKey, 0600)
	writeFile(cfg.Output.Dir, "ca.pem", caChain, 0644)
	writeFile(cfg.Output.Dir, "fullchain.pem", serverCert+"\n"+caChain, 0644)

	// --- expiration ---
	var expInt int64
	switch v := sec.Data["expiration"].(type) {
	case int64:
		expInt = v
	case float64:
		expInt = int64(v)
	case json.Number:
		expInt, _ = v.Int64()
	default:
		expInt = time.Now().Add(24 * time.Hour).Unix()
	}

	meta := Meta{Expiration: expInt}
	metaBytes, _ := json.Marshal(meta)
	writeFile(cfg.Output.Dir, "meta.json", string(metaBytes), 0644)

	log.Printf("[%s] Certificate issued, expires at: %s", cfg.Name, time.Unix(expInt, 0))
	return true, nil
}

func NeedsRenew(cfg Config) bool {
	metaPath := filepath.Join(cfg.Output.Dir, "meta.json")
	b, err := os.ReadFile(metaPath)
	if err != nil {
		log.Printf("[%s] meta.json not found, certificate needs renewal", cfg.Name)
		return true
	}

	var m Meta
	if err := json.Unmarshal(b, &m); err != nil {
		log.Printf("[%s] Failed to parse meta.json, renewing certificate", cfg.Name)
		return true
	}

	renewBefore, err := time.ParseDuration(cfg.Daemon.RenewBefore)
	if err != nil {
		renewBefore = 72 * time.Hour
	}

	need := time.Until(time.Unix(m.Expiration, 0)) < renewBefore
	if need {
		log.Printf("[%s] Certificate expires soon, needs renewal", cfg.Name)
	}
	return need
}
