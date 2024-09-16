package claimsponsor

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"math/big"
	"time"

	dbCommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/iter"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

type ClaimStatus string

const (
	PendingClaimStatus = "pending"
	WIPStatus          = "work in progress"
	SuccessClaimStatus = "success"
	FailedClaimStatus  = "failed"

	claimTable = "claimsponsor-tx"
	queueTable = "claimsponsor-queue"
)

var (
	ErrInvalidClaim = errors.New("invalid claim")
	ErrNotFound     = errors.New("not found")
)

// Claim representation of a claim event
type Claim struct {
	LeafType            uint8
	ProofLocalExitRoot  [32]common.Hash
	ProofRollupExitRoot [32]common.Hash
	GlobalIndex         *big.Int
	MainnetExitRoot     common.Hash
	RollupExitRoot      common.Hash
	OriginNetwork       uint32
	OriginTokenAddress  common.Address
	DestinationNetwork  uint32
	DestinationAddress  common.Address
	Amount              *big.Int
	Metadata            []byte

	Status ClaimStatus
	TxID   string
}

func (c *Claim) Key() []byte {
	return c.GlobalIndex.Bytes()
}

type ClaimSender interface {
	checkClaim(ctx context.Context, claim *Claim) error
	sendClaim(ctx context.Context, claim *Claim) (string, error)
	claimStatus(ctx context.Context, id string) (ClaimStatus, error)
}

type ClaimSponsor struct {
	logger                *log.Logger
	db                    kv.RwDB
	sender                ClaimSender
	rh                    *sync.RetryHandler
	waitTxToBeMinedPeriod time.Duration
	waitOnEmptyQueue      time.Duration
}

func newClaimSponsor(
	logger *log.Logger,
	dbPath string,
	sender ClaimSender,
	retryAfterErrorPeriod time.Duration,
	maxRetryAttemptsAfterError int,
	waitTxToBeMinedPeriod time.Duration,
	waitOnEmptyQueue time.Duration,
) (*ClaimSponsor, error) {
	tableCfgFunc := func(defaultBuckets kv.TableCfg) kv.TableCfg {
		cfg := kv.TableCfg{
			claimTable: {},
			queueTable: {},
		}

		return cfg
	}
	db, err := mdbx.NewMDBX(nil).
		Path(dbPath).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, err
	}
	rh := &sync.RetryHandler{
		MaxRetryAttemptsAfterError: maxRetryAttemptsAfterError,
		RetryAfterErrorPeriod:      retryAfterErrorPeriod,
	}

	return &ClaimSponsor{
		logger:                logger,
		db:                    db,
		sender:                sender,
		rh:                    rh,
		waitTxToBeMinedPeriod: waitTxToBeMinedPeriod,
		waitOnEmptyQueue:      waitOnEmptyQueue,
	}, nil
}

func (c *ClaimSponsor) Start(ctx context.Context) {
	var (
		attempts int
		err      error
	)
	for {
		if err != nil {
			attempts++
			c.rh.Handle("claimsponsor main loop", attempts)
		}
		tx, err2 := c.db.BeginRw(ctx)
		if err2 != nil {
			err = err2
			c.logger.Errorf("error calling BeginRw: %v", err)
			continue
		}
		queueIndex, globalIndex, err2 := getFirstQueueIndex(tx)
		if err2 != nil {
			err = err2
			tx.Rollback()
			if errors.Is(err, ErrNotFound) {
				c.logger.Debugf("queue is empty")
				err = nil
				time.Sleep(c.waitOnEmptyQueue)

				continue
			}
			c.logger.Errorf("error calling getFirstQueueIndex: %v", err)
			continue
		}
		claim, err2 := getClaim(tx, globalIndex)
		if err2 != nil {
			err = err2
			tx.Rollback()
			c.logger.Errorf("error calling getClaim with globalIndex %s: %v", globalIndex.String(), err)
			continue
		}
		if claim.TxID == "" {
			txID, err2 := c.sender.sendClaim(ctx, claim)
			if err2 != nil {
				err = err2
				tx.Rollback()
				c.logger.Errorf("error calling sendClaim with globalIndex %s: %v", globalIndex.String(), err)
				continue
			}
			claim.TxID = txID
			claim.Status = WIPStatus
			err2 = putClaim(tx, claim)
			if err2 != nil {
				err = err2
				tx.Rollback()
				c.logger.Errorf("error calling putClaim with globalIndex %s: %v", globalIndex.String(), err)
				continue
			}
		}
		err2 = tx.Commit()
		if err2 != nil {
			err = err2
			c.logger.Errorf("error calling tx.Commit after putting claim: %v", err)
			continue
		}

		c.logger.Infof("waiting for tx %s with global index %s to succeed or fail", claim.TxID, globalIndex.String())
		status, err2 := c.waitTxToBeSuccessOrFail(ctx, claim.TxID)
		if err2 != nil {
			err = err2
			c.logger.Errorf("error calling waitTxToBeSuccessOrFail for tx %s: %v", claim.TxID, err)
			continue
		}
		c.logger.Infof("tx %s with global index %s concluded with status: %s", claim.TxID, globalIndex.String(), status)
		tx, err2 = c.db.BeginRw(ctx)
		if err2 != nil {
			err = err2
			c.logger.Errorf("error calling BeginRw: %v", err)
			continue
		}
		claim.Status = status
		err2 = putClaim(tx, claim)
		if err2 != nil {
			err = err2
			tx.Rollback()
			c.logger.Errorf("error calling putClaim with globalIndex %s: %v", globalIndex.String(), err)
			continue
		}
		err2 = tx.Delete(queueTable, dbCommon.Uint64ToBytes(queueIndex))
		if err2 != nil {
			err = err2
			tx.Rollback()
			c.logger.Errorf("error calling delete on the queue table with index %d: %v", queueIndex, err)
			continue
		}
		err2 = tx.Commit()
		if err2 != nil {
			err = err2
			c.logger.Errorf("error calling tx.Commit after putting claim: %v", err)
			continue
		}

		attempts = 0
	}
}

func (c *ClaimSponsor) waitTxToBeSuccessOrFail(ctx context.Context, txID string) (ClaimStatus, error) {
	t := time.NewTicker(c.waitTxToBeMinedPeriod)
	for {
		select {
		case <-ctx.Done():
			return "", errors.New("context cancelled")
		case <-t.C:
			status, err := c.sender.claimStatus(ctx, txID)
			if err != nil {
				return "", err
			}
			if status == FailedClaimStatus || status == SuccessClaimStatus {
				return status, nil
			}
		}
	}
}

func (c *ClaimSponsor) AddClaimToQueue(ctx context.Context, claim *Claim) error {
	if claim.GlobalIndex == nil {
		return ErrInvalidClaim
	}
	claim.Status = PendingClaimStatus
	tx, err := c.db.BeginRw(ctx)
	if err != nil {
		return err
	}

	_, err = getClaim(tx, claim.GlobalIndex)
	if !errors.Is(err, ErrNotFound) {
		if err != nil {
			tx.Rollback()

			return err
		} else {
			tx.Rollback()

			return errors.New("claim already added")
		}
	}

	err = putClaim(tx, claim)
	if err != nil {
		tx.Rollback()

		return err
	}

	var queuePosition uint64
	lastQueuePosition, _, err := getLastQueueIndex(tx)
	if errors.Is(err, ErrNotFound) {
		queuePosition = 0
	} else if err != nil {
		tx.Rollback()

		return err
	} else {
		queuePosition = lastQueuePosition + 1
	}
	err = tx.Put(queueTable, dbCommon.Uint64ToBytes(queuePosition), claim.Key())
	if err != nil {
		tx.Rollback()

		return err
	}

	return tx.Commit()
}

func putClaim(tx kv.RwTx, claim *Claim) error {
	value, err := json.Marshal(claim)
	if err != nil {
		return err
	}

	return tx.Put(claimTable, claim.Key(), value)
}

func (c *ClaimSponsor) getClaimByQueueIndex(ctx context.Context, queueIndex uint64) (*Claim, error) {
	tx, err := c.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	globalIndexBytes, err := tx.GetOne(queueTable, dbCommon.Uint64ToBytes(queueIndex))
	if err != nil {
		return nil, err
	}
	if globalIndexBytes == nil {
		return nil, ErrNotFound
	}

	return getClaim(tx, new(big.Int).SetBytes(globalIndexBytes))
}

func getLastQueueIndex(tx kv.Tx) (uint64, *big.Int, error) {
	iter, err := tx.RangeDescend(
		queueTable,
		dbCommon.Uint64ToBytes(math.MaxUint64),
		dbCommon.Uint64ToBytes(0), 1,
	)
	if err != nil {
		return 0, nil, err
	}

	return getIndex(iter)
}

func getFirstQueueIndex(tx kv.Tx) (uint64, *big.Int, error) {
	iter, err := tx.RangeAscend(
		queueTable,
		dbCommon.Uint64ToBytes(0),
		nil, 1,
	)
	if err != nil {
		return 0, nil, err
	}

	return getIndex(iter)
}

func getIndex(iter iter.KV) (uint64, *big.Int, error) {
	k, v, err := iter.Next()
	if err != nil {
		return 0, nil, err
	}
	if k == nil {
		return 0, nil, ErrNotFound
	}
	globalIndex := new(big.Int).SetBytes(v)

	return dbCommon.BytesToUint64(k), globalIndex, nil
}

func (c *ClaimSponsor) GetClaim(ctx context.Context, globalIndex *big.Int) (*Claim, error) {
	tx, err := c.db.BeginRo(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	return getClaim(tx, globalIndex)
}

func getClaim(tx kv.Tx, globalIndex *big.Int) (*Claim, error) {
	claimBytes, err := tx.GetOne(claimTable, globalIndex.Bytes())
	if err != nil {
		return nil, err
	}
	if claimBytes == nil {
		return nil, ErrNotFound
	}
	claim := &Claim{}
	err = json.Unmarshal(claimBytes, claim)

	return claim, err
}
