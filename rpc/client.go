package rpc

// ClientInterface is the interface that defines the implementation of all the endpoints
type ClientInterface interface {
	BridgeClientInterface
}

// ClientFactoryInterface interface for the client factory
type ClientFactoryInterface interface {
	NewClient(url string) ClientInterface
}

// ClientFactory is the implementation of the data committee client factory
type ClientFactory struct{}

// NewClient returns an implementation of the data committee node client
func (f *ClientFactory) NewClient(url string) ClientInterface {
	return NewClient(url)
}

// Client wraps all the available endpoints of the data abailability committee node server
type Client struct {
	url string
}

// NewClient returns a client ready to be used
func NewClient(url string) *Client {
	return &Client{
		url: url,
	}
}
