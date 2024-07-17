package etherman

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethmanTypes "github.com/0xPolygon/cdk/aggregator/ethmantypes"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// BuildTrustedVerifyBatchesTxData builds a []bytes to be sent to the PoE SC method TrustedVerifyBatches.
func (etherMan *Client) BuildTrustedVerifyBatchesTxData(lastVerifiedBatch, newVerifiedBatch uint64, inputs *ethmanTypes.FinalProofInputs, beneficiary common.Address) (to *common.Address, data []byte, err error) {
	opts, err := etherMan.generateRandomAuth()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build trusted verify batches, err: %w", err)
	}
	opts.NoSend = true
	// force nonce, gas limit and gas price to avoid querying it from the chain
	opts.Nonce = big.NewInt(1)
	opts.GasLimit = uint64(1)
	opts.GasPrice = big.NewInt(1)

	var newLocalExitRoot [32]byte
	copy(newLocalExitRoot[:], inputs.NewLocalExitRoot)

	var newStateRoot [32]byte
	copy(newStateRoot[:], inputs.NewStateRoot)

	proof, err := convertProof(inputs.FinalProof.Proof)
	if err != nil {
		log.Errorf("error converting proof. Error: %v, Proof: %s", err, inputs.FinalProof.Proof)
		return nil, nil, err
	}

	const pendStateNum = 0 // TODO hardcoded for now until we implement the pending state feature

	tx, err := etherMan.Contracts.RollupManager(nil).VerifyBatchesTrustedAggregator(
		&opts,
		etherMan.RollupID,
		pendStateNum,
		lastVerifiedBatch,
		newVerifiedBatch,
		newLocalExitRoot,
		newStateRoot,
		beneficiary,
		proof,
	)
	if err != nil {
		if parsedErr, ok := TryParseError(err); ok {
			err = parsedErr
		}
		return nil, nil, err
	}

	return tx.To(), tx.Data(), nil
}

// GetBatchAccInputHash gets the batch accumulated input hash from the ethereum
func (etherman *Client) GetBatchAccInputHash(ctx context.Context, batchNumber uint64) (common.Hash, error) {
	rollupData, err := etherman.Contracts.RollupManager(nil).GetRollupSequencedBatches(etherman.RollupID, batchNumber)
	if err != nil {
		return common.Hash{}, err
	}
	return rollupData.AccInputHash, nil
}

// GetRollupId returns the rollup id
func (etherMan *Client) GetRollupId() uint32 {
	return etherMan.RollupID
}

// generateRandomAuth generates an authorization instance from a
// randomly generated private key to be used to estimate gas for PoE
// operations NOT restricted to the Trusted Sequencer
func (etherMan *Client) generateRandomAuth() (bind.TransactOpts, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return bind.TransactOpts{}, errors.New("failed to generate a private key to estimate L1 txs")
	}
	chainID := big.NewInt(0).SetUint64(etherMan.l1Cfg.L1ChainID)
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		return bind.TransactOpts{}, errors.New("failed to generate a fake authorization to estimate L1 txs")
	}

	return *auth, nil
}

func convertProof(p string) ([24][32]byte, error) {
	if len(p) != 24*32*2+2 {
		return [24][32]byte{}, fmt.Errorf("invalid proof length. Length: %d", len(p))
	}
	p = strings.TrimPrefix(p, "0x")
	proof := [24][32]byte{}
	for i := 0; i < 24; i++ {
		data := p[i*64 : (i+1)*64]
		p, err := DecodeBytes(&data)
		if err != nil {
			return [24][32]byte{}, fmt.Errorf("failed to decode proof, err: %w", err)
		}
		var aux [32]byte
		copy(aux[:], p)
		proof[i] = aux
	}
	return proof, nil
}

// DecodeBytes decodes a hex string into a []byte
func DecodeBytes(val *string) ([]byte, error) {
	if val == nil {
		return []byte{}, nil
	}
	return hex.DecodeString(strings.TrimPrefix(*val, "0x"))
}
