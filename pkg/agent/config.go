package agent

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type AgentConfig struct {
	Central CentralConfig `yaml:"central"`
	Agent   AgentMeta     `yaml:"agent"`
	Targets []TargetRef   `yaml:"targets"`
	TLS     TLSConfig     `yaml:"tls"`
	Vault   VaultConfig   `yaml:"vault"`
	Sinks   SinksConfig   `yaml:"sinks"`
}

type CentralConfig struct {
	URL           string `yaml:"url"`
	CAFingerprint string `yaml:"ca_fingerprint"`
}

type AgentMeta struct {
	ID       string `yaml:"id"`
	HostPort int    `yaml:"host_port"`
}

type TargetRef struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	VaultPath string `yaml:"vault_path"`
}

type TLSConfig struct {
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	CAFile   string `yaml:"ca_file"`
}

type VaultConfig struct {
	Address      string `yaml:"address"`
	AuthMethod   string `yaml:"auth_method"`
	RoleIDEnv    string `yaml:"role_id_env"`
	SecretIDFile string `yaml:"secret_id_file"`
}

type SinksConfig struct {
	VictoriaMetrics *VMSinkConfig    `yaml:"victoria_metrics,omitempty"`
	ConfigStore     *ConfigStoreSink `yaml:"config_store,omitempty"`
}

type VMSinkConfig struct {
	URL string `yaml:"url"`
}

type ConfigStoreSink struct {
	URL string `yaml:"url"`
}

func LoadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg AgentConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *AgentConfig) validate() error {
	if c.Central.URL == "" {
		return fmt.Errorf("central.url is required")
	}
	if err := requireTLS(c.Central.URL, "central.url"); err != nil {
		return err
	}
	if c.Sinks.VictoriaMetrics != nil {
		if err := requireTLS(c.Sinks.VictoriaMetrics.URL, "sinks.victoria_metrics.url"); err != nil {
			return err
		}
	}
	if c.Vault.Address != "" {
		if err := requireTLS(c.Vault.Address, "vault.address"); err != nil {
			return err
		}
	}
	if c.Agent.ID == "" {
		return fmt.Errorf("agent.id is required")
	}
	if c.Agent.HostPort == 0 {
		c.Agent.HostPort = 9100
	}
	return nil
}

func requireTLS(rawURL, field string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%s: invalid URL: %w", field, err)
	}
	if strings.EqualFold(u.Scheme, "http") {
		return fmt.Errorf("FATAL: plaintext communication not supported for %s. All endpoints must use TLS. See https://docs.dbx.io/security/tls", field)
	}
	return nil
}
