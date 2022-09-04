package main

import (
	"cic-keep-alive/internal/health"
	"cic-keep-alive/pkg/util"
	"context"
	"sync"

	"github.com/hibiken/asynq"
	"github.com/knadh/koanf"
)

type daemonContainer struct {
	scheduler *asynq.Scheduler
	processor *asynq.Server
	mux       *asynq.ServeMux
}

func initScheduler(ko *koanf.Koanf) (*asynq.Scheduler, error) {
	scheduler := asynq.NewScheduler(asynq.RedisClientOpt{
		Addr: ko.MustString("redis.dsn"),
	}, &asynq.SchedulerOpts{
		Logger: util.AsynqCompatibleLogger(lo),
	})

	tokenCheck := asynq.NewTask(health.TypePeriodicHealthCheckQueuer, nil)
	_, err := scheduler.Register(ko.MustString("periodic.token"), tokenCheck)
	if err != nil {
		return nil, err
	}

	return scheduler, nil
}

func initProcessor(ko *koanf.Koanf) (*asynq.Server, *asynq.ServeMux) {
	processorServer := asynq.NewServer(
		asynq.RedisClientOpt{
			Addr: ko.MustString("redis.dsn"),
		},
		asynq.Config{
			Concurrency: 5,
			Logger:      util.AsynqCompatibleLogger(lo),
		},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc(health.TypePeriodicHealthCheckQueuer, healthCheck.PeriodicAllTokensHealthCheckQueuer)
	mux.HandleFunc(health.TypeHealthCheck, healthCheck.HealthCheckProcessor)
	mux.HandleFunc(health.TypeChangePeriod, healthCheck.ChangePeriodProcessor)

	return processorServer, mux
}

func bootstrapDaemon(ctx context.Context, container daemonContainer) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := container.scheduler.Run(); err != nil {
			lo.Fatal("could not start scheduler")
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := container.processor.Run(container.mux); err != nil {
			lo.Fatal("could not start job processor")
		}
	}()

	gracefulShutdown(ctx, container)
	wg.Wait()
}

func gracefulShutdown(ctx context.Context, container daemonContainer) {
	<-ctx.Done()
	lo.Info("graceful shutdown triggered")

	container.scheduler.Shutdown()
	lo.Info("scheduler successfully shutdown")

	container.processor.Shutdown()
	lo.Info("job processor successfully shutdown")
}
