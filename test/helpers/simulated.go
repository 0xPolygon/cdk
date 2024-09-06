package helpers

import (
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/rpc"
)

type ClientRenamed simulated.Client

type TestClient struct {
	ClientRenamed
}

func (tc TestClient) Client() *rpc.Client {
	return nil
}
