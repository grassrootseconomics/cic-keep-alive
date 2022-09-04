package main

import (
	"cic-keep-alive/internal/health"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/zerodha/logf"
)

func initConfig(configFilePath string) *koanf.Koanf {
	var (
		ko = koanf.New(".")
	)

	confFile := file.Provider(configFilePath)
	if err := ko.Load(confFile, toml.Parser()); err != nil {
		fmt.Printf("could not load config file: %v\n", err)
		os.Exit(1)
	}

	if err := ko.Load(env.Provider("", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(
			strings.TrimPrefix(s, "")), "_", ".")
	}), nil); err != nil {
		fmt.Printf("could not override config from env vars: %v\n", err)
		os.Exit(1)
	}

	return ko
}

func initLogger(ko *koanf.Koanf) logf.Logger {
	loOpts := logf.Opts{
		EnableColor:  true,
		EnableCaller: true,
	}

	if ko.Bool("service.debug") {
		loOpts.Level = logf.DebugLevel
	} else {
		loOpts.Level = logf.InfoLevel
	}

	lo := logf.New(loOpts)

	return lo
}

func initDb(ko *koanf.Koanf) (*pgxpool.Pool, error) {
	pgxConfig := ko.MustString("postgres.dsn")

	db, err := pgxpool.Connect(context.Background(), pgxConfig)
	if err != nil {
		return nil, err
	}

	lo.Debug("successfully connected to pg pool")
	return db, nil
}

func initProvider(ko *koanf.Koanf) (*chain.Provider, error) {
	provider, err := chain.NewProvider(ko.MustString("chain.endpoint"))
	if err != nil {
		return nil, err
	}

	lo.Debug("successfully parsed kitabu rpc endpoint")
	return provider, nil
}

func initAsynqClient(ko *koanf.Koanf) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: ko.MustString("redis.dsn")})
}

func initHealthCheckContraints(ko *koanf.Koanf) *health.HealthCheckConstraints {
	return &health.HealthCheckConstraints{
		GasLimit:                        uint64(ko.MustInt64("chain.gaslimit")),
		MinLastDemurrageApplyDayElapsed: ko.MustFloat64("checks.min"),
	}
}
