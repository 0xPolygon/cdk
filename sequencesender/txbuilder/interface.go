package txbuilder

import (
	"context"

	"github.com/0xPolygon/cdk/etherman/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilder interface {
	BuildSequenceBatchesTx(ctx context.Context, sequences []types.Sequence, lastSequence types.Sequence, firstSequence types.Sequence) (*ethtypes.Transaction, error)
}
