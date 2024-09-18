package ethtxmanager

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ethermanInterface interface {
	GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error)
	GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
	WaitTxToBeMined(ctx context.Context, tx *types.Transaction, timeout time.Duration) (bool, error)
	SendTx(ctx context.Context, tx *types.Transaction) error
	CurrentNonce(ctx context.Context, account common.Address) (uint64, error)
	PendingNonce(ctx context.Context, account common.Address) (uint64, error)
	SuggestedGasPrice(ctx context.Context) (*big.Int, error)
	EstimateGas(ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte) (uint64, error)
	EstimateGasBlobTx(
		ctx context.Context,
		from common.Address,
		to *common.Address,
		gasFeeCap *big.Int,
		gasTipCap *big.Int,
		value *big.Int,
		data []byte,
	) (uint64, error)
	CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error)
	SignTx(ctx context.Context, sender common.Address, tx *types.Transaction) (*types.Transaction, error)
	GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error)
	GetLatestBlockNumber(ctx context.Context) (uint64, error)
	GetHeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	GetSuggestGasTipCap(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
}

type storageInterface interface {
	Add(ctx context.Context, mTx monitoredTx) error
	Remove(ctx context.Context, id common.Hash) error
	Get(ctx context.Context, id common.Hash) (monitoredTx, error)
	GetByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]monitoredTx, error)
	GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]monitoredTx, error)
	Update(ctx context.Context, mTx monitoredTx) error
	Empty(ctx context.Context) error
}
