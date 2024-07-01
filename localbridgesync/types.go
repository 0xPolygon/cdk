package localbridgesync

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Bridge struct {
	LeafType           uint8
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationNetwork uint32
	DestinationAddress common.Address
	Amount             *big.Int
	Metadata           []byte
	DepositCount       uint32
}

type Claim struct {
	// TODO: pre uLxLy there was Index instead of GlobalIndex, should we treat this differently?
	GlobalIndex        *big.Int
	OriginNetwork      uint32
	OriginAddress      common.Address
	DestinationAddress common.Address
	Amount             *big.Int
}

type bridgeEvents struct {
	Bridges []Bridge
	Claims  []Claim
}

type block struct {
	blockHeader
	Events bridgeEvents
}

type blockHeader struct {
	Num  uint64
	Hash common.Hash
}
