package main

import (
	"log/slog"
	"os"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/murilopereirame/dyndns/internal/config"
	"github.com/murilopereirame/dyndns/internal/runner"
	"github.com/murilopereirame/dyndns/pkg/client"
	"github.com/murilopereirame/dyndns/pkg/ddns"
	zone_manager "github.com/murilopereirame/dyndns/pkg/dns"
	"github.com/murilopereirame/dyndns/pkg/nic"
	"github.com/murilopereirame/dyndns/pkg/notification"
	"github.com/murilopereirame/dyndns/pkg/tplink"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	configFile, configExists := os.LookupEnv("CONFIG_PATH")
	if !configExists {
		configFile = "config.json"
	}

	loadedConfig, err := config.NewConfigHandler(configFile).Load()

	if err != nil {
		slog.Error("Error while loading config file", "file", configFile)
		panic("Failed to load config file")
	}

	var authRouter *tplink.AuthTPLinkRouter

	if loadedConfig.Router.IPv4 || loadedConfig.Router.IPv6 || loadedConfig.Firewall.Enabled {
		httpClient := client.Client{
			BaseUrl: loadedConfig.Router.Endpoint,
		}

		router, err := tplink.NewRouter(
			loadedConfig.Router.Endpoint,
			loadedConfig.Router.User,
			loadedConfig.Router.Password,
			httpClient,
		)

		if err != nil {
			slog.Error("Failed to create Router", "error", err.Error())
			panic("Failed to create Router")
		}

		authRouter, err = router.Authenticate()

		if err != nil {
			slog.Error("Failed to authenticate in router", "error", err.Error())
			panic("Failed to authenticate in router")
		}
	}

	zManager := zone_manager.ZoneManager{
		ZoneId:  loadedConfig.Domain.ZoneId,
		Domain:  loadedConfig.Domain.Domain,
		Proxied: loadedConfig.Domain.Proxied,
		Client:  cloudflare.NewClient(option.WithAPIKey(loadedConfig.Domain.Token)),
	}

	nic := nic.NIC{
		Interface: loadedConfig.NIC.Interface,
	}

	ddns := ddns.NewDDNS(&zManager, authRouter, &nic, loadedConfig)

	notifier := notification.NotificationSender{
		Url:        loadedConfig.Webhook.Url,
		Secret:     loadedConfig.Webhook.Secret,
		AuthHeader: loadedConfig.Webhook.AuthHeader,
	}

	runner := runner.NewRunner(*ddns, notifier, loadedConfig.Interval)

	runner.Run()
}
