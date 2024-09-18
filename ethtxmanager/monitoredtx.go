package ethtxmanager

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

const (
	// MonitoredTxStatusCreated means the tx was just added to the storage
	MonitoredTxStatusCreated = MonitoredTxStatus("created")

	// MonitoredTxStatusSent means that at least a eth tx was sent to the network
	MonitoredTxStatusSent = MonitoredTxStatus("sent")

	// MonitoredTxStatusFailed means the tx was already mined and failed with an
	// error that can't be recovered automatically, ex: the data in the tx is invalid
	// and the tx gets reverted
	MonitoredTxStatusFailed = MonitoredTxStatus("failed")

	// MonitoredTxStatusMined means the tx was already mined and the receipt
	// status is Successful
	MonitoredTxStatusMined = MonitoredTxStatus("mined")

	// MonitoredTxStatusSafe means the tx was already mined N blocks ago
	MonitoredTxStatusSafe = MonitoredTxStatus("safe")

	// MonitoredTxStatusFinalized means the tx was already mined M (M > N) blocks ago
	MonitoredTxStatusFinalized = MonitoredTxStatus("finalized")
)

// MonitoredTxStatus represents the status of a monitored tx
type MonitoredTxStatus string

// String returns a string representation of the status
func (s MonitoredTxStatus) String() string {
	return string(s)
}

// monitoredTx represents a set of information used to build tx
// plus information to monitor if the transactions was sent successfully
type monitoredTx struct {
	// ID is the tx identifier controller by the caller
	ID common.Hash `mapstructure:"id"`

	// sender of the tx, used to identify which private key should be used to sing the tx
	From common.Address `mapstructure:"from"`

	// receiver of the tx
	To *common.Address `mapstructure:"to"`

	// Nonce used to create the tx
	Nonce uint64 `mapstructure:"nonce"`

	// tx Value
	Value *big.Int `mapstructure:"value"`

	// tx Data
	Data []byte `mapstructure:"data"`

	// tx Gas
	Gas uint64 `mapstructure:"gas"`

	// tx gas offset
	GasOffset uint64 `mapstructure:"gasOffset"`

	// tx gas price
	GasPrice *big.Int `mapstructure:"gasPrice"`

	// blob Sidecar
	BlobSidecar *types.BlobTxSidecar `mapstructure:"blobSidecar"`

	// blob Gas
	BlobGas uint64 `mapstructure:"blobGas"`

	// blob gas price
	BlobGasPrice *big.Int `mapstructure:"blobGasPrice"`

	// gas tip cap
	GasTipCap *big.Int `mapstructure:"gasTipCap"`

	// Status of this monitoring
	Status MonitoredTxStatus `mapstructure:"status"`

	// BlockNumber represents the block where the tx was identified
	// to be mined, it's the same as the block number found in the
	// tx receipt, this is used to control reorged monitored txs
	BlockNumber *big.Int `mapstructure:"blockNumber"`

	// History represent all transaction hashes from
	// transactions created using this struct data and
	// sent to the network
	History map[common.Hash]bool `mapstructure:"history"`

	// CreatedAt date time it was created
	CreatedAt time.Time `mapstructure:"createdAt"`

	// UpdatedAt last date time it was updated
	UpdatedAt time.Time `mapstructure:"updatedAt"`

	// EstimateGas indicates if gas should be estimated or last value shold be reused
	EstimateGas bool `mapstructure:"estimateGas"`
}

// Tx uses the current information to build a tx
func (mTx monitoredTx) Tx() *types.Transaction {
	var tx *types.Transaction
	if mTx.BlobSidecar == nil {
		tx = types.NewTx(&types.LegacyTx{
			To:       mTx.To,
			Nonce:    mTx.Nonce,
			Value:    mTx.Value,
			Data:     mTx.Data,
			Gas:      mTx.Gas + mTx.GasOffset,
			GasPrice: mTx.GasPrice,
		})
	} else {
		tx = types.NewTx(&types.BlobTx{
			To:         *mTx.To,
			Nonce:      mTx.Nonce,
			Value:      uint256.MustFromBig(mTx.Value),
			Data:       mTx.Data,
			GasFeeCap:  uint256.MustFromBig(mTx.GasPrice),
			GasTipCap:  uint256.MustFromBig(mTx.GasTipCap),
			Gas:        mTx.Gas + mTx.GasOffset,
			BlobFeeCap: uint256.MustFromBig(mTx.BlobGasPrice),
			BlobHashes: mTx.BlobSidecar.BlobHashes(),
			Sidecar:    mTx.BlobSidecar,
		})
	}

	return tx
}

// AddHistory adds a transaction to the monitoring history
func (mTx monitoredTx) AddHistory(tx *types.Transaction) error {
	if _, found := mTx.History[tx.Hash()]; found {
		return ErrAlreadyExists
	}
	mTx.History[tx.Hash()] = true
	return nil
}

// historyHashSlice returns the current history field as a string slice
func (mTx *monitoredTx) historyHashSlice() []common.Hash {
	history := make([]common.Hash, 0, len(mTx.History))
	for h := range mTx.History {
		history = append(history, h)
	}
	return history
}

// MonitoredTxResult represents the result of a execution of a monitored tx
type MonitoredTxResult struct {
	ID                 common.Hash
	To                 *common.Address
	Nonce              uint64
	Value              *big.Int
	Data               []byte
	MinedAtBlockNumber *big.Int
	Status             MonitoredTxStatus
	Txs                map[common.Hash]TxResult
}

// TxResult represents the result of a execution of a ethereum transaction in the block chain
type TxResult struct {
	Tx            *types.Transaction
	Receipt       *types.Receipt
	RevertMessage string
}
