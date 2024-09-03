package l1infotreesync

import (
	"fmt"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk-contracts-tooling/contracts/etrog/polygonrollupmanager"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	updateL1InfoTreeSignatureV1 = crypto.Keccak256Hash([]byte("UpdateL1InfoTree(bytes32,bytes32)"))
	updateL1InfoTreeSignatureV2 = crypto.Keccak256Hash([]byte("UpdateL1InfoTreeV2(bytes32,uint32,uint256,uint64)"))
	verifyBatchesSignature      = crypto.Keccak256Hash(
		[]byte("VerifyBatches(uint32,uint64,bytes32,bytes32,address)"),
	)
	verifyBatchesTrustedAggregatorSignature = crypto.Keccak256Hash(
		[]byte("VerifyBatchesTrustedAggregator(uint32,uint64,bytes32,bytes32,address)"),
	)
	initL1InfoRootMapSignature = crypto.Keccak256Hash([]byte("InitL1InfoRootMap(uint32,bytes32)"))
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
	appender[initL1InfoRootMapSignature] = func(b *sync.EVMBlock, l types.Log) error {
		init, err := ger.ParseInitL1InfoRootMap(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using ger.ParseInitL1InfoRootMap: %w",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{InitL1InfoRootMap: &InitL1InfoRootMap{
			LeafCount:         init.LeafCount,
			CurrentL1InfoRoot: init.CurrentL1InfoRoot,
		}})

		return nil
	}
	appender[updateL1InfoTreeSignatureV1] = func(b *sync.EVMBlock, l types.Log) error {
		l1InfoTreeUpdate, err := ger.ParseUpdateL1InfoTree(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using ger.ParseUpdateL1InfoTree: %w",
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

	// TODO: integrate this event to perform sanity checks
	appender[updateL1InfoTreeSignatureV2] = func(b *sync.EVMBlock, l types.Log) error { //nolint:unparam
		l1InfoTreeUpdate, err := ger.ParseUpdateL1InfoTreeV2(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using ger.ParseUpdateL1InfoTreeV2: %w",
				l, err,
			)
		}
		log.Infof("updateL1InfoTreeSignatureV2: expected root: %s", common.Bytes2Hex(l1InfoTreeUpdate.CurrentL1InfoRoot[:]))

		return nil
	}
	appender[verifyBatchesSignature] = func(b *sync.EVMBlock, l types.Log) error {
		verifyBatches, err := rm.ParseVerifyBatches(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using rm.ParseVerifyBatches: %w",
				l, err,
			)
		}
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
				"error parsing log %+v using rm.ParseVerifyBatches: %w",
				l, err,
			)
		}
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
