package orchestrators

/* type SensorGateway struct {
	dbRepo            ports.SensorGatewayRepository
	dbInstanceRepo    ports.SensorInstanceRepository
	redisRepo         redisadapter.SensorGateway
	redisInstanceRepo redisadapter.SensorInstance
	localCache        cache.Local[*domain.SensorGateway]
	logger            log.Logger
	asyncPool         ports.AsyncSubmitter[ports.Task]
}

var _ ports.SensorGatewayOrchestrator = (*SensorGateway)(nil)

func NewSensorGateway(
	dbRepo ports.SensorGatewayRepository,
	dbInstanceRepo ports.SensorInstanceRepository,
	redisRepo redisadapter.SensorGateway,
	redisInstanceRepo redisadapter.SensorInstance,
	localCache cache.Local[*domain.SensorGateway],
	logger log.Logger,
	asyncPool ports.AsyncSubmitter[ports.Task],
) *SensorGateway {
	return &SensorGateway{
		dbRepo:            dbRepo,
		dbInstanceRepo:    dbInstanceRepo,
		redisRepo:         redisRepo,
		redisInstanceRepo: redisInstanceRepo,
		localCache:        localCache,
		logger:            logger,
		asyncPool:         asyncPool,
	}
}

func (o *SensorGateway) GetOneWithInstances(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error) {
	hydratedKey := cache.FormatHydratedKey(cache.SensorGatewayKey, id.String(), cache.WithSensorInstancesKey)

	if gw, ok := o.localCache.Get(hydratedKey); ok {
		return gw, nil
	}

	if redisCli := o.redisRepo.Client(); redisCli.IsAlive() {
		if gw, found := o.getFromRedisWithInstances(ctx, id); found {
			return gw, nil
		}
	}

	gw, err := o.getFromDB(ctx, id)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	o.localCache.Set(hydratedKey, gw)

	if redisCli := o.redisRepo.Client(); redisCli.IsAlive() {
		o.asyncSetRedis(gw)
	}

	return gw, nil
}

// getFromRedis tries to fetch gateway + instances from Redis
// Returns the gateway and true if gateway was found (even if instances were missing)
func (o *SensorGateway) getFromRedisWithInstances(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, bool) {
	strID := id.String()

	var (
		gateway     *domain.SensorGateway
		instanceIDs []string
		instances   []*domain.SensorInstance
		redisErrors []error
	)

	defer func() {
		if len(redisErrors) > 0 {
			o.logger.Warn(errors.JoinWrap(redisErrors...))
		}
	}()

	redisCli := o.redisRepo.Client()
	pipe, ctxPipe := redisCli.NewPipeline(ctx)

	gatewayCmd := o.redisRepo.CmdGetOneByStrID(ctxPipe, strID)
	idsCmd := o.redisInstanceRepo.CmdListIDsByGatewayID(ctxPipe, strID)

	if _, err := pipe.Exec(ctxPipe); err != nil {
		redisErrors = append(redisErrors, redisCli.HandleError(err))
	}

	gateway, err := gatewayCmd.Result()
	if err != nil {
		if !errors.Is(err, errors.ErrNotFound) {
			redisErrors = append(redisErrors, err)
		}
		return nil, false
	}

	instanceIDs, err = idsCmd.Result()
	if err != nil {
		if !errors.Is(err, errors.ErrNotFound) {
			redisErrors = append(redisErrors, err)
		}
	}

	if len(instanceIDs) > 0 {
		var err error
		instances, err = o.redisInstanceRepo.GetManyByStrIDs(ctx, instanceIDs)
		if err != nil {
			if !errors.Is(err, errors.ErrNotFound) {
				redisErrors = append(redisErrors, err)
			}
		}
	}

	if len(instances) > 0 {
		gateway.SensorInstances = instances
	} else {
		insts, err := o.dbInstanceRepo.ListByGatewayID(ctx, id)
		if err != nil {
			o.logger.Warn(errors.WrapErr(err))
		} else {
			gateway.SensorInstances = insts
		}
	}

	hydratedKey := cache.FormatHydratedKey(cache.SensorGatewayKey, strID, cache.WithSensorInstancesKey)
	o.localCache.Set(hydratedKey, gateway)

	return gateway, true
}

// getFromDB fetches gateway + instances from DB
func (o *SensorGateway) getFromDB(ctx context.Context, id uuid.UUID) (*domain.SensorGateway, error) {
	gw, err := o.dbRepo.GetOneByID(ctx, id)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	insts, err := o.dbInstanceRepo.ListByGatewayID(ctx, id)
	if err != nil {
		return nil, errors.WrapErr(err)
	}

	gw.SensorInstances = insts
	return gw, nil
}

// asyncSetRedis pushes gateway + instances to Redis asynchronously
func (o *SensorGateway) asyncSetRedis(gw *domain.SensorGateway) {
	if err := o.asyncPool.Submit(worker.VoidTask(func() {
		var redisErrors []error
		defer func() {
			if len(redisErrors) > 0 {
				o.logger.Warn(errors.JoinWrap(redisErrors...))
			}
		}()

		redisCli := o.redisRepo.Client()
		pipe, ctxPipe := redisCli.NewPipeline(context.Background())

		setGtwCmd, err := o.redisRepo.CmdSetOne(ctxPipe, gw)
		if err != nil {
			redisErrors = append(redisErrors, redisCli.HandleError(err))
			return
		}
		setInstsCmd, setInstsTtlCmds, err := o.redisInstanceRepo.CmdSetMany(ctxPipe, gw.SensorInstances)
		if err != nil {
			redisErrors = append(redisErrors, redisCli.HandleError(err))
			return
		}

		if _, err := pipe.Exec(ctxPipe); err != nil {
			redisErrors = append(redisErrors, redisCli.HandleError(err))
		}

		if _, err := setGtwCmd.Result(); err != nil {
			redisErrors = append(redisErrors, err)
		}
		if _, err := setInstsCmd.Result(); err != nil {
			redisErrors = append(redisErrors, err)
		}
		if _, err := setInstsTtlCmds.Result(); err != nil {
			redisErrors = append(redisErrors, err)
		}

		if len(redisErrors) == 0 {
			if err := o.redisInstanceRepo.SetIDsByGatewayID(context.Background(), gw.ID, gw.SensorInstances); err != nil {
				redisErrors = append(redisErrors, err)
			}
		}

	})); err != nil {
		o.logger.Warn(errors.WrapErr(err))
	}
}
*/
