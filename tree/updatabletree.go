package tree

import (
	"context"
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

type UpdatableTree struct {
	*Tree
}

func NewUpdatable(ctx context.Context, dbPath, dbPrefix string) (*UpdatableTree, error) {
	t, err := newTree(dbPath, dbPrefix)
	if err != nil {
		return nil, err
	}
	ut := &UpdatableTree{Tree: t}
	return ut, nil
}

func (t *UpdatableTree) UpsertLeaf(ctx context.Context, index uint32, leafHash common.Hash, expectedRoot *common.Hash) error {
	return errors.New("not implemented")
}
