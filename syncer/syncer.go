package syncer

type Syncer interface {
	// Start starts the syncer
	Start() error
	// Stop stops the syncer
	Stop() error
}
