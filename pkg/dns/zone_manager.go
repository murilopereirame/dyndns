package zone_manager

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/packages/pagination"
)

type DNSRecord string

const (
	RecordTypeA    DNSRecord = DNSRecord(dns.RecordEditParamsBodyTypeA)
	RecordTypeAAAA DNSRecord = DNSRecord(dns.RecordEditParamsBodyTypeAAAA)
)

type ZoneManager struct {
	ZoneId  string
	Domain  string
	Proxied bool
	Client  *cloudflare.Client
}

func (z *ZoneManager) UpdateDNSRecord(ip string) error {
	ipv4Entry, err := z.FetchIPv4Records()

	if err != nil {
		return err
	}

	if len(ipv4Entry.Result) == 0 {
		return z.createRecord(ip, RecordTypeA)
	}

	recordID := ipv4Entry.Result[0].ID
	return z.updateRecord(recordID, ip, RecordTypeA)
}

func (z *ZoneManager) UpdateDNS6Record(ip string) error {
	ipv6Entry, err := z.Client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F(z.ZoneId),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(z.Domain),
		}),
		Type: cloudflare.F(dns.RecordListParamsTypeAAAA),
	})

	if err != nil {
		return err
	}

	if len(ipv6Entry.Result) == 0 {
		return z.createRecord(ip, RecordTypeAAAA)
	}

	recordID := ipv6Entry.Result[0].ID
	return z.updateRecord(recordID, ip, RecordTypeAAAA)
}

func (z *ZoneManager) FetchIPv4Records() (*pagination.V4PagePaginationArray[dns.RecordResponse], error) {
	return z.Client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F(z.ZoneId),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(z.Domain),
		}),
		Type: cloudflare.F(dns.RecordListParamsTypeA),
	})
}

func (z *ZoneManager) FetchIPv6Records() (*pagination.V4PagePaginationArray[dns.RecordResponse], error) {
	return z.Client.DNS.Records.List(context.TODO(), dns.RecordListParams{
		ZoneID: cloudflare.F(z.ZoneId),
		Name: cloudflare.F(dns.RecordListParamsName{
			Exact: cloudflare.F(z.Domain),
		}),
		Type: cloudflare.F(dns.RecordListParamsTypeAAAA),
	})
}

func (z *ZoneManager) createRecord(ip string, entryType DNSRecord) error {
	_, err := z.Client.DNS.Records.New(context.TODO(), dns.RecordNewParams{
		ZoneID: cloudflare.F(z.ZoneId),
		Body: dns.RecordNewParamsBody{
			Name:    cloudflare.F(z.Domain),
			TTL:     cloudflare.F(dns.TTL1),
			Type:    cloudflare.F(dns.RecordNewParamsBodyType(entryType)),
			Content: cloudflare.F(ip),
			Proxied: cloudflare.F(z.Proxied),
			Comment: cloudflare.F("Updated using murilopereirame/dyndns"),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (z *ZoneManager) updateRecord(recordID string, ip string, recordType DNSRecord) error {
	_, err := z.Client.DNS.Records.Update(context.TODO(), recordID, dns.RecordUpdateParams{
		ZoneID: cloudflare.F(z.ZoneId),
		Body: dns.RecordUpdateParamsBody{
			Name:    cloudflare.F(z.Domain),
			TTL:     cloudflare.F(dns.TTL1),
			Content: cloudflare.F(ip),
			Proxied: cloudflare.F(z.Proxied),
			Comment: cloudflare.F("Updated using murilopereirame/dyndns"),
			Type:    cloudflare.F(dns.RecordUpdateParamsBodyType(recordType)),
		},
	})

	if err != nil {
		return err
	}

	return nil
}
