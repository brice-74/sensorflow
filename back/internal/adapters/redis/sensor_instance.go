package redis

import (
	"context"
	"time"

	"github.com/brice-74/sensorflow/internal/cache"
	"github.com/brice-74/sensorflow/internal/core/domain"
	"github.com/brice-74/sensorflow/internal/core/domain/common"
	"github.com/brice-74/sensorflow/pkg/ulid"
)

type sensorInstance struct {
	repo[domain.SensorInstance]
}

var _ SensorInstance = (*sensorInstance)(nil)

func NewSensorInstance(client *HealthyClient, defaultTTL time.Duration) *sensorInstance {
	return &sensorInstance{
		repo: repo[domain.SensorInstance]{
			HealthyClient: client,
			Key:           cache.SensorInstanceKey,
			DefaultTTL:    defaultTTL,
		},
	}
}

func (r *sensorInstance) CmdSetMany(ctx context.Context, insts []*domain.SensorInstance) (*StatusCmd, *MultiBoolCmd) {
	return r.repo.CmdSetManyByStrID(ctx, common.SliceToMapByID(insts))
}

func (r *sensorInstance) CmdListIDsByGatewayID(ctx context.Context, gtwID string) *StringSliceCmd {
	return r.repo.CmdListStrIDsByParentID(ctx, gtwID, cache.SensorGatewayKey)
}

func (r *sensorInstance) ListByGatewayID(ctx context.Context, gtwID ulid.ULID) ([]*domain.SensorInstance, error) {
	return r.repo.GetManyByParentStrID(ctx, gtwID.String(), cache.SensorGatewayKey)
}

func (r *sensorInstance) SetIDsByGatewayID(ctx context.Context, gtwID ulid.ULID, insts []*domain.SensorInstance) error {
	return r.repo.SetStrIDsByParentID(ctx, gtwID.String(), common.SliceToIDs(insts), cache.SensorGatewayKey)
}

func (r *sensorInstance) GetOneByID(ctx context.Context, ID ulid.ULID) (*domain.SensorInstance, error) {
	return r.repo.GetOneByStrID(ctx, ID.String())
}
