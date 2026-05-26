package main

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v7/dns"
	"github.com/cloudflare/cloudflare-go/v7/zones"
	"github.com/manifoldco/promptui"
)

func selectFromSlice[T any](slice []T, label string, name func(T) string) (*T, error) {
	nameEntries := make([]string, len(slice))
	for i, val := range slice {
		nameEntries[i] = name(val)
	}

	prompt := promptui.Select{
		Label: label,
		Items: nameEntries,
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return nil, err
	}
	return &slice[idx], nil
}

func getNewEntry(ctx context.Context) (*DynamicDnsEntry, error) {
	zoneList, err := getZones(ctx)
	if err != nil {
		return nil, err
	}

	zone, err := selectFromSlice(zoneList, "select zone", func(zone zones.Zone) string {
		return zone.Name
	})
	if err != nil {
		return nil, err
	}

	recordList, err := getRecords(ctx, zone.ID)
	if err != nil {
		return nil, err
	}

	record, err := selectFromSlice(recordList, "select record", func(rec dns.RecordResponse) string {
		return rec.Name
	})
	if err != nil {
		return nil, err
	}

	return &DynamicDnsEntry{
		ZoneId:   zone.ID,
		RecordId: record.ID,
		Name:     record.Name,
	}, nil
}
