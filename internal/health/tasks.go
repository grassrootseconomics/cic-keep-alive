package health

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/hibiken/asynq"
)

const (
	TypeHealthCheck  = "token:health_check"
	TypeChangePeriod = "token:change_period"
)

type TaskPayload struct {
	TokenAddress common.Address
}

func (j *Health) NewHealthCheckJob(tokenAddress common.Address) (*asynq.Task, error) {
	payload, err := json.Marshal(TaskPayload{
		TokenAddress: tokenAddress,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeHealthCheck, payload), nil
}

func (j *Health) NewChangePeriodJob(tokenAddress common.Address) (*asynq.Task, error) {
	payload, err := json.Marshal(TaskPayload{
		TokenAddress: tokenAddress,
	})
	if err != nil {
		return nil, err
	}

	return asynq.NewTask(TypeChangePeriod, payload), nil
}
