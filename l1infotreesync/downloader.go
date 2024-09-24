package l1infotreesync

import (
	"context"
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

func checkAddrIsContract(client EthClienter, addr common.Address) error {
	code, err := client.CodeAt(context.Background(), addr, nil)
	if err != nil {
		return err
	}
	if len(code) == 0 {
		return fmt.Errorf("address %s is not a contract", addr.String())
	}
	return nil
}

func checkSMCIsRollupManager(rollupManagerAddr common.Address,
	rollupManagerContract *polygonrollupmanager.Polygonrollupmanager) error {
	bridgeAddr, err := rollupManagerContract.BridgeAddress(nil)
	if err != nil {
		return fmt.Errorf("fail sanity check RollupManager(%s) Contract. Err: %w", rollupManagerAddr.String(), err)
	}
	log.Debugf("sanity check rollupManager (%s) OK. bridgeAddr: %s", rollupManagerAddr.String(), bridgeAddr.String())
	return nil
}

func createContracts(client EthClienter, globalExitRoot, rollupManager common.Address, doSanityCheck bool) (
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
	if doSanityCheck {
		depositCount, err := gerContract.DepositCount(nil)
		if err != nil {
			return nil, nil, fmt.Errorf("fail sanity check GlobalExitRoot(%s) Contract. Err: %w", globalExitRoot.String(), err)
		}
		log.Debugf("sanity check GlobalExitRoot OK. DepositCount: %v", depositCount)
		zeroAddr := common.Address{}
		if rollupManager != zeroAddr {
			err := checkSMCIsRollupManager(rollupManager, rollupManagerContract)
			if err != nil {
				log.Warnf("fail sanity check RollupManager(%s) Contract. Err: %w", rollupManager.String(), err)
				err = checkAddrIsContract(client, rollupManager)
				if err != nil {
					return nil, nil, fmt.Errorf("fail sanity check RollupManager(%s) "+
						"Contract addr doesnt contain any contract.  Err: %w", rollupManager.String(), err)
				}
				log.Warnf("RollupManager(%s) is not the expected RollupManager but it is a contract", rollupManager.String())
			}
		} else {
			log.Warnf("RollupManager contract addr not set. Skipping sanity check. No VerifyBatches events expected")
		}
	}
	return gerContract, rollupManagerContract, nil
}

func buildAppender(client EthClienter, globalExitRoot, rollupManager common.Address) (sync.LogAppenderMap, error) {
	ger, rm, err := createContracts(client, globalExitRoot, rollupManager, true)
	if err != nil {
		return nil, fmt.Errorf("buildAppender: fails contracts creation. Err:%w", err)
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
