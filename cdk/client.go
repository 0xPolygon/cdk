package cdk

import (
	"context"

	"github.com/0xPolygon/cdk/executionlayer"
)

type Client struct {
	executionLayer executionlayer.ExecutionLayer
}

func NewClient(executionLayer executionlayer.ExecutionLayer) *Client {
	return &Client{executionLayer: executionLayer}
}

func (c *Client) Start() error {
	return c.executionLayer.Start(context.Background())
}
