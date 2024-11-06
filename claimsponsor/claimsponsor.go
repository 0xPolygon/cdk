package claimsponsor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/0xPolygon/cdk/claimsponsor/migrations"
	"github.com/0xPolygon/cdk/db"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sync"
	tree "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/russross/meddler"
)

type ClaimStatus string

const (
	PendingClaimStatus ClaimStatus = "pending"
	WIPClaimStatus     ClaimStatus = "work in progress"
	SuccessClaimStatus ClaimStatus = "success"
	FailedClaimStatus  ClaimStatus = "failed"
)

var (
	ErrInvalidClaim     = errors.New("invalid claim")
	ErrClaimDoesntExist = errors.New("the claim requested to be updated does not exist")
)

// Claim representation of a claim event
type Claim struct {
	LeafType            uint8          `meddler:"leaf_type"`
	ProofLocalExitRoot  tree.Proof     `meddler:"proof_local_exit_root,merkleproof"`
	ProofRollupExitRoot tree.Proof     `meddler:"proof_rollup_exit_root,merkleproof"`
	GlobalIndex         *big.Int       `meddler:"global_index,bigint"`
	MainnetExitRoot     common.Hash    `meddler:"mainnet_exit_root,hash"`
	RollupExitRoot      common.Hash    `meddler:"rollup_exit_root,hash"`
	OriginNetwork       uint32         `meddler:"origin_network"`
	OriginTokenAddress  common.Address `meddler:"origin_token_address,address"`
	DestinationNetwork  uint32         `meddler:"destination_network"`
	DestinationAddress  common.Address `meddler:"destination_address,address"`
	Amount              *big.Int       `meddler:"amount,bigint"`
	Metadata            []byte         `meddler:"metadata"`
	Status              ClaimStatus    `meddler:"status"`
	TxID                string         `meddler:"tx_id"`
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
	db                    *sql.DB
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
	err := migrations.RunMigrations(dbPath)
	if err != nil {
		return nil, err
	}
	db, err := db.NewSQLiteDB(dbPath)
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
	)
	for {
		err := c.claim(ctx)
		if err != nil {
			attempts++
			c.logger.Error(err)
			c.rh.Handle("claimsponsor main loop", attempts)
		} else {
			attempts = 0
		}
	}
}

func (c *ClaimSponsor) claim(ctx context.Context) error {
	claim, err := c.getWIPClaim()
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return fmt.Errorf("error getting WIP claim: %w", err)
	}
	if errors.Is(err, db.ErrNotFound) || claim == nil {
		// there is no WIP claim, go for the next pending claim
		claim, err = c.getFirstPendingClaim()
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				c.logger.Debugf("queue is empty")
				time.Sleep(c.waitOnEmptyQueue)
				return nil
			}
			return fmt.Errorf("error calling getClaim with globalIndex %s: %w", claim.GlobalIndex.String(), err)
		}
		txID, err := c.sender.sendClaim(ctx, claim)
		if err != nil {
			return fmt.Errorf("error getting sending claim: %w", err)
		}
		if err := c.updateClaimTxID(claim.GlobalIndex, txID); err != nil {
			return fmt.Errorf("error updating claim txID: %w", err)
		}
	}

	c.logger.Infof("waiting for tx %s with global index %s to succeed or fail", claim.TxID, claim.GlobalIndex.String())
	status, err := c.waitTxToBeSuccessOrFail(ctx, claim.TxID)
	if err != nil {
		return fmt.Errorf("error calling waitTxToBeSuccessOrFail for tx %s: %w", claim.TxID, err)
	}
	c.logger.Infof("tx %s with global index %s concluded with status: %s", claim.TxID, claim.GlobalIndex.String(), status)
	return c.updateClaimStatus(claim.GlobalIndex, status)
}

func (c *ClaimSponsor) getWIPClaim() (*Claim, error) {
	claim := &Claim{}
	err := meddler.QueryRow(
		c.db, claim,
		`SELECT * FROM claim WHERE status = $1 ORDER BY rowid ASC LIMIT 1;`,
		WIPClaimStatus,
	)
	return claim, db.ReturnErrNotFound(err)
}

func (c *ClaimSponsor) getFirstPendingClaim() (*Claim, error) {
	claim := &Claim{}
	err := meddler.QueryRow(
		c.db, claim,
		`SELECT * FROM claim WHERE status = $1 ORDER BY rowid ASC LIMIT 1;`,
		PendingClaimStatus,
	)
	return claim, db.ReturnErrNotFound(err)
}

func (c *ClaimSponsor) updateClaimTxID(globalIndex *big.Int, txID string) error {
	res, err := c.db.Exec(
		`UPDATE claim SET tx_id = $1 WHERE global_index = $2`,
		txID, globalIndex.String(),
	)
	if err != nil {
		return fmt.Errorf("error updating claim status: %w", err)
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAff == 0 {
		return ErrClaimDoesntExist
	}
	return nil
}

func (c *ClaimSponsor) updateClaimStatus(globalIndex *big.Int, status ClaimStatus) error {
	res, err := c.db.Exec(
		`UPDATE claim SET status = $1 WHERE global_index = $2`,
		status, globalIndex.String(),
	)
	if err != nil {
		return fmt.Errorf("error updating claim status: %w", err)
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}
	if rowsAff == 0 {
		return ErrClaimDoesntExist
	}
	return nil
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

func (c *ClaimSponsor) AddClaimToQueue(claim *Claim) error {
	claim.Status = PendingClaimStatus
	return meddler.Insert(c.db, "claim", claim)
}

func (c *ClaimSponsor) GetClaim(globalIndex *big.Int) (*Claim, error) {
	claim := &Claim{}
	err := meddler.QueryRow(
		c.db, claim, `SELECT * FROM claim WHERE global_index = $1`, globalIndex.String(),
	)
	return claim, db.ReturnErrNotFound(err)
}
