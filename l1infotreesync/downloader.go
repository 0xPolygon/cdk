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

func checkSMCIsRollupManager(rollupManagerAddr common.Address,
	rollupManagerContract *polygonrollupmanager.Polygonrollupmanager) error {
	bridgeAddr, err := rollupManagerContract.BridgeAddress(nil)
	if err != nil {
		return fmt.Errorf("fail sanity check RollupManager(%s) Contract. Err: %w", rollupManagerAddr.String(), err)
	}
	log.Infof("sanity check rollupManager(%s) OK. bridgeAddr: %s", rollupManagerAddr.String(), bridgeAddr.String())
	return nil
}

func checkSMCIsGlobalExitRoot(globalExitRootAddr common.Address,
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2) error {
	depositCount, err := gerContract.DepositCount(nil)
	if err != nil {
		return fmt.Errorf("fail sanity check GlobalExitRoot(%s) Contract. Err: %w", globalExitRootAddr.String(), err)
	}
	log.Infof("sanity check GlobalExitRoot(%s) OK. DepositCount: %v", globalExitRootAddr.String(), depositCount)
	return nil
}

func sanityCheckContracts(globalExitRoot, rollupManager common.Address,
	gerContract *polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	rollupManagerContract *polygonrollupmanager.Polygonrollupmanager) error {
	errGER := checkSMCIsGlobalExitRoot(globalExitRoot, gerContract)
	errRollup := checkSMCIsRollupManager(rollupManager, rollupManagerContract)
	if errGER != nil || errRollup != nil {
		err := fmt.Errorf("sanityCheckContracts: fails sanity check contracts. ErrGER: %w, ErrRollup: %w", errGER, errRollup)
		log.Error(err)
		return err
	}
	return nil
}

func createContracts(client EthClienter, globalExitRoot, rollupManager common.Address) (
	*polygonzkevmglobalexitrootv2.Polygonzkevmglobalexitrootv2,
	*polygonrollupmanager.Polygonrollupmanager,
	error) {
	gerContract, err := polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2(globalExitRoot, client)
	if err != nil {
		return nil, nil, err
	}

	rollupManagerContract, err := polygonrollupmanager.NewPolygonrollupmanager(rollupManager, client)
	if err != nil {
		return nil, nil, err
	}
	return gerContract, rollupManagerContract, nil
}

func buildAppender(client EthClienter, globalExitRoot,
	rollupManager common.Address, flags CreationFlags) (sync.LogAppenderMap, error) {
	ger, rm, err := createContracts(client, globalExitRoot, rollupManager)
	if err != nil {
		err := fmt.Errorf("buildAppender: fails contracts creation. Err:%w", err)
		log.Error(err)
		return nil, err
	}
	err = sanityCheckContracts(globalExitRoot, rollupManager, ger, rm)
	if err != nil && flags&FlagAllowWrongContractsAddrs == 0 {
		return nil, fmt.Errorf("buildAppender: fails sanity check contracts. Err:%w", err)
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
			BlockPosition:   uint64(l.Index),
			MainnetExitRoot: l1InfoTreeUpdate.MainnetExitRoot,
			RollupExitRoot:  l1InfoTreeUpdate.RollupExitRoot,
			ParentHash:      b.ParentHash,
			Timestamp:       b.Timestamp,
		}})

		return nil
	}

	appender[updateL1InfoTreeSignatureV2] = func(b *sync.EVMBlock, l types.Log) error {
		l1InfoTreeUpdateV2, err := ger.ParseUpdateL1InfoTreeV2(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using ger.ParseUpdateL1InfoTreeV2: %w",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{UpdateL1InfoTreeV2: &UpdateL1InfoTreeV2{
			CurrentL1InfoRoot: l1InfoTreeUpdateV2.CurrentL1InfoRoot,
			LeafCount:         l1InfoTreeUpdateV2.LeafCount,
			Blockhash:         common.BytesToHash(l1InfoTreeUpdateV2.Blockhash.Bytes()),
			MinTimestamp:      l1InfoTreeUpdateV2.MinTimestamp,
		}})

		return nil
	}
	// This event is coming from RollupManager
	appender[verifyBatchesSignature] = func(b *sync.EVMBlock, l types.Log) error {
		verifyBatches, err := rm.ParseVerifyBatches(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using rm.ParseVerifyBatches: %w",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{VerifyBatches: &VerifyBatches{
			BlockPosition: uint64(l.Index),
			RollupID:      verifyBatches.RollupID,
			NumBatch:      verifyBatches.NumBatch,
			StateRoot:     verifyBatches.StateRoot,
			ExitRoot:      verifyBatches.ExitRoot,
			Aggregator:    verifyBatches.Aggregator,
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
			BlockPosition: uint64(l.Index),
			RollupID:      verifyBatches.RollupID,
			NumBatch:      verifyBatches.NumBatch,
			StateRoot:     verifyBatches.StateRoot,
			ExitRoot:      verifyBatches.ExitRoot,
			Aggregator:    verifyBatches.Aggregator,
		}})

		return nil
	}

	return appender, nil
}
