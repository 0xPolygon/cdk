package types

import (
	"fmt"

	"github.com/0xPolygon/cdk/bridgesync"
)

const (
	EstimatedSizeBridgeExit = 250
	EstimatedSizeClaim      = 44000
	byteArrayJsonSizeFactor = 1.5
)

// CertificateBuildParams is a struct that holds the parameters to build a certificate
type CertificateBuildParams struct {
	FromBlock uint64
	ToBlock   uint64
	Bridges   []bridgesync.Bridge
	Claims    []bridgesync.Claim
}

func (c *CertificateBuildParams) String() string {
	return fmt.Sprintf("FromBlock: %d, ToBlock: %d, numBridges: %d, numClaims: %d",
		c.FromBlock, c.ToBlock, c.NumberOfBridges(), c.NumberOfClaims())
}

// Range create a new CertificateBuildParams with the given range
func (c *CertificateBuildParams) Range(fromBlock, toBlock uint64) (*CertificateBuildParams, error) {
	if c.FromBlock == fromBlock && c.ToBlock == toBlock {
		return c, nil
	}
	if c.FromBlock > fromBlock || c.ToBlock < toBlock {
		return nil, fmt.Errorf("invalid range")
	}
	newCert := &CertificateBuildParams{
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

// NumberOfBridges returns the number of bridges in the certificate
func (c *CertificateBuildParams) NumberOfBridges() int {
	if c == nil {
		return 0
	}
	return len(c.Bridges)
}

// NumberOfClaims returns the number of claims in the certificate
func (c *CertificateBuildParams) NumberOfClaims() int {
	if c == nil {
		return 0
	}
	return len(c.Claims)
}

// NumberOfBlocks returns the number of blocks in the certificate
func (c *CertificateBuildParams) NumberOfBlocks() int {
	if c == nil {
		return 0
	}
	return int(c.ToBlock - c.FromBlock + 1)
}

// EstimatedSize returns the estimated size of the certificate
func (c *CertificateBuildParams) EstimatedSize() uint {
	if c == nil {
		return 0
	}
	numBridges := len(c.Bridges)
	sizeClaims := int(0)
	for _, claim := range c.Claims {
		sizeClaims += EstimatedSizeClaim
		sizeClaims += int(byteArrayJsonSizeFactor * float32(len(claim.Metadata)))
	}
	return uint(numBridges*EstimatedSizeBridgeExit + sizeClaims)
}

// IsEmpty returns true if the certificate is empty
func (c *CertificateBuildParams) IsEmpty() bool {
	return c.NumberOfBridges() == 0 && c.NumberOfClaims() == 0
}

// MaxDepoitCount returns the maximum deposit count in the certificate
func (c *CertificateBuildParams) MaxDepoitCount() uint32 {
	if c == nil || c.NumberOfBridges() == 0 {
		return 0
	}
	return c.Bridges[len(c.Bridges)-1].DepositCount
}
