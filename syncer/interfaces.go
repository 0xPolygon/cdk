package syncer

type EventSyncer interface {
	GetData() []byte
}

type EventProcessor interface {
	Process(data []byte) error
}