package dnsimple

import (
	"context"
	"strconv"
	"strings"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dissolve/dnssd/advertiser/dnsimple/internal/dnsimplex"
	"golang.org/x/exp/slices"
)

func (a *Advertiser) findTXT(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
) ([]dnsimple.ZoneRecord, error) {
	return dnsimplex.All(
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
					Type: dnsimple.String("TXT"),
				},
			)
			if err != nil {
				return nil, nil, dnsimplex.Errorf("unable to list TXT records: %w", err)
			}

			return res.Pagination, res.Data, nil
		},
	)
}

func (a *Advertiser) syncTXT(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, err := a.findTXT(ctx, zone, inst)
	if err != nil {
		return err
	}

	var desired []dnsimple.ZoneRecordAttributes

	for _, r := range dnssd.NewTXTRecords(inst) {
		content := strings.TrimPrefix(r.String(), r.Hdr.String())

		if content == `""` {
			// DNSimple does not allow empty TXT records, but
			// https://datatracker.ietf.org/doc/html/rfc6763#section-6 requires
			// TXT record in all cases.
			//
			// As a workaround, we exploit the requirement at
			// https://datatracker.ietf.org/doc/html/rfc6763#section-6.4, which
			// states:
			//
			// > DNS-SD TXT record strings beginning with an '=' character >
			// (i.e., the key is missing) MUST be silently ignored.
			content = `"="`
		}

		desired = append(
			desired,
			dnsimple.ZoneRecordAttributes{
				ZoneID: zone.Name,
				Type:   "TXT",
				Name: dnsimple.String(
					dnssd.EscapeInstance(inst.Name) + "." + inst.ServiceType,
				),
				Content: content,
				TTL:     int(inst.TTL.Seconds()),
			},
		)
	}

next:
	for _, c := range current {
		for i, d := range desired {
			if c.Content == d.Content {
				// We consider a TXT record with the same content to be the same
				// record.
				desired = slices.Delete(desired, i, i+1)
				cs.Update(c, d)
				continue next
			}
		}

		cs.Delete(c)
	}

	for _, attr := range desired {
		cs.Create(attr)
	}

	return nil
}

func (a *Advertiser) deleteTXT(
	ctx context.Context,
	zone *dnsimple.Zone,
	inst dnssd.ServiceInstance,
	cs *changeSet,
) error {
	current, err := a.findTXT(ctx, zone, inst)
	if err != nil {
		return err
	}

	for _, c := range current {
		cs.Delete(c)
	}

	return nil
}
