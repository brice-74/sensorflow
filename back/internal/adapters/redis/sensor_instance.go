package redis

import (
	"context"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/internal/ports"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type SensorInstance struct {
	repo[domain.SensorInstance]
}

var _ ports.SensorInstanceRepository = (*SensorInstance)(nil)

func NewSensorInstance(client *HealthyClient) *SensorInstance {
	return &SensorInstance{
		repo: repo[domain.SensorInstance]{
			Rdb: client,
			Key: cache.SensorInstanceKey,
		},
	}
}

func (r *SensorInstance) CmdSetMany(ctx context.Context, insts []*domain.SensorInstance) (*StatusCmd, *MultiBoolCmd) {
	return r.repo.CmdSetMany(ctx, common.SliceToMapByID(insts))
}

func (r *SensorInstance) CmdListIDsByGatewayID(ctx context.Context, gtwID string) *StringSliceCmd {
	return r.repo.CmdListIDsByParentID(ctx, gtwID, cache.SensorGatewayKey)
}

func (r *SensorInstance) ListByGatewayID(ctx context.Context, gtwID ulid.ULID) ([]*domain.SensorInstance, error) {
	return r.repo.GetManyByParentID(ctx, gtwID.String(), cache.SensorGatewayKey)
}

func (r *SensorInstance) SetIDsByGatewayID(ctx context.Context, gtwID ulid.ULID, insts []*domain.SensorInstance) error {
	return r.repo.SetIDsByParentID(ctx, gtwID.String(), common.SliceToIDs(insts), cache.SensorGatewayKey)
}
