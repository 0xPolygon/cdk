package types

import (
	"fmt"

	"github.com/0xPolygon/cdk/bridgesync"
)

const (
	EstimatedSizeBridgeExit = 250
	EstimatedSizeClaim      = 44000
)

type CertificateLocalParams struct {
	FromBlock uint64
	ToBlock   uint64
	Bridges   []bridgesync.Bridge
	Claims    []bridgesync.Claim
}

func (c *CertificateLocalParams) String() string {
	return fmt.Sprintf("FromBlock: %d, ToBlock: %d, numBridges: %d, numClaims: %d", c.FromBlock, c.ToBlock, c.NumberOfBridges(), c.NumberOfClaims())
}

func (c *CertificateLocalParams) Range(fromBlock, toBlock uint64) (*CertificateLocalParams, error) {
	if c.FromBlock == fromBlock && c.ToBlock == toBlock {
		return c, nil
	}
	if c.FromBlock > fromBlock || c.ToBlock < toBlock {
		return nil, fmt.Errorf("invalid range")
	}
	newCert := &CertificateLocalParams{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Bridges:   make([]bridgesync.Bridge, 0),
		Claims:    make([]bridgesync.Claim, 0),
	}

	for _, bridge := range c.Bridges {
		if bridge.BlockNum >= fromBlock && bridge.BlockNum <= toBlock {
			newCert.Bridges = append(newCert.Bridges, bridge)
		}
	}

	for _, claim := range c.Claims {
		if claim.BlockNum >= fromBlock && claim.BlockNum <= toBlock {
			newCert.Claims = append(newCert.Claims, claim)
		}
	}
	return newCert, nil
}

func (c *CertificateLocalParams) NumberOfBridges() int {
	if c == nil {
		return 0
	}
	return len(c.Bridges)
}

func (c *CertificateLocalParams) NumberOfClaims() int {
	if c == nil {
		return 0
	}
	return len(c.Claims)
}

func (c *CertificateLocalParams) NumberOfBlocks() int {
	if c == nil {
		return 0
	}
	return int(c.ToBlock - c.FromBlock + 1)
}

func (c *CertificateLocalParams) EstimatedSize() uint {
	if c == nil {
		return 0
	}
	numBridges := len(c.Bridges)
	numClaims := len(c.Claims)
	return uint(numBridges*EstimatedSizeBridgeExit + numClaims*EstimatedSizeClaim)
}

func (c *CertificateLocalParams) IsEmpty() bool {
	return c.NumberOfBridges() == 0 && c.NumberOfClaims() == 0
}

func (c *CertificateLocalParams) MaxDepoitCount() uint32 {
	if c == nil || c.NumberOfBridges() == 0 {
		return 0
	}
	return c.Bridges[len(c.Bridges)-1].DepositCount
}
