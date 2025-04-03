package dnsimple

import (
	"context"
	"fmt"
	"strconv"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dissolve/dnssd/advertiser/dnsimple/internal/dnsimplex"
)

func (a *Advertiser) findSRV(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
) (dnsimple.ZoneRecord, bool, error) {
	return dnsimplex.One(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.ZoneRecord, error) {
			res, err := a.Client.Zones.ListRecords(
				ctx,
				strconv.FormatInt(zone.AccountID, 10),
				zone.Name,
				&dnsimple.ZoneRecordListOptions{
					ListOptions: opts,
					Name: dnsimple.String(
						dnssd.EscapeInstance(inst.Name) + "." + inst.ServiceType,
					),
					Type: dnsimple.String("SRV"),
				},
			)
			if err != nil {
				return nil, nil, dnsimplex.Errorf("unable to list SRV records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
	)
}

func (a *Advertiser) syncSRV(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findSRV(ctx, zone, inst)
	if err != nil {
		return err
	}

	desired := dnsimple.ZoneRecordAttributes{
		ZoneID: zone.Name,
		Type:   "SRV",
		Name: dnsimple.String(
			dnssd.EscapeInstance(inst.Name) + "." + inst.ServiceType,
		),
		Content: fmt.Sprintf(
			"%d %d %s",
			inst.Weight,
			inst.TargetPort,
			inst.TargetHost,
		),
		TTL:      int(inst.TTL.Seconds()),
		Priority: int(inst.Priority),
	}

	if ok {
		cs.Update(current, desired)
	} else {
		cs.Create(desired)
	}

	return nil
}

func (a *Advertiser) deleteSRV(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, ok, err := a.findSRV(ctx, zone, inst)
	if !ok || err != nil {
		return err
	}

	cs.Delete(current)

	return nil
}
