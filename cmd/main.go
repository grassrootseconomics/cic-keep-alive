package main

import (
	"cic-keep-alive/internal/health"
	"cic-keep-alive/pkg/nonce"
	"context"
	"os"
	"os/signal"
	"syscall"

	eth_crypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/knadh/koanf"
	"github.com/lmittmann/w3"
	"github.com/zerodha/logf"
)

var (
	lo                     logf.Logger
	ko                     *koanf.Koanf
	pg                     *pgxpool.Pool
	provider               *chain.Provider
	asynqClient            *asynq.Client
	healthCheck            *health.Health
	healthCheckConstraints *health.HealthCheckConstraints
)

func init() {
	var err error

	ko = initConfig("config.toml")

	lo = initLogger(ko)

	healthCheckConstraints = initHealthCheckContraints(ko)

	pg, err = initDb(ko)
	if err != nil {
		lo.Fatal("could not connect to postgres")
	}

	provider, err = initProvider(ko)
	if err != nil {
		lo.Fatal("could not connect to rpc endpoint")
	}

	asynqClient = initAsynqClient(ko)
}

func main() {
	var dContainer daemonContainer

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	nonceManager, err := nonce.Register(provider, w3.A(ko.MustString("signer.address")))
	if err != nil {
		lo.Fatal("could not init nonce manager", "err", err)
	}
	lo.Debug("initialized nonce value", "nonce", nonceManager.PeekNonce())

	loadPrivateKey, err := eth_crypto.HexToECDSA(ko.MustString("signer.key"))
	if err != nil {
		lo.Fatal("could not load eth private key")
	}

	healthCheck = health.Register(&health.HealthOpts{
		Lo:           lo,
		Pg:           pg,
		Provider:     provider,
		AsynqClient:  asynqClient,
		NonceManager: nonceManager,
		AccountKey:   loadPrivateKey,
		Constraints:  healthCheckConstraints,
	})

	scheduler, err := initScheduler(ko)
	if err != nil {
		lo.Fatal("could not init scheduler")
	}
	dContainer.scheduler = scheduler

	server, mux := initProcessor(ko)
	dContainer.processor = server
	dContainer.mux = mux

	bootstrapDaemon(ctx, dContainer)
}
