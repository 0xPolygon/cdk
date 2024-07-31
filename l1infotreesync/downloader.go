package l1infotreesync

import (
	"fmt"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonrollupmanager"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	updateL1InfoTreeSignature               = crypto.Keccak256Hash([]byte("UpdateL1InfoTree(bytes32,bytes32)"))
	verifyBatchesSignature                  = crypto.Keccak256Hash([]byte("VerifyBatches(uint32,uint64,bytes32,bytes32,address)"))
	verifyBatchesTrustedAggregatorSignature = crypto.Keccak256Hash([]byte("VerifyBatchesTrustedAggregator(uint32,uint64,bytes32,bytes32,address)"))
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

func buildAppender(client EthClienter, globalExitRoot, rollupManager common.Address) (sync.LogAppenderMap, error) {
	ger, err := polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2(globalExitRoot, client)
	if err != nil {
		return nil, err
	}
	rm, err := polygonrollupmanager.NewPolygonrollupmanager(rollupManager, client)
	if err != nil {
		return nil, err
	}
	appender := make(sync.LogAppenderMap)
	appender[updateL1InfoTreeSignature] = func(b *sync.EVMBlock, l types.Log) error {
		l1InfoTreeUpdate, err := ger.ParseUpdateL1InfoTree(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using ger.ParseUpdateL1InfoTree: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{UpdateL1InfoTree: &UpdateL1InfoTree{
			MainnetExitRoot: l1InfoTreeUpdate.MainnetExitRoot,
			RollupExitRoot:  l1InfoTreeUpdate.RollupExitRoot,
			ParentHash:      b.ParentHash,
			Timestamp:       b.Timestamp,
		}})
		return nil
	}
	appender[verifyBatchesSignature] = func(b *sync.EVMBlock, l types.Log) error {
		verifyBatches, err := rm.ParseVerifyBatches(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using rm.ParseVerifyBatches: %v",
				l, err,
			)
		}
		fmt.Println(verifyBatches)
		b.Events = append(b.Events, Event{VerifyBatches: &VerifyBatches{
			RollupID:   verifyBatches.RollupID,
			NumBatch:   verifyBatches.NumBatch,
			StateRoot:  verifyBatches.StateRoot,
			ExitRoot:   verifyBatches.ExitRoot,
			Aggregator: verifyBatches.Aggregator,
		}})
		return nil
	}
	appender[verifyBatchesTrustedAggregatorSignature] = func(b *sync.EVMBlock, l types.Log) error {
		verifyBatches, err := rm.ParseVerifyBatchesTrustedAggregator(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using rm.ParseVerifyBatches: %v",
				l, err,
			)
		}
		fmt.Println(verifyBatches)
		b.Events = append(b.Events, Event{VerifyBatches: &VerifyBatches{
			RollupID:   verifyBatches.RollupID,
			NumBatch:   verifyBatches.NumBatch,
			StateRoot:  verifyBatches.StateRoot,
			ExitRoot:   verifyBatches.ExitRoot,
			Aggregator: verifyBatches.Aggregator,
		}})
		return nil
	}

	return appender, nil
}
