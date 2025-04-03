package dnsimple

import (
	"context"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dissolve/dnssd/advertiser/dnsimple/internal/dnsimplex"
)

func (a *Advertiser) findPTR(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
) (dnsimple.ZoneRecord, bool, error) {
	return dnsimplex.First(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := a.Client.Zones.ListRecords(
				ctx,
				strconv.FormatInt(zone.AccountID, 10),
				zone.Name,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name:        dnsimple.String(inst.ServiceType),
					Type:        dnsimple.String("PTR"),
				},
			)
			if err != nil {
				return nil, nil, dnsimplex.Errorf("unable to list PTR records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
		func(candidate dnsimple.ZoneRecord) bool {
			return candidate.Content == strings.TrimRight(inst.Absolute(), ".")
		},
	)
}

func (a *Advertiser) syncPTR(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findPTR(ctx, zone, inst)
	if err != nil {
		return err
	}

	desired := dnsimple.ZoneRecordAttributes{
		ZoneID:  zone.Name,
		Type:    "PTR",
		Name:    dnsimple.String(inst.ServiceType),
		Content: strings.TrimRight(inst.Absolute(), "."),
		TTL:     int(inst.TTL.Seconds()),
	}

	if ok {
		cs.Update(current, desired)
	} else {
		cs.Create(desired)
	}

	return nil
}

func (a *Advertiser) deletePTR(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findPTR(ctx, zone, inst)
	if !ok || err != nil {
		return err
	}

	cs.Delete(current)

	return nil
}
