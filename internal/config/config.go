package config

import (
	"encoding/json"
	"os"

	"github.com/murilopereirame/dyndns/pkg/notification"
	"github.com/murilopereirame/dyndns/pkg/tplink"
)

type RouterConfig struct {
	Endpoint string
	User     string
	Password string
	IPv4     bool
	IPv6     bool
}

type WebhookConfig struct {
	Url        string
	Secret     string
	AuthHeader string
	Priority   notification.NotificationPriority
	Enabled    bool
}

type DomainConfig struct {
	Domain  string
	Token   string
	ZoneId  string
	Proxied bool
}

type NICConfig struct {
	Interface string
	IPv4      bool
	IPv6      bool
}

type FirewallConfig struct {
	Port       int
	Name       string
	Protocol   tplink.NetworkProtocol
	OnConflict tplink.RuleConflictResolution
	Enabled    bool
}

type Config struct {
	Interval int64
	Webhook  WebhookConfig
	Router   RouterConfig
	Domain   DomainConfig
	NIC      NICConfig
	Firewall FirewallConfig
}

type ConfigHandler struct {
	configPath string
}

func NewConfigHandler(configPath string) *ConfigHandler {
	return &ConfigHandler{
		configPath: configPath,
	}
}

func (c *ConfigHandler) Load() (*Config, error) {
	config := &Config{}

	content, readError := os.ReadFile(c.configPath)

	if readError != nil {
		return nil, readError
	}

	err := json.Unmarshal(content, &config)

	if err != nil {
		return nil, err
	}

	return config, nil
}
