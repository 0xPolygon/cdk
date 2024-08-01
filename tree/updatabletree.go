package tree

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
)

type UpdatableTree struct {
	*Tree
}

func NewUpdatable(ctx context.Context, db kv.RwDB, dbPrefix string) *UpdatableTree {
	t := newTree(db, dbPrefix)
	ut := &UpdatableTree{Tree: t}
	return ut
}

func (t *UpdatableTree) UpsertLeaf(tx kv.RwTx, index uint32, leafHash common.Hash, expectedRoot *common.Hash) (func(), error) {
	return func() {}, errors.New("not implemented")
}

func (t *UpdatableTree) Reorg(tx kv.RwTx, firstReorgedIndex uint32) (func(), error) {
	return func() {}, errors.New("not implemented")
}
