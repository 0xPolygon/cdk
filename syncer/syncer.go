package syncer

type Syncer interface {
	// Start starts the syncer
	Start() error
	// Stop stops the syncer
	Stop() error
	// Synced returns true if the syncer is synced
	Synced() bool
	// SyncedChan returns a channel that is closed when the syncer is synced
	SyncedChan() <-chan struct{}
	// LatestVerifiedBlock returns the latest verified block
	LatestVerifiedBlock() uint64
	// LatestBlock returns the latest block
	LatestBlock() uint64
	// SyncedBlock returns the latest block that is synced
	SyncedBlock() uint64
	// SyncedBlockChan returns a channel that is closed when the latest block is synced
	SyncedBlockChan() <-chan struct{}
}

type Sync struct {
}
