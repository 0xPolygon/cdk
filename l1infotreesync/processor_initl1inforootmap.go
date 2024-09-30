package l1infotreesync

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/russross/meddler"
)

func processEventInitL1InfoRootMap(tx db.Txer, blockNumber uint64, event *InitL1InfoRootMap) error {
	if event == nil {
		return nil
	}
	info := &L1InfoTreeInitial{
		BlockNumber: blockNumber,
		LeafCount:   event.LeafCount,
		L1InfoRoot:  event.CurrentL1InfoRoot,
	}
	log.Infof("insert InitL1InfoRootMap %s ", info.String())
	if err := meddler.Insert(tx, "l1info_initial", info); err != nil {
		return fmt.Errorf("err: %w", err)
	}
	return nil
}

// GetInitL1InfoRootMap returns the initial L1 info root map, nil if no root map has been set
func (p *processor) GetInitL1InfoRootMap(tx db.Txer) (*L1InfoTreeInitial, error) {
	info := &L1InfoTreeInitial{}
	err := meddler.QueryRow(p.getDBQuerier(tx), info, `SELECT block_num, leaf_count,l1_info_root  FROM l1info_initial`)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return info, err
}
