package storage

// Storage is an interface for the storage of the syncer
type Storage interface {
	GetData() []byte
	GetLatestVerifiedBlock() uint64
}
