package dnsimple

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/dnsimple/dnsimple-go/v4/dnsimple"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/dogmatiq/dissolve/dnssd/advertiser/dnsimple/internal/dnsimplex"
)

// Advertiser is a [dnssd.Advertiser] implementation that advertises DNS-SD service
// instances on domain names hosted by dnsimple.com.
type Advertiser struct {
	Client *dnsimple.Client

	zones sync.Map // map[string]*dnsimple.Zone
}

// Advertise creates and/or updates DNS records to advertise the given service
// instance.
//
// It returns true if any changes to DNS records were made, or false if the
// service was already advertised as-is.
func (a *Advertiser) Advertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
	options ...dnssd.AdvertiseOption,
) (bool, error) {
	if len(options) != 0 {
		return false, errors.New("advertise options are not yet supported")
	}

	zone, err := a.lookupZone(ctx, inst.Domain)
	if err != nil {
		return false, err
	}

	cs := &changeSet{}

	if err := a.syncPTR(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	if err := a.syncSRV(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	if err := a.syncTXT(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	return a.apply(ctx, zone, cs)
}

// Unadvertise removes and/or updates DNS records to stop advertising the given
// service instance.
//
// It true if any changes to DNS records were made, or false if the service was
// not advertised.
func (a *Advertiser) Unadvertise(
	ctx context.Context,
	inst dnssd.ServiceInstance,
) (bool, error) {
	zone, err := a.lookupZone(ctx, inst.Domain)
	if err != nil {
		return false, err
	}

	cs := &changeSet{}

	if err := a.deletePTR(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	if err := a.deleteSRV(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	if err := a.deleteTXT(ctx, zone, inst, cs); err != nil {
		return false, err
	}

	return a.apply(ctx, zone, cs)
}

func (a *Advertiser) apply(
	ctx context.Context,
	zone *dnsimple.Zone,
	cs *changeSet,
) (bool, error) {
	if cs.IsEmpty() {
		return false, nil
	}

	accountID := strconv.FormatInt(zone.AccountID, 10)

	for _, rec := range cs.deletes {
		if _, err := a.Client.Zones.DeleteRecord(ctx, accountID, zone.Name, rec.ID); err != nil {
			return false, dnsimplex.Errorf("unable to delete %s record: %w", rec.Type, err)
		}
	}

	for _, up := range cs.updates {
		if _, err := a.Client.Zones.UpdateRecord(ctx, accountID, zone.Name, up.Before.ID, up.After); err != nil {
			return false, dnsimplex.Errorf("unable to update %s record: %w", up.Before.Type, err)
		}
	}

	for _, attr := range cs.creates {
		if _, err := a.Client.Zones.CreateRecord(ctx, accountID, zone.Name, attr); err != nil {
			return false, dnsimplex.Errorf("unable to create %s record: %w", attr.Type, err)
		}
	}

	return true, nil
}

func (a *Advertiser) lookupZone(
	ctx context.Context,
	domain string,
) (*dnsimple.Zone, error) {
	if zone, ok := a.zones.Load(domain); ok {
		return zone.(*dnsimple.Zone), nil
	}

	zone, ok, err := dnsimplex.Find(
		ctx,
		func(opts dnsimple.ListOptions) (*dnsimple.Pagination, []dnsimple.Account, error) {
			res, err := a.Client.Accounts.ListAccounts(ctx, &opts)
			if err != nil {
				return nil, nil, dnsimplex.Errorf("unable to list accounts: %w", err)
			}
			return res.Pagination, res.Data, err
		},
		func(acc dnsimple.Account) (*dnsimple.Zone, bool, error) {
			res, err := a.Client.Zones.GetZone(
				ctx,
				strconv.FormatInt(acc.ID, 10),
				domain,
			)
			if err != nil {
				if dnsimplex.IsNotFound(err) {
					return nil, false, nil
				}

				return nil, false, dnsimplex.Errorf(
					"unable to get zone for %q on account %d: %w",
					domain,
					acc.ID,
					err,
				)
			}

			return res.Data, true, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, dnssd.UnsupportedDomainError{
			Domain: domain,
			Cause:  fmt.Errorf("no DNSimple zone found for %q", domain),
		}
	}

	v, _ := a.zones.LoadOrStore(domain, zone)
	return v.(*dnsimple.Zone), nil
}
