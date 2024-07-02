package etherman2

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/etherman2/internal"
	"github.com/ethereum/go-ethereum/ethclient"
)

type etherman2Impl struct {
	Etherman2ChainReader
	Etherman2AuthStorerRW
}

/*
Example of usage:
	etherman2builder := etherman2.NewEtherman2Builder().
		ChainReader("http://localhost:8545").
		AuthStore(nil)
	etherman2, err:= etherman2builder.Build(context.Background())
*/

type Etherman2Builder struct {
	L1EthClient *ethclient.Client
	chainReader struct {
		enabled bool
		l1url   string
	}
	authStore *struct {
		enabled       bool
		forcedChainID *uint64
	}

	result etherman2Impl
}

func NewEtherman2Builder() *Etherman2Builder {
	return &Etherman2Builder{}
}

func (e *Etherman2Builder) ChainReader(l1url string) *Etherman2Builder {
	e.chainReader.enabled = true
	e.chainReader.l1url = l1url
	return e
}

func (e *Etherman2Builder) AuthStore(forcedChainID *uint64) *Etherman2Builder {
	e.authStore.enabled = true
	e.authStore.forcedChainID = forcedChainID
	return e
}

func (e *Etherman2Builder) Build(ctx context.Context) (*etherman2Impl, error) {
	if e.chainReader.enabled {
		err := e.newChainReader()
		if err != nil {
			return nil, err
		}
	}
	if e.authStore.enabled {
		err := e.newAuthStore(ctx)
		if err != nil {
			return nil, err
		}
	}

	return &e.result, nil
}

func (e *Etherman2Builder) newChainReader() error {
	if e.chainReader.enabled {
		client, err := ethclient.Dial(e.chainReader.l1url)
		if err != nil {
			return err
		}
		e.L1EthClient = client
	}
	if e.chainReader.enabled {
		var err error
		e.result.Etherman2ChainReader, err = internal.NewEtherman2ChainReader(e.L1EthClient)
		if err != nil {
			return err
		}

		return nil
	}
	return nil
}

func (e *Etherman2Builder) newAuthStore(ctx context.Context) error {
	if e.authStore.enabled {
		chainID, err := e.getChainID(ctx)
		e.result.Etherman2AuthStorerRW, err = internal.NewEtherman2L1Auth(chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Etherman2Builder) getChainID(ctx context.Context) (uint64, error) {
	var chainID uint64
	if e.authStore.forcedChainID != nil {
		chainID = *e.authStore.forcedChainID
		return chainID, nil

	}
	if e.result.Etherman2ChainReader == nil {
		return 0, fmt.Errorf("chain reader not initialized, so or force ChainID or build also a chain reader (ChainReader() )")
	}
	chainID, err := e.result.Etherman2ChainReader.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return chainID, nil
}
