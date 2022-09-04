package health

import (
	"cic-keep-alive/pkg/nonce"
	"crypto/ecdsa"

	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zerodha/logf"
)

type HealthCheckConstraints struct {
	GasLimit                        uint64
	MinLastDemurrageApplyDayElapsed float64
}

type HealthOpts struct {
	Lo           logf.Logger
	Pg           *pgxpool.Pool
	Provider     *chain.Provider
	AsynqClient  *asynq.Client
	NonceManager *nonce.NonceManager
	AccountKey   *ecdsa.PrivateKey
	Constraints  *HealthCheckConstraints
}

type Health struct {
	lo          logf.Logger
	pg          *pgxpool.Pool
	provider    *chain.Provider
	asynqClient *asynq.Client
	nm          *nonce.NonceManager
	ak          *ecdsa.PrivateKey
	c           *HealthCheckConstraints
}

func Register(opts *HealthOpts) *Health {
	return &Health{
		lo:          opts.Lo,
		pg:          opts.Pg,
		provider:    opts.Provider,
		asynqClient: opts.AsynqClient,
		nm:          opts.NonceManager,
		ak:          opts.AccountKey,
		c:           opts.Constraints,
	}
}
