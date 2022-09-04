package health

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/golang-module/carbon"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/hibiken/asynq"
	"github.com/lmittmann/w3"
	"github.com/lmittmann/w3/module/eth"
)

var (
	PERIOD_DURATION float64 = 604800.0
	SECONDS_DAY     float64 = 86400.0
)

func (j *Health) HealthCheckProcessor(ctx context.Context, t *asynq.Task) error {
	var (
		p               TaskPayload
		redistributions [32]byte
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	tokenInfo, err := j.provider.ERC20TokenInfo(ctx, p.TokenAddress)
	if err != nil {
		return err
	}

	periodsElapsed := (float64(carbon.Now().Timestamp()) - float64(tokenInfo.PeriodStart.Int64())) / PERIOD_DURATION
	lastDemurrageElapsedDays := (float64(carbon.Now().Timestamp()) - float64(tokenInfo.DemurrageTimestamp.Int64())) / SECONDS_DAY

	err = j.provider.EthClient.CallCtx(
		ctx,
		eth.CallFunc(w3.MustNewFunc("redistributions(uint256)", "bytes32"), p.TokenAddress, big.NewInt(int64(periodsElapsed))).Returns(&redistributions),
	)
	if err != nil {
		task, err := j.NewChangePeriodJob(p.TokenAddress)
		if err != nil {
			return err
		}

		t, err := j.asynqClient.Enqueue(task)
		if err != nil {
			return err
		}

		j.lo.Info("changePeriodNotCalled", "token", tokenInfo.Symbol, "task", t.ID)
		return nil
	}

	if lastDemurrageElapsedDays > j.c.MinLastDemurrageApplyDayElapsed {
		task, err := j.NewChangePeriodJob(p.TokenAddress)
		if err != nil {
			return err
		}

		t, err := j.asynqClient.Enqueue(task)
		if err != nil {
			return err
		}

		j.lo.Info("demurrageTimeStampLag", "token", tokenInfo.Symbol, "task", t.ID)
		return nil
	}

	return nil
}

func (j *Health) ChangePeriodProcessor(ctx context.Context, t *asynq.Task) error {
	var (
		changePeriodSig = w3.MustNewFunc("changePeriod()", "bool")

		p      TaskPayload
		txHash common.Hash
	)

	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	preparedInputData, err := changePeriodSig.EncodeArgs()
	if err != nil {
		return fmt.Errorf("sig encoding failed: %v: %w", err, asynq.SkipRetry)
	}

	txData := chain.TransactionData{
		To:        p.TokenAddress,
		InputData: preparedInputData,
		GasLimit:  j.c.GasLimit,
		Nonce:     j.nm.AcquireNonce(),
	}

	builtTx, err := j.provider.BuildKitabuTx(j.ak, txData)
	if err != nil {
		j.nm.ReturnNonce()
		return fmt.Errorf("building tx failed: %v: %w", err, asynq.SkipRetry)
	}

	err = j.provider.EthClient.CallCtx(
		ctx,
		eth.SendTransaction(builtTx).Returns(&txHash),
	)
	if err != nil {
		return fmt.Errorf("tx submission failed: %v: %w", err, asynq.SkipRetry)
	}

	j.lo.Info("changePeriod tx successfully submitted", "hash", txHash.Hex(), "token", p.TokenAddress.String())
	return nil
}
