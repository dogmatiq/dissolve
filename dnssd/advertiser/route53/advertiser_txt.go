package route53

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/dogmatiq/dissolve/dnssd"
)

func (a *Advertiser) findTXT(
	ctx context.Context,
	zoneID string,
	inst dnssd.ServiceInstance,
) (types.ResourceRecordSet, bool, error) {
	return a.findResourceRecordSet(
		ctx,
		zoneID,
		instanceName(inst),
		types.RRTypeTxt,
	)
}

func (a *Advertiser) syncTXT(
	ctx context.Context,
	zoneID string,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	desired := types.ResourceRecordSet{
		Name: instanceName(inst),
		Type: types.RRTypeTxt,
		TTL:  aws.Int64(int64(inst.TTL.Seconds())),
		ResourceRecords: convertRecords(
			dnssd.NewTXTRecords(inst)...,
		),
	}

	current, ok, err := a.findTXT(ctx, zoneID, inst)
	if err != nil {
		return err
	}

	if !ok {
		cs.Changes = append(
			cs.Changes,
			types.Change{
				Action:            types.ChangeActionCreate,
				ResourceRecordSet: &desired,
			},
		)
		return nil
	}

	if reflect.DeepEqual(current, desired) {
		return nil
	}

	cs.Changes = append(
		cs.Changes,
		types.Change{
			Action:            types.ChangeActionUpsert,
			ResourceRecordSet: &desired,
		},
	)

	return nil
}

func (a *Advertiser) deleteTXT(
	ctx context.Context,
	zoneID string,
	inst dnssd.ServiceInstance,
	cs *types.ChangeBatch,
) error {
	current, ok, err := a.findTXT(ctx, zoneID, inst)
	if err != nil {
		return err
	}

	if ok {
		cs.Changes = append(
			cs.Changes,
			types.Change{
				Action:            types.ChangeActionDelete,
				ResourceRecordSet: &current,
			},
		)
	}

	return nil
}
