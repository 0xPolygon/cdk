package ethtxmanager

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
)

// MemStorage hold txs to be managed
type MemStorage struct {
	TxsMutex            sync.RWMutex
	FileMutex           sync.RWMutex
	Transactions        map[common.Hash]monitoredTx
	PersistenceFilename string
}

// NewMemStorage creates a new instance of storage
func NewMemStorage(persistenceFilename string) *MemStorage {
	transactions := make(map[common.Hash]monitoredTx)
	if persistenceFilename != "" {
		// Check if the file exists
		if _, err := os.Stat(persistenceFilename); os.IsNotExist(err) {
			log.Infof("Persistence file %s does not exist", persistenceFilename)
		} else {
			ReadFile, err := os.ReadFile(persistenceFilename)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			err = json.Unmarshal(ReadFile, &transactions)
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
			log.Infof("Persistence file %s loaded", persistenceFilename)
		}
	}

	return &MemStorage{TxsMutex: sync.RWMutex{},
		Transactions:        transactions,
		PersistenceFilename: persistenceFilename,
	}
}

// Persist the memory storage
func (s *MemStorage) persist() {
	if s.PersistenceFilename != "" {
		s.TxsMutex.RLock()
		defer s.TxsMutex.RUnlock()
		s.FileMutex.Lock()
		defer s.FileMutex.Unlock()
		jsonFile, _ := json.Marshal(s.Transactions)
		err := os.WriteFile(s.PersistenceFilename+".tmp", jsonFile, 0644) //nolint:gosec,mnd
		if err != nil {
			log.Error(err)
		}
		err = os.Rename(s.PersistenceFilename+".tmp", s.PersistenceFilename)
		if err != nil {
			log.Error(err)
		}
	}
}

// Add persist a monitored tx
func (s *MemStorage) Add(ctx context.Context, mTx monitoredTx) error {
	mTx.CreatedAt = time.Now()
	s.TxsMutex.Lock()
	if _, exists := s.Transactions[mTx.ID]; exists {
		return ErrAlreadyExists
	}
	s.Transactions[mTx.ID] = mTx
	s.TxsMutex.Unlock()
	s.persist()
	return nil
}

// Remove a persisted monitored tx
func (s *MemStorage) Remove(ctx context.Context, id common.Hash) error {
	s.TxsMutex.Lock()
	if _, exists := s.Transactions[id]; !exists {
		return ErrNotFound
	}
	delete(s.Transactions, id)
	s.TxsMutex.Unlock()
	s.persist()
	return nil
}

// Get loads a persisted monitored tx
func (s *MemStorage) Get(ctx context.Context, id common.Hash) (monitoredTx, error) {
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()
	if mTx, exists := s.Transactions[id]; exists {
		return mTx, nil
	}
	return monitoredTx{}, ErrNotFound
}

// GetByStatus loads all monitored tx that match the provided status
func (s *MemStorage) GetByStatus(ctx context.Context, statuses []MonitoredTxStatus) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()
	for _, mTx := range s.Transactions {
		if len(statuses) > 0 {
			for _, status := range statuses {
				if mTx.Status == status {
					mTxs = append(mTxs, mTx)
				}
			}
		} else {
			mTxs = append(mTxs, mTx)
		}
	}

	// ensure the transactions are ordered by creation date
	// (oldest first)
	for i := 0; i < len(mTxs); i++ {
		for j := i + 1; j < len(mTxs); j++ {
			if mTxs[i].CreatedAt.After(mTxs[j].CreatedAt) {
				mTxs[i], mTxs[j] = mTxs[j], mTxs[i]
			}
		}
	}

	return mTxs, nil
}

// GetByBlock loads all monitored tx that have the blockNumber between
// fromBlock and toBlock
func (s *MemStorage) GetByBlock(ctx context.Context, fromBlock, toBlock *uint64) ([]monitoredTx, error) {
	mTxs := []monitoredTx{}
	s.TxsMutex.RLock()
	defer s.TxsMutex.RUnlock()
	for _, mTx := range s.Transactions {
		if fromBlock != nil && mTx.BlockNumber.Uint64() < *fromBlock {
			continue
		}
		if toBlock != nil && mTx.BlockNumber.Uint64() > *toBlock {
			continue
		}
		mTxs = append(mTxs, mTx)
	}
	return mTxs, nil
}

// Update a persisted monitored tx
func (s *MemStorage) Update(ctx context.Context, mTx monitoredTx) error {
	mTx.UpdatedAt = time.Now()
	s.TxsMutex.Lock()

	if _, exists := s.Transactions[mTx.ID]; !exists {
		return ErrNotFound
	}
	s.Transactions[mTx.ID] = mTx
	s.TxsMutex.Unlock()
	s.persist()
	return nil
}

// Empty the storage
func (s *MemStorage) Empty(ctx context.Context) error {
	s.TxsMutex.Lock()
	s.Transactions = make(map[common.Hash]monitoredTx)
	s.TxsMutex.Unlock()
	s.persist()
	return nil
}
