package l1bridge2infoindexsync

import (
	"context"
	"math/big"

	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
)

type downloader struct {
	l1Bridge *bridgesync.BridgeSync
	l1Info   *l1infotreesync.L1InfoTreeSync
	l1Client ethereum.ChainReader
}

func newDownloader(
	l1Bridge *bridgesync.BridgeSync,
	l1Info *l1infotreesync.L1InfoTreeSync,
	l1Client ethereum.ChainReader,
) *downloader {
	return &downloader{
		l1Bridge: l1Bridge,
		l1Info:   l1Info,
		l1Client: l1Client,
	}
}

func (d *downloader) getLastFinalizedL1Block(ctx context.Context) (uint64, error) {
	b, err := d.l1Client.BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
	if err != nil {
		return 0, err
	}
	return b.NumberU64(), nil
}

func (d *downloader) getLastProcessedBlockBridge(ctx context.Context) (uint64, error) {
	return d.l1Bridge.GetLastProcessedBlock(ctx)
}

func (d *downloader) getLastProcessedBlockL1InfoTree(ctx context.Context) (uint64, error) {
	return d.l1Info.GetLastProcessedBlock(ctx)
}

func (d *downloader) getLastL1InfoIndexUntilBlock(ctx context.Context, blockNum uint64) (uint32, error) {
	info, err := d.l1Info.GetLatestInfoUntilBlock(ctx, blockNum)
	if err != nil {
		return 0, err
	}
	return info.L1InfoTreeIndex, nil
}

func (d *downloader) getMainnetExitRootAtL1InfoTreeIndex(ctx context.Context, index uint32) (common.Hash, error) {
	leaf, err := d.l1Info.GetInfoByIndex(ctx, index)
	if err != nil {
		return common.Hash{}, err
	}
	return leaf.MainnetExitRoot, nil
}

func (d *downloader) getBridgeIndex(ctx context.Context, mainnetExitRoot common.Hash) (uint32, error) {
	return d.l1Bridge.GetBridgeIndexByRoot(ctx, mainnetExitRoot)
}
