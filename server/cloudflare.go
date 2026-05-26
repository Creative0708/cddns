package main

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/zones"
)

var globalClient *cloudflare.Client = nil
var globalClientLock sync.Once = sync.Once{}

func getClient() *cloudflare.Client {
	globalClientLock.Do(func() {
		globalClient = cloudflare.NewClient()
	})
	return globalClient
}

func updateServerIp(ctx context.Context, zoneId string, recordId string, address net.IP) {
	var body dns.RecordEditParamsBodyUnion
	if ip := address.To4(); ip != nil {
		body = dns.ARecordParam{
			Content: cloudflare.F(ip.String()),
		}
	} else if ip := address.To16(); ip != nil {
		body = dns.AAAARecordParam{
			Content: cloudflare.F(ip.String()),
		}
	}
	client := getClient()
	_, err := client.DNS.Records.Edit(ctx, recordId, dns.RecordEditParams{
		ZoneID: cloudflare.F(zoneId),
		Body:   body,
	})
	if err != nil {
		log.Println("error setting dns record: ", err)
	}
}
func getZones(ctx context.Context) ([]zones.Zone, error) {
	client := getClient()
	iter := client.Zones.ListAutoPaging(ctx, zones.ZoneListParams{})

	slice := make([]zones.Zone, 0)
	for iter.Next() {
		slice = append(slice, iter.Current())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return slice, nil
}
func getRecords(ctx context.Context, zoneId string) ([]dns.RecordResponse, error) {
	client := getClient()
	iter := client.DNS.Records.ListAutoPaging(ctx, dns.RecordListParams{
		ZoneID: cloudflare.F(zoneId),
	})

	slice := make([]dns.RecordResponse, 0)
	for iter.Next() {
		slice = append(slice, iter.Current())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return slice, nil
}
