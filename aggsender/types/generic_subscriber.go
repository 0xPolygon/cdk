package types

type GenericSubscriber[T any] interface {
	Subscribe(subscriberName string) <-chan T
	Publish(data T)
}
