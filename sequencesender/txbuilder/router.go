package txbuilder

import (
	"context"
	"fmt"

	"github.com/0xPolygon/cdk/etherman/types"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type TxBuilderRouter struct {
	implementations map[string]TxBuilder
	selector        Selector
}

func NewTxBuilderSelector(selector Selector) *TxBuilderRouter {
	return &TxBuilderRouter{
		implementations: make(map[string]TxBuilder),
		selector:        selector,
	}
}

func (t *TxBuilderRouter) Register(name string, txBuilder TxBuilder) *TxBuilderRouter {
	t.implementations[name] = txBuilder
	return t
}

func (t *TxBuilderRouter) Get(name string) TxBuilder {
	impl, ok := t.implementations[name]
	if !ok {
		return nil
	}
	return impl
}

func (t *TxBuilderRouter) BuildSequenceBatchesTx(ctx context.Context, sequences []types.Sequence, lastSequence types.Sequence, firstSequence types.Sequence) (*ethtypes.Transaction, error) {
	which, err := t.selector.Get(ctx)
	if err != nil {
		return nil, err
	}
	impl := t.Get(which)
	if impl == nil {
		return nil, fmt.Errorf("no implementation found for selector %s", which)
	}
	return impl.BuildSequenceBatchesTx(ctx, sequences, lastSequence, firstSequence)

}
