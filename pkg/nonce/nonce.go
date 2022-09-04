package nonce

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/cic-go-sdk/chain"
	"github.com/lmittmann/w3/module/eth"
)

type NonceManager struct {
	mx             sync.Mutex
	nonceValue     uint64
	provider       *chain.Provider
	accountAddress common.Address
}

func Register(provider *chain.Provider, accountAddress common.Address) (*NonceManager, error) {
	var (
		networkNonce uint64
	)

	err := provider.EthClient.CallCtx(
		context.Background(),
		eth.Nonce(accountAddress, nil).Returns(&networkNonce),
	)
	if err != nil {
		return nil, err
	}

	return &NonceManager{
		nonceValue:     networkNonce,
		provider:       provider,
		accountAddress: accountAddress,
	}, nil
}

func (n *NonceManager) PeekNonce() uint64 {
	n.mx.Lock()
	defer n.mx.Unlock()

	return n.nonceValue
}

func (n *NonceManager) AcquireNonce() uint64 {
	n.mx.Lock()
	defer n.mx.Unlock()

	nextNonce := n.nonceValue
	n.nonceValue++

	return nextNonce
}

func (n *NonceManager) ReturnNonce() {
	n.mx.Lock()
	defer n.mx.Unlock()

	if n.nonceValue > 0 {
		n.nonceValue--
	}
}
