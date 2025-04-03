package route53

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
	"github.com/miekg/dns"
)

// Advertiser is a [dnssd.Advertiser] implementation that advertises DNS-SD service
// instances on domain names hosted by Amazon Route 53.
type Advertiser struct {
	Client      *route53.Client
	PartitionID string

	zoneIDs sync.Map // map[string]string
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

	zoneID, err := a.lookupZoneID(ctx, inst.Domain)
	if err != nil {
		return false, err
	}

	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"advertising DNS-SD %s instance: %s ",
			inst.ServiceType,
			inst.Name,
		)),
	}

	if err := a.syncPTR(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	if err := a.syncSRV(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	if err := a.syncTXT(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	return a.apply(ctx, zoneID, cs)
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
	zoneID, err := a.lookupZoneID(ctx, inst.Domain)
	if err != nil {
		return false, err
	}

	cs := &types.ChangeBatch{
		Comment: aws.String(fmt.Sprintf(
			"unadvertising DNS-SD %s instance: %s ",
			inst.ServiceType,
			inst.Name,
		)),
	}

	if err := a.deletePTR(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	if err := a.deleteSRV(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	if err := a.deleteTXT(ctx, zoneID, inst, cs); err != nil {
		return false, err
	}

	return a.apply(ctx, zoneID, cs)
}

func (a *Advertiser) lookupZoneID(
	ctx context.Context,
	domain string,
) (string, error) {
	if id, ok := a.zoneIDs.Load(domain); ok {
		return id.(string), nil
	}

	out, err := a.Client.ListHostedZonesByName(
		ctx,
		&route53.ListHostedZonesByNameInput{
			DNSName:  aws.String(domain + "."),
			MaxItems: aws.Int32(1),
		},
	)
	if err != nil {
		return "", fmt.Errorf("unable to list zones: %w", err)
	}

	if len(out.HostedZones) > 0 {
		zone := out.HostedZones[0]

		if *zone.Name == domain+"." {
			id, _ := a.zoneIDs.LoadOrStore(domain, *zone.Id)
			return id.(string), nil
		}
	}

	return "", dnssd.UnsupportedDomainError{
		Domain: domain,
		Cause:  fmt.Errorf("no AWS Route53 zone found for %q", domain),
	}
}

func (a *Advertiser) apply(
	ctx context.Context,
	zoneID string,
	cs *types.ChangeBatch,
) (bool, error) {
	if len(cs.Changes) == 0 {
		return false, nil
	}

	_, err := a.Client.ChangeResourceRecordSets(
		ctx,
		&route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(zoneID),
			ChangeBatch:  cs,
		},
	)

	return true, err
}

func (a *Advertiser) findResourceRecordSet(
	ctx context.Context,
	zoneID string,
	name *string,
	recordType types.RRType,
) (types.ResourceRecordSet, bool, error) {
	out, err := a.Client.ListResourceRecordSets(
		ctx,
		&route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(zoneID),
			StartRecordName: name,
			StartRecordType: recordType,
			MaxItems:        aws.Int32(1),
		},
	)
	if err != nil {
		return types.ResourceRecordSet{}, false, err
	}

	if len(out.ResourceRecordSets) == 0 {
		return types.ResourceRecordSet{}, false, nil
	}

	set := out.ResourceRecordSets[0]

	if !strings.EqualFold(*set.Name, *name) {
		return types.ResourceRecordSet{}, false, nil
	}

	if set.Type != recordType {
		return types.ResourceRecordSet{}, false, nil
	}

	return set, true, nil
}

func instanceName(inst dnssd.ServiceInstance) *string {
	return aws.String(
		inst.Absolute(),
	)
}

func serviceName(inst dnssd.ServiceInstance) *string {
	return aws.String(
		dnssd.AbsoluteInstanceEnumerationDomain(inst.ServiceType, inst.Domain),
	)
}

func convertRecords[
	R interface {
		Header() *dns.RR_Header
		String() string
	},
](records ...R) []types.ResourceRecord {
	var result []types.ResourceRecord

	for _, rec := range records {
		result = append(
			result,
			types.ResourceRecord{
				Value: aws.String(
					strings.TrimPrefix(
						rec.String(),
						rec.Header().String(),
					),
				),
			},
		)
	}

	return result
}
