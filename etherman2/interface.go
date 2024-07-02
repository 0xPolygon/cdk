package etherman2

import (
	"context"
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type Etherman2ChainReader interface {
	LastL1BlockNumber(ctx context.Context) (uint64, error)
	ChainID(ctx context.Context) (uint64, error)
}

type Etherman2AuthStorerRO interface {
	GetAuthByAddress(addr common.Address) (bind.TransactOpts, bool)
}

// Etherman2AuthStorerRW is an interface that extends Etherman2AuthStorerRO providing write capabilities
type Etherman2AuthStorerRW interface {
	Etherman2AuthStorerRO
	LoadAuthFromKeyStore(path, password string) (*bind.TransactOpts, *ecdsa.PrivateKey, error)
	AddOrReplaceAuth(auth bind.TransactOpts) error
}

type Etherman2Composite interface {
	Etherman2ChainReader
	Etherman2AuthStorerRW
}
