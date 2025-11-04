package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type sensorInstance struct {
	repo[domain.SensorInstance]
}

var _ ports.SensorInstanceRepository = (*sensorInstance)(nil)

func NewSensorInstance(client *HealthyClient) *sensorInstance {
	return &sensorInstance{
		repo: repo[domain.SensorInstance]{
			Rdb: client,
			Key: cache.SensorInstanceKey,
		},
	}
}

func (r *sensorInstance) CmdSetMany(ctx context.Context, insts []*domain.SensorInstance) (*StatusCmd, *MultiBoolCmd) {
	return r.repo.CmdSetMany(ctx, common.SliceToMapByID(insts))
}

func (r *sensorInstance) CmdListIDsByGatewayID(ctx context.Context, gtwID string) *StringSliceCmd {
	return r.repo.CmdListIDsByParentID(ctx, gtwID, cache.SensorGatewayKey)
}

func (r *sensorInstance) ListByGatewayID(ctx context.Context, gtwID ulid.ULID) ([]*domain.SensorInstance, error) {
	return r.repo.GetManyByParentID(ctx, gtwID.String(), cache.SensorGatewayKey)
}

func (r *sensorInstance) SetIDsByGatewayID(ctx context.Context, gtwID ulid.ULID, insts []*domain.SensorInstance) error {
	return r.repo.SetIDsByParentID(ctx, gtwID.String(), common.SliceToIDs(insts), cache.SensorGatewayKey)
}
