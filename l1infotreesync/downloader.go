package l1infotreesync

import (
	"fmt"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/elderberry/polygonzkevmglobalexitrootv2"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	updateL1InfoTreeSignature = crypto.Keccak256Hash([]byte("UpdateL1InfoTree(bytes32,bytes32)"))
)

type EthClienter interface {
	ethereum.LogFilterer
	ethereum.BlockNumberReader
	ethereum.ChainReader
	bind.ContractBackend
}

func buildAppender(client EthClienter, globalExitRoot common.Address) (sync.LogAppenderMap, error) {
	contract, err := polygonzkevmglobalexitrootv2.NewPolygonzkevmglobalexitrootv2(globalExitRoot, client)
	if err != nil {
		return nil, err
	}
	appender := make(sync.LogAppenderMap)
	appender[updateL1InfoTreeSignature] = func(b *sync.EVMBlock, l types.Log) error {
		l1InfoTreeUpdate, err := contract.ParseUpdateL1InfoTree(l)
		if err != nil {
			return fmt.Errorf(
				"error parsing log %+v using contract.ParseUpdateL1InfoTree: %v",
				l, err,
			)
		}
		b.Events = append(b.Events, Event{
			MainnetExitRoot: l1InfoTreeUpdate.MainnetExitRoot,
			RollupExitRoot:  l1InfoTreeUpdate.RollupExitRoot,
			ParentHash:      b.ParentHash,
			Timestamp:       b.Timestamp,
		})
		return nil
	}
	return appender, nil
}
