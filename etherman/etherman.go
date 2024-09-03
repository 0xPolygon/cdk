package etherman

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/idataavailabilityprotocol"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// EventOrder is the type used to identify the events order
type EventOrder string

const (
	// GlobalExitRootsOrder identifies a GlobalExitRoot event
	GlobalExitRootsOrder EventOrder = "GlobalExitRoots"
	// L1InfoTreeOrder identifies a L1InTree event
	L1InfoTreeOrder EventOrder = "L1InfoTreeOrder"
	// SequenceBatchesOrder identifies a VerifyBatch event
	SequenceBatchesOrder EventOrder = "SequenceBatches"
	// ForcedBatchesOrder identifies a ForcedBatches event
	ForcedBatchesOrder EventOrder = "ForcedBatches"
	// TrustedVerifyBatchOrder identifies a TrustedVerifyBatch event
	TrustedVerifyBatchOrder EventOrder = "TrustedVerifyBatch"
	// VerifyBatchOrder identifies a VerifyBatch event
	VerifyBatchOrder EventOrder = "VerifyBatch"
	// SequenceForceBatchesOrder identifies a SequenceForceBatches event
	SequenceForceBatchesOrder EventOrder = "SequenceForceBatches"
	// ForkIDsOrder identifies an updateZkevmVersion event
	ForkIDsOrder EventOrder = "forkIDs"
)

type ethereumClient interface {
	ethereum.ChainReader
	ethereum.ChainStateReader
	ethereum.ContractCaller
	ethereum.GasEstimator
	ethereum.GasPricer
	ethereum.GasPricer1559
	ethereum.LogFilterer
	ethereum.TransactionReader
	ethereum.TransactionSender

	bind.DeployBackend
}

// L1Config represents the configuration of the network used in L1
type L1Config struct {
	// Chain ID of the L1 network
	L1ChainID uint64 `json:"chainId" mapstructure:"ChainID"`
	// ZkEVMAddr Address of the L1 contract polygonZkEVMAddress
	ZkEVMAddr common.Address `json:"polygonZkEVMAddress" mapstructure:"ZkEVMAddr"`
	// RollupManagerAddr Address of the L1 contract
	RollupManagerAddr common.Address `json:"polygonRollupManagerAddress" mapstructure:"RollupManagerAddr"`
	// PolAddr Address of the L1 Pol token Contract
	PolAddr common.Address `json:"polTokenAddress" mapstructure:"PolAddr"`
	// GlobalExitRootManagerAddr Address of the L1 GlobalExitRootManager contract
	GlobalExitRootManagerAddr common.Address `json:"polygonZkEVMGlobalExitRootAddress" mapstructure:"GlobalExitRootManagerAddr"` //nolint:lll
}

// Client is a simple implementation of EtherMan.
type Client struct {
	EthClient  ethereumClient
	DAProtocol *idataavailabilityprotocol.Idataavailabilityprotocol

	Contracts *contracts.Contracts
	RollupID  uint32

	l1Cfg config.L1Config
	cfg   config.Config
	auth  map[common.Address]bind.TransactOpts // empty in case of read-only client
}

// NewClient creates a new etherman.
func NewClient(cfg config.Config, l1Config config.L1Config, commonConfig cdkcommon.Config) (*Client, error) {
	// Connect to ethereum node
	ethClient, err := ethclient.Dial(cfg.EthermanConfig.URL)
	if err != nil {
		log.Errorf("error connecting to %s: %+v", cfg.EthermanConfig.URL, err)

		return nil, err
	}
	L1chainID, err := ethClient.ChainID(context.Background())
	if err != nil {
		log.Errorf("error getting L1chainID from %s: %+v", cfg.EthermanConfig.URL, err)

		return nil, err
	}
	log.Infof("L1ChainID: %d", L1chainID.Uint64())
	contracts, err := contracts.NewContracts(l1Config, ethClient)
	if err != nil {
		return nil, err
	}
	log.Info(contracts.String())
	// Get RollupID
	rollupID, err := contracts.Banana.RollupManager.RollupAddressToID(&bind.CallOpts{Pending: false}, l1Config.ZkEVMAddr)
	if err != nil {
		log.Errorf("error getting rollupID from %s : %+v", contracts.Banana.RollupManager.String(), err)

		return nil, err
	}
	if rollupID == 0 {
		return nil, errors.New(
			"rollupID is 0, is not a valid value. Check that rollup Address is correct " +
				l1Config.ZkEVMAddr.String(),
		)
	}
	log.Infof("rollupID: %d (obtenied from SMC: %s )", rollupID, contracts.Banana.RollupManager.String())

	client := &Client{
		EthClient: ethClient,
		Contracts: contracts,

		RollupID: rollupID,
		l1Cfg:    l1Config,
		cfg:      cfg,
		auth:     map[common.Address]bind.TransactOpts{},
	}

	if commonConfig.IsValidiumMode {
		dapAddr, err := contracts.Banana.Rollup.DataAvailabilityProtocol(&bind.CallOpts{Pending: false})
		if err != nil {
			return nil, err
		}

		client.DAProtocol, err = idataavailabilityprotocol.NewIdataavailabilityprotocol(dapAddr, ethClient)
		if err != nil {
			return nil, err
		}
	}

	return client, nil
}

// Order contains the event order to let the synchronizer store the information following this order.
type Order struct {
	Name EventOrder
	Pos  int
}

// WaitTxToBeMined waits for an L1 tx to be mined. It will return error if the tx is reverted or timeout is exceeded
func (etherMan *Client) WaitTxToBeMined(
	ctx context.Context, tx *types.Transaction, timeout time.Duration,
) (bool, error) {
	// err := operations.WaitTxToBeMined(ctx, etherMan.EthClient, tx, timeout)
	// if errors.Is(err, context.DeadlineExceeded) {
	// 	return false, nil
	// }
	// if err != nil {
	// 	return false, err
	// }
	return true, nil
}

// GetSendSequenceFee get super/trusted sequencer fee
func (etherMan *Client) GetSendSequenceFee(numBatches uint64) (*big.Int, error) {
	f, err := etherMan.Contracts.Banana.RollupManager.GetBatchFee(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, err
	}
	fee := new(big.Int).Mul(f, new(big.Int).SetUint64(numBatches))

	return fee, nil
}

// TrustedSequencer gets trusted sequencer address
func (etherMan *Client) TrustedSequencer() (common.Address, error) {
	return etherMan.Contracts.Banana.Rollup.TrustedSequencer(&bind.CallOpts{Pending: false})
}

// HeaderByNumber returns a block header from the current canonical chain. If number is
// nil, the latest known header is returned.
func (etherMan *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return etherMan.EthClient.HeaderByNumber(ctx, number)
}

// EthBlockByNumber function retrieves the ethereum block information by ethereum block number.
func (etherMan *Client) EthBlockByNumber(ctx context.Context, blockNumber uint64) (*types.Block, error) {
	block, err := etherMan.EthClient.BlockByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	if err != nil {
		if errors.Is(err, ethereum.NotFound) || err.Error() == "block does not exist in blockchain" {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return block, nil
}

// GetLatestBatchNumber function allows to retrieve the latest proposed batch in the smc
func (etherMan *Client) GetLatestBatchNumber() (uint64, error) {
	rollupData, err := etherMan.Contracts.Banana.RollupManager.RollupIDToRollupData(
		&bind.CallOpts{Pending: false},
		etherMan.RollupID,
	)
	if err != nil {
		return 0, err
	}

	return rollupData.LastBatchSequenced, nil
}

// GetLatestBlockNumber gets the latest block number from the ethereum
func (etherMan *Client) GetLatestBlockNumber(ctx context.Context) (uint64, error) {
	return etherMan.getBlockNumber(ctx, rpc.LatestBlockNumber)
}

// GetSafeBlockNumber gets the safe block number from the ethereum
func (etherMan *Client) GetSafeBlockNumber(ctx context.Context) (uint64, error) {
	return etherMan.getBlockNumber(ctx, rpc.SafeBlockNumber)
}

// GetFinalizedBlockNumber gets the Finalized block number from the ethereum
func (etherMan *Client) GetFinalizedBlockNumber(ctx context.Context) (uint64, error) {
	return etherMan.getBlockNumber(ctx, rpc.FinalizedBlockNumber)
}

// getBlockNumber gets the block header by the provided block number from the ethereum
func (etherMan *Client) getBlockNumber(ctx context.Context, blockNumber rpc.BlockNumber) (uint64, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil || header == nil {
		return 0, err
	}

	return header.Number.Uint64(), nil
}

// GetLatestBlockTimestamp gets the latest block timestamp from the ethereum
func (etherMan *Client) GetLatestBlockTimestamp(ctx context.Context) (uint64, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, nil)
	if err != nil || header == nil {
		return 0, err
	}

	return header.Time, nil
}

// GetLatestVerifiedBatchNum gets latest verified batch from ethereum
func (etherMan *Client) GetLatestVerifiedBatchNum() (uint64, error) {
	rollupData, err := etherMan.Contracts.Banana.RollupManager.RollupIDToRollupData(
		&bind.CallOpts{Pending: false},
		etherMan.RollupID,
	)
	if err != nil {
		return 0, err
	}

	return rollupData.LastVerifiedBatch, nil
}

// GetTx function get ethereum tx
func (etherMan *Client) GetTx(ctx context.Context, txHash common.Hash) (*types.Transaction, bool, error) {
	return etherMan.EthClient.TransactionByHash(ctx, txHash)
}

// GetTxReceipt function gets ethereum tx receipt
func (etherMan *Client) GetTxReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	return etherMan.EthClient.TransactionReceipt(ctx, txHash)
}

// GetTrustedSequencerURL Gets the trusted sequencer url from rollup smc
func (etherMan *Client) GetTrustedSequencerURL() (string, error) {
	return etherMan.Contracts.Banana.Rollup.TrustedSequencerURL(&bind.CallOpts{Pending: false})
}

// GetL2ChainID returns L2 Chain ID
func (etherMan *Client) GetL2ChainID() (uint64, error) {
	rollupData, err := etherMan.Contracts.Banana.RollupManager.RollupIDToRollupData(
		&bind.CallOpts{Pending: false},
		etherMan.RollupID,
	)
	log.Debug("chainID read from rollupManager: ", rollupData.ChainID)
	if err != nil {
		log.Debug("error from rollupManager: ", err)

		return 0, err
	} else if rollupData.ChainID == 0 {
		return rollupData.ChainID, fmt.Errorf("error: chainID received is 0")
	}

	return rollupData.ChainID, nil
}

// SendTx sends a tx to L1
func (etherMan *Client) SendTx(ctx context.Context, tx *types.Transaction) error {
	return etherMan.EthClient.SendTransaction(ctx, tx)
}

// CurrentNonce returns the current nonce for the provided account
func (etherMan *Client) CurrentNonce(ctx context.Context, account common.Address) (uint64, error) {
	return etherMan.EthClient.NonceAt(ctx, account, nil)
}

// EstimateGas returns the estimated gas for the tx
func (etherMan *Client) EstimateGas(
	ctx context.Context, from common.Address, to *common.Address, value *big.Int, data []byte,
) (uint64, error) {
	return etherMan.EthClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    to,
		Value: value,
		Data:  data,
	})
}

// CheckTxWasMined check if a tx was already mined
func (etherMan *Client) CheckTxWasMined(ctx context.Context, txHash common.Hash) (bool, *types.Receipt, error) {
	receipt, err := etherMan.EthClient.TransactionReceipt(ctx, txHash)
	if errors.Is(err, ethereum.NotFound) {
		return false, nil, nil
	} else if err != nil {
		return false, nil, err
	}

	return true, receipt, nil
}

// SignTx tries to sign a transaction accordingly to the provided sender
func (etherMan *Client) SignTx(
	ctx context.Context, sender common.Address, tx *types.Transaction,
) (*types.Transaction, error) {
	auth, err := etherMan.getAuthByAddress(sender)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrPrivateKeyNotFound
	}
	signedTx, err := auth.Signer(auth.From, tx)
	if err != nil {
		return nil, err
	}

	return signedTx, nil
}

// GetRevertMessage tries to get a revert message of a transaction
func (etherMan *Client) GetRevertMessage(ctx context.Context, tx *types.Transaction) (string, error) {
	// if tx == nil {
	// 	return "", nil
	// }

	// receipt, err := etherMan.GetTxReceipt(ctx, tx.Hash())
	// if err != nil {
	// 	return "", err
	// }

	// if receipt.Status == types.ReceiptStatusFailed {
	// 	revertMessage, err := operations.RevertReason(ctx, etherMan.EthClient, tx, receipt.BlockNumber)
	// 	if err != nil {
	// 		return "", err
	// 	}
	// 	return revertMessage, nil
	// }
	return "", nil
}

// AddOrReplaceAuth adds an authorization or replace an existent one to the same account
func (etherMan *Client) AddOrReplaceAuth(auth bind.TransactOpts) error {
	log.Infof("added or replaced authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth

	return nil
}

// LoadAuthFromKeyStore loads an authorization from a key store file
func (etherMan *Client) LoadAuthFromKeyStore(path, password string) (*bind.TransactOpts, *ecdsa.PrivateKey, error) {
	auth, pk, err := newAuthFromKeystore(path, password, etherMan.l1Cfg.L1ChainID)
	if err != nil {
		return nil, nil, err
	}

	log.Infof("loaded authorization for address: %v", auth.From.String())
	etherMan.auth[auth.From] = auth

	return &auth, pk, nil
}

// newKeyFromKeystore creates an instance of a keystore key from a keystore file
func newKeyFromKeystore(path, password string) (*keystore.Key, error) {
	if path == "" && password == "" {
		return nil, nil
	}
	keystoreEncrypted, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}
	log.Infof("decrypting key from: %v", path)
	key, err := keystore.DecryptKey(keystoreEncrypted, password)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// newAuthFromKeystore an authorization instance from a keystore file
func newAuthFromKeystore(path, password string, chainID uint64) (bind.TransactOpts, *ecdsa.PrivateKey, error) {
	log.Infof("reading key from: %v", path)
	key, err := newKeyFromKeystore(path, password)
	if err != nil {
		return bind.TransactOpts{}, nil, err
	}
	if key == nil {
		return bind.TransactOpts{}, nil, nil
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key.PrivateKey, new(big.Int).SetUint64(chainID))
	if err != nil {
		return bind.TransactOpts{}, nil, err
	}

	return *auth, key.PrivateKey, nil
}

// getAuthByAddress tries to get an authorization from the authorizations map
func (etherMan *Client) getAuthByAddress(addr common.Address) (bind.TransactOpts, error) {
	auth, found := etherMan.auth[addr]
	if !found {
		return bind.TransactOpts{}, ErrNotFound
	}

	return auth, nil
}

// GetLatestBlockHeader gets the latest block header from the ethereum
func (etherMan *Client) GetLatestBlockHeader(ctx context.Context) (*types.Header, error) {
	header, err := etherMan.EthClient.HeaderByNumber(ctx, big.NewInt(int64(rpc.LatestBlockNumber)))
	if err != nil || header == nil {
		return nil, err
	}

	return header, nil
}

// GetDAProtocolAddr returns the address of the data availability protocol
func (etherMan *Client) GetDAProtocolAddr() (common.Address, error) {
	return etherMan.Contracts.Banana.Rollup.DataAvailabilityProtocol(&bind.CallOpts{Pending: false})
}

// GetDAProtocolName returns the name of the data availability protocol
func (etherMan *Client) GetDAProtocolName() (string, error) {
	return etherMan.DAProtocol.GetProcotolName(&bind.CallOpts{Pending: false})
}

// LastAccInputHash gets the last acc input hash from the SC
func (etherMan *Client) LastAccInputHash() (common.Hash, error) {
	return etherMan.Contracts.Banana.Rollup.LastAccInputHash(&bind.CallOpts{Pending: false})
}

// GetL1InfoRoot gets the L1 info root from the SC
func (etherMan *Client) GetL1InfoRoot(indexL1InfoRoot uint32) (common.Hash, error) {
	// Get lastL1InfoTreeRoot (if index==0 then root=0, no call is needed)
	var (
		lastL1InfoTreeRoot common.Hash
		err                error
	)

	if indexL1InfoRoot > 0 {
		lastL1InfoTreeRoot, err = etherMan.Contracts.Banana.GlobalExitRoot.L1InfoRootMap(
			&bind.CallOpts{Pending: false},
			indexL1InfoRoot,
		)
		if err != nil {
			log.Errorf("error calling SC globalexitroot L1InfoLeafMap: %v", err)
		}
	}

	return lastL1InfoTreeRoot, err
}
