package health

import (
	"context"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/lmittmann/w3"
)

const (
	TypePeriodicHealthCheckQueuer = "token:health_check_queuer"

	getTokensQuery = "SELECT token_address, token_symbol FROM tokens;"
)

type tokensRes struct {
	Id           int    `db:"id" json:"id"`
	TokenSymbol  string `db:"token_symbol" json:"token_symbol"`
	TokenName    string `db:"token_name" json:"token_name"`
	TokenAddress string `db:"token_address" json:"token_addres"`
}

func (j *Health) PeriodicAllTokensHealthCheckQueuer(ctx context.Context, _ *asynq.Task) error {
	rows, err := j.pg.Query(ctx, getTokensQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	rs := pgxscan.NewRowScanner(rows)

	for rows.Next() {
		var token tokensRes
		if err := rs.Scan(&token); err != nil {
			return err
		}

		task, err := j.NewHealthCheckJob(w3.A(chain.SarafuAddressToChecksum(token.TokenAddress)))
		if err != nil {
			return err
		}

		_, err = j.asynqClient.Enqueue(task)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}
