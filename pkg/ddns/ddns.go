package ddns

import (
	"log/slog"
	"time"

	"github.com/murilopereirame/dyndns/internal/config"
	zone_manager "github.com/murilopereirame/dyndns/pkg/dns"
	"github.com/murilopereirame/dyndns/pkg/nic"
	"github.com/murilopereirame/dyndns/pkg/tplink"
)

type RemoteIPValues struct {
	IPv4 string
	IPv6 string
}

type DDNSState struct {
	LastCheck    int64
	RemoteValues RemoteIPValues
}

type DDNS struct {
	ZoneManager *zone_manager.ZoneManager
	Router      *tplink.AuthTPLinkRouter
	NIC         *nic.NIC
	Config      *config.Config
	State       DDNSState
}

func NewDDNS(
	manager *zone_manager.ZoneManager,
	router *tplink.AuthTPLinkRouter,
	nic *nic.NIC,
	config *config.Config,
) *DDNS {
	state := fetchCurrentState(manager)

	return &DDNS{
		ZoneManager: manager,
		Router:      router,
		NIC:         nic,
		Config:      config,
		State:       state,
	}
}

func fetchCurrentState(manager *zone_manager.ZoneManager) DDNSState {
	ip4records, err := manager.FetchIPv4Records()

	ipv4 := ""
	if err == nil {
		ipv4 = ip4records.Result[0].Content
	} else {
		slog.Error("Failed to load IPv4 record from remote", "error", err)
	}

	ip6records, err := manager.FetchIPv6Records()

	ipv6 := ""
	if err == nil {
		ipv6 = ip6records.Result[0].Comment
	} else {
		slog.Error("Failed to load IPv6 record from remote", "error", err)
	}

	return DDNSState{
		LastCheck: time.Now().Unix(),
		RemoteValues: RemoteIPValues{
			IPv4: ipv4,
			IPv6: ipv6,
		},
	}
}

// Returns a boolean if any remote record has been updated
// together with the most recent DNS State
func (d *DDNS) CheckEntries() (bool, DDNSState) {
	ip4, err := d.updateIPv4()

	if err != nil {
		slog.Error("Error while updating IPv4", "error", err)
		ip4 = d.State.RemoteValues.IPv4
	}

	ip6, err := d.updateIPv6()

	if err != nil {
		slog.Error("Error while updating IPv6", "error", err)
		ip6 = d.State.RemoteValues.IPv6
	}

	newState := &DDNSState{
		LastCheck: time.Now().Unix(),
		RemoteValues: RemoteIPValues{
			IPv4: ip4,
			IPv6: ip6,
		},
	}

	slog.Info("Entries checked", "old", d.State, "new", newState)

	hasIP6Changed := d.State.RemoteValues.IPv6 != newState.RemoteValues.IPv6
	hasIP4Changed := d.State.RemoteValues.IPv4 != newState.RemoteValues.IPv4

	if hasIP6Changed && d.Config.Firewall.Enabled {
		_, error := d.checkIPv6Firewall(ip6)

		if error != nil {
			slog.Error("Failed to update Firewall rules", "error", error)
		}
	}

	hasChanges := hasIP4Changed || hasIP6Changed

	d.State = *newState

	return hasChanges, *newState
}

func (d *DDNS) updateIPv6() (string, error) {
	if !d.Config.Router.IPv6 && !d.Config.NIC.IPv6 {
		return d.State.RemoteValues.IPv6, nil
	}

	var ipv6 string
	var err error

	if d.Config.Router.IPv6 {
		ipv6, err = d.Router.GetIPv6Address()
	} else if d.Config.NIC.IPv6 {
		ipv6, err = d.NIC.GetIPv4Address()
	}

	if err != nil {
		slog.Error("Failed to fetch IPv6 address", "error", err)
		return d.State.RemoteValues.IPv6, err
	}

	if ipv6 == "" || d.State.RemoteValues.IPv4 == "" || ipv6 == d.State.RemoteValues.IPv4 {
		return d.State.RemoteValues.IPv6, nil
	}

	err = d.ZoneManager.UpdateDNS6Record(ipv6)

	if err != nil {
		slog.Error("Failed to update IPv6 DNS record", "error", err)
		return d.State.RemoteValues.IPv6, err
	}

	return ipv6, nil
}

func (d *DDNS) updateIPv4() (string, error) {
	if !d.Config.Router.IPv4 && !d.Config.NIC.IPv4 {
		return d.State.RemoteValues.IPv4, nil
	}

	var ipv4 string
	var err error

	if d.Config.Router.IPv4 {
		ipv4, err = d.Router.GetIPv4Address()
	} else if d.Config.NIC.IPv4 {
		ipv4, err = d.NIC.GetIPv4Address()
	}

	if err != nil {
		slog.Error("Failed to fetch IPv4 address", "error", err)
		return d.State.RemoteValues.IPv4, err
	}

	if ipv4 == "" || d.State.RemoteValues.IPv4 == "" || ipv4 == d.State.RemoteValues.IPv4 {
		return d.State.RemoteValues.IPv4, nil
	}

	err = d.ZoneManager.UpdateDNSRecord(ipv4)

	if err != nil {
		slog.Error("Failed to update IPv4 DNS record", "error", err)
		return d.State.RemoteValues.IPv4, err
	}

	return ipv4, nil
}

func (d *DDNS) checkIPv6Firewall(ip string) (bool, error) {
	if d.Router == nil {
		slog.Info("Router not initialized, skipping firewall checks")
		return false, nil
	}

	newRule := &tplink.IPv6FirewallRule{
		Port:     d.Config.Firewall.Port,
		Name:     d.Config.Firewall.Name,
		Enable:   "on",
		Protocol: d.Config.Firewall.Protocol,
		IP:       ip,
		Type:     "CUSTOM",
	}

	return d.Router.AddPortInIPv6Firewall(*newRule, d.Config.Firewall.OnConflict)
}
