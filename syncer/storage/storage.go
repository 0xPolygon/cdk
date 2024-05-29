package storage

// Storage is an interface for the storage of the syncer
type StorageService interface {
	GetData() []byte
	GetLatestVerifiedBlock() uint64
}
