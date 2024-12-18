package main

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/rpcclient"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/log"
)

const (
	errLevelUnexpected         = 1
	errLevelWrongParams        = 2
	errLevelComms              = 3
	errLevelNotFound           = 4
	errLevelFoundButNotSettled = 5
)

func unmarshalGlobalIndex(globalIndex string) (*agglayer.GlobalIndex, error) {
	var globalIndexParsed agglayer.GlobalIndex
	// First try if it's already decomposed
	err := json.Unmarshal([]byte(globalIndex), &globalIndexParsed)
	if err != nil {
		bigInt := new(big.Int)
		_, ok := bigInt.SetString(globalIndex, 10)
		if !ok {
			return nil, fmt.Errorf("invalid global index: %v", globalIndex)
		}
		mainnetFlag, rollupIndex, leafIndex, err := bridgesync.DecodeGlobalIndex(bigInt)
		if err != nil {
			return nil, fmt.Errorf("invalid global index, fail to decode: %v", globalIndex)
		}
		globalIndexParsed.MainnetFlag = mainnetFlag
		globalIndexParsed.RollupIndex = rollupIndex
		globalIndexParsed.LeafIndex = leafIndex
	}
	return &globalIndexParsed, nil
}

// This function find out the certificate for a deposit
// It use the aggsender RPC
func certContainsGlobalIndex(cert *types.CertificateInfo, globalIndex *agglayer.GlobalIndex) (bool, error) {
	if cert == nil {
		return false, nil
	}
	var certSigned agglayer.SignedCertificate
	err := json.Unmarshal([]byte(cert.SignedCertificate), &certSigned)
	if err != nil {
		return false, err
	}
	for _, importedBridge := range certSigned.ImportedBridgeExits {
		if *importedBridge.GlobalIndex == *globalIndex {
			return true, nil
		}

	}
	return false, nil
}

func main() {
	aggsenderRPC := os.Args[1]
	globalIndex := os.Args[2]
	decodedGlobalIndex, err := unmarshalGlobalIndex(globalIndex)
	if err != nil {
		log.Errorf("Error unmarshalGlobalIndex: %v", err)
		os.Exit(errLevelWrongParams)
	}
	aggsenderClient := rpcclient.NewClient(aggsenderRPC)
	// Get first certificate
	cert, err := aggsenderClient.GetCertificateHeaderPerHeight(nil)
	if err != nil {
		log.Errorf("Error: %v", err)
		os.Exit(errLevelComms)
	}
	currentHeight := cert.Height
	for cert != nil {
		found, err := certContainsGlobalIndex(cert, decodedGlobalIndex)
		if err != nil {
			log.Errorf("Error: %v", err)
			os.Exit(1)
		}
		if found {
			log.Infof("Found certificate for global index: %v", globalIndex)
			if cert.Status.IsSettled() {
				log.Infof("Certificate is settled")
				os.Exit(0)
			}
			log.Errorf("Certificate is not settled")
			os.Exit(errLevelFoundButNotSettled)
		} else {
			log.Debugf("Certificate not found for global index: %v", globalIndex)
		}
		// We have check the oldest cert
		if currentHeight == 0 {
			log.Errorf("Checked all certs and it's not found")
			os.Exit(errLevelNotFound)
		}
		log.Infof("Checking previous certificate, height: %v", currentHeight)
		cert, err = aggsenderClient.GetCertificateHeaderPerHeight(&currentHeight)
		if err != nil {
			log.Errorf("Error: %v", err)
			os.Exit(errLevelComms)
		}
		currentHeight--
	}

}
