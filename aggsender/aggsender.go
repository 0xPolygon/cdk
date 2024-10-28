package aggsender

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	"github.com/0xPolygon/cdk/aggsender/types"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const signatureSize = 65

var (
	errNoBridgesAndClaims   = errors.New("no bridges and claims to build certificate")
	errInvalidSignatureSize = errors.New("invalid signature size")

	zeroLER = common.HexToHash("0x27ae5ba08d7291c96c8cbddcc148bf48a6d68c7974b94356f53754ef6171d757")
)

// AggSender is a component that will send certificates to the aggLayer
type AggSender struct {
	log aggsendertypes.Logger

	l2Syncer         aggsendertypes.L2BridgeSyncer
	l1infoTreeSyncer aggsendertypes.L1InfoTreeSyncer

	storage        db.AggSenderStorage
	aggLayerClient agglayer.AgglayerClientInterface

	cfg Config

	sequencerKey *ecdsa.PrivateKey
}

// New returns a new AggSender
func New(
	ctx context.Context,
	logger *log.Logger,
	cfg Config,
	aggLayerClient agglayer.AgglayerClientInterface,
	l1InfoTreeSyncer *l1infotreesync.L1InfoTreeSync,
	l2Syncer *bridgesync.BridgeSync) (*AggSender, error) {
	storage, err := db.NewAggSenderSQLStorage(logger, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	sequencerPrivateKey, err := cdkcommon.NewKeyFromKeystore(cfg.AggsenderPrivateKey)
	if err != nil {
		return nil, err
	}

	return &AggSender{
		cfg:              cfg,
		log:              logger,
		storage:          storage,
		l2Syncer:         l2Syncer,
		aggLayerClient:   aggLayerClient,
		l1infoTreeSyncer: l1InfoTreeSyncer,
		sequencerKey:     sequencerPrivateKey,
	}, nil
}

// Start starts the AggSender
func (a *AggSender) Start(ctx context.Context) {
	go a.sendCertificates(ctx)
	go a.checkIfCertificatesAreSettled(ctx)
}

// sendCertificates sends certificates to the aggLayer
func (a *AggSender) sendCertificates(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.BlockGetInterval.Duration)

	for {
		select {
		case <-ticker.C:
			if err := a.sendCertificate(ctx); err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			a.log.Info("AggSender stopped")
			return
		}
	}
}

// sendCertificate sends certificate for a network
func (a *AggSender) sendCertificate(ctx context.Context) error {
	a.log.Infof("trying to send a new certificate...")

	shouldSend, err := a.shouldSendCertificate(ctx)
	if err != nil {
		return err
	}

	if !shouldSend {
		a.log.Infof("waiting for pending certificates to be settled")
		return nil
	}

	lasL2BlockSynced, err := a.l2Syncer.GetLastProcessedBlock(ctx)
	if err != nil {
		return fmt.Errorf("error getting last processed block from l2: %w", err)
	}

	lastSentCertificateInfo, err := a.storage.GetLastSentCertificate(ctx)
	if err != nil {
		return err
	}

	previousToBlock := lastSentCertificateInfo.ToBlock
	if lastSentCertificateInfo.Status == agglayer.InError {
		// if the last certificate was in error, we need to resend it
		// from the block before the error
		previousToBlock = lastSentCertificateInfo.FromBlock - 1
	}

	if previousToBlock >= lasL2BlockSynced {
		a.log.Infof("no new blocks to send a certificate, last certificate block: %d, last L2 block: %d",
			previousToBlock, lasL2BlockSynced)
		return nil
	}

	fromBlock := previousToBlock + 1
	toBlock := lasL2BlockSynced

	bridges, err := a.l2Syncer.GetBridgesPublished(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("error getting bridges: %w", err)
	}

	if len(bridges) == 0 {
		a.log.Infof("no bridges consumed, no need to send a certificate from block: %d to block: %d", fromBlock, toBlock)
		return nil
	}

	claims, err := a.l2Syncer.GetClaims(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("error getting claims: %w", err)
	}

	a.log.Infof("building certificate for block: %d to block: %d", fromBlock, toBlock)

	certificate, err := a.buildCertificate(ctx, bridges, claims, lastSentCertificateInfo)
	if err != nil {
		return fmt.Errorf("error building certificate: %w", err)
	}

	signedCertificate, err := a.signCertificate(certificate)
	if err != nil {
		return fmt.Errorf("error signing certificate: %w", err)
	}

	a.logJSONCertificate(signedCertificate)
	a.saveCertificate(signedCertificate)

	certificateHash, err := a.aggLayerClient.SendCertificate(signedCertificate)
	if err != nil {
		return fmt.Errorf("error sending certificate: %w", err)
	}
	log.Infof("certificate send: Height: %d hash: %s", signedCertificate.Height, certificateHash.String())

	if err := a.storage.SaveLastSentCertificate(ctx, aggsendertypes.CertificateInfo{
		Height:           certificate.Height,
		CertificateID:    certificateHash,
		NewLocalExitRoot: certificate.NewLocalExitRoot,
		FromBlock:        fromBlock,
		ToBlock:          toBlock,
	}); err != nil {
		return fmt.Errorf("error saving last sent certificate in db: %w", err)
	}

	a.log.Infof("certificate: %s sent successfully for range of l2 blocks (from block: %d, to block: %d)",
		certificateHash, fromBlock, toBlock)

	return nil
}
func (a *AggSender) saveCertificate(signedCertificate *agglayer.SignedCertificate) {
	if signedCertificate == nil {
		return
	}
	fn := fmt.Sprintf("/tmp/certificate_%04d.json", signedCertificate.Height)
	a.log.Infof("saving certificate to file: %s", fn)
	jsonData, err := json.Marshal(signedCertificate)
	if err != nil {
		a.log.Errorf("error marshalling certificate: %w", err)
	}
	// write json data to file
	err = os.WriteFile(fn, jsonData, 0644)
	if err != nil {
		a.log.Errorf("error writing certificate to file: %w", err)
	}
}

// logJSONCertificate logs the certificate in JSON format to the logs
func (a *AggSender) logJSONCertificate(certificate *agglayer.SignedCertificate) {
	raw, err := json.Marshal(certificate)
	if err != nil {
		a.log.Errorf("error marshalling certificate: %w", err)
		return
	}

	a.log.Debug("JSON certificate:")
	a.log.Debug(string(raw))
}

// buildCertificate builds a certificate from the bridge events
func (a *AggSender) buildCertificate(ctx context.Context,
	bridges []bridgesync.Bridge,
	claims []bridgesync.Claim,
	lastSentCertificateInfo types.CertificateInfo) (*agglayer.Certificate, error) {
	if len(bridges) == 0 && len(claims) == 0 {
		return nil, errNoBridgesAndClaims
	}

	bridgeExits := a.getBridgeExits(bridges)
	importedBridgeExits, err := a.getImportedBridgeExits(ctx, claims)
	if err != nil {
		return nil, fmt.Errorf("error getting imported bridge exits: %w", err)
	}

	var depositCount uint32
	if len(bridges) > 0 {
		depositCount = bridges[len(bridges)-1].DepositCount
	}

	exitRoot, err := a.l2Syncer.GetExitRootByIndex(ctx, depositCount)
	if err != nil {
		return nil, fmt.Errorf("error getting exit root by index: %d. Error: %w", depositCount, err)
	}

	height := lastSentCertificateInfo.Height + 1
	previousLER := lastSentCertificateInfo.NewLocalExitRoot
	if lastSentCertificateInfo.NewLocalExitRoot == (common.Hash{}) {
		// meaning this is the first certificate
		height = 0
		previousLER = zeroLER
	}

	return &agglayer.Certificate{
		NetworkID:           a.l2Syncer.OriginNetwork(),
		PrevLocalExitRoot:   previousLER,
		NewLocalExitRoot:    exitRoot.Hash,
		BridgeExits:         bridgeExits,
		ImportedBridgeExits: importedBridgeExits,
		Height:              height,
	}, nil
}

// convertClaimToImportedBridgeExit converts a claim to an ImportedBridgeExit object
func (a *AggSender) convertClaimToImportedBridgeExit(claim bridgesync.Claim) (*agglayer.ImportedBridgeExit, error) {
	leafType := agglayer.LeafTypeAsset
	if claim.IsMessage {
		leafType = agglayer.LeafTypeMessage
	}

	bridgeExit := &agglayer.BridgeExit{
		LeafType: leafType,
		TokenInfo: &agglayer.TokenInfo{
			OriginNetwork:      claim.OriginNetwork,
			OriginTokenAddress: claim.OriginAddress,
		},
		DestinationNetwork: claim.DestinationNetwork,
		DestinationAddress: claim.DestinationAddress,
		Amount:             claim.Amount,
		Metadata:           claim.Metadata,
	}

	mainnetFlag, rollupIndex, leafIndex, err := bridgesync.DecodeGlobalIndex(claim.GlobalIndex)
	if err != nil {
		return nil, fmt.Errorf("error decoding global index: %w", err)
	}

	return &agglayer.ImportedBridgeExit{
		BridgeExit: bridgeExit,
		GlobalIndex: &agglayer.GlobalIndex{
			MainnetFlag: mainnetFlag,
			RollupIndex: rollupIndex,
			LeafIndex:   leafIndex,
		},
	}, nil
}

// getBridgeExits converts bridges to agglayer.BridgeExit objects
func (a *AggSender) getBridgeExits(bridges []bridgesync.Bridge) []*agglayer.BridgeExit {
	bridgeExits := make([]*agglayer.BridgeExit, 0, len(bridges))

	for _, bridge := range bridges {
		bridgeExits = append(bridgeExits, &agglayer.BridgeExit{
			LeafType: agglayer.LeafType(bridge.LeafType),
			TokenInfo: &agglayer.TokenInfo{
				OriginNetwork:      bridge.OriginNetwork,
				OriginTokenAddress: bridge.OriginAddress,
			},
			DestinationNetwork: bridge.DestinationNetwork,
			DestinationAddress: bridge.DestinationAddress,
			Amount:             bridge.Amount,
			Metadata:           bridge.Metadata,
		})
	}

	return bridgeExits
}

// getImportedBridgeExits converts claims to agglayer.ImportedBridgeExit objects and calculates necessary proofs
func (a *AggSender) getImportedBridgeExits(ctx context.Context,
	claims []bridgesync.Claim) ([]*agglayer.ImportedBridgeExit, error) {
	var (
		importedBridgeExits     = make([]*agglayer.ImportedBridgeExit, 0, len(claims))
		greatestL1InfoTreeIndex = uint32(0)
		ger                     common.Hash
		timestamp               uint64
		blockHash               common.Hash
		greatestClaimIndex      = 0
	)

	l1Infos := make([]*l1infotreesync.L1InfoTreeLeaf, 0, len(claims))
	for i, claim := range claims[:] {
		a.log.Debugf("claim[%d]: destAddr: %s GER:%s", i, claim.DestinationAddress.String(), claim.GlobalExitRoot.String())
		info, err := a.l1infoTreeSyncer.GetInfoByGlobalExitRoot(claim.GlobalExitRoot)
		if err != nil {
			return nil, fmt.Errorf("error getting info by global exit root: %w", err)
		}

		if greatestL1InfoTreeIndex < info.L1InfoTreeIndex {
			greatestL1InfoTreeIndex = info.L1InfoTreeIndex
			ger = claim.GlobalExitRoot
			timestamp = info.Timestamp
			blockHash = info.PreviousBlockHash
			greatestClaimIndex = i
		}

		importedBridgeExit, err := a.convertClaimToImportedBridgeExit(claim)
		if err != nil {
			return nil, fmt.Errorf("error converting claim to imported bridge exit: %w", err)
		}

		importedBridgeExits = append(importedBridgeExits, importedBridgeExit)
		l1Infos = append(l1Infos, info)
	}

	for i, ibe := range importedBridgeExits {
		l1Info := l1Infos[i]

		gerToL1Proof, err := a.l1infoTreeSyncer.GetL1InfoTreeMerkleProofFromIndexToRoot(ctx, l1Info.L1InfoTreeIndex, ger)
		if err != nil {
			return nil, fmt.Errorf("error getting L1 Info tree merkle proof for leaf index: %d. GER: %s. Error: %w",
				l1Info.L1InfoTreeIndex, ger, err)
		}

		claim := claims[i]
		if ibe.GlobalIndex.MainnetFlag {
			ibe.ClaimData = &agglayer.ClaimFromMainnnet{
				L1Leaf: &agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: greatestL1InfoTreeIndex,
					RollupExitRoot:  claims[greatestClaimIndex].RollupExitRoot,
					MainnetExitRoot: claims[greatestClaimIndex].MainnetExitRoot,
					Inner: &agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: ger,
						Timestamp:      timestamp,
						BlockHash:      blockHash,
					},
				},
				ProofLeafMER: &agglayer.MerkleProof{
					Root:  claims[greatestClaimIndex].MainnetExitRoot,
					Proof: claims[greatestClaimIndex].ProofLocalExitRoot,
				},
				ProofGERToL1Root: &agglayer.MerkleProof{
					Root:  ger,
					Proof: gerToL1Proof,
				},
			}
		} else {
			ibe.ClaimData = &agglayer.ClaimFromRollup{
				L1Leaf: &agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: greatestL1InfoTreeIndex,
					RollupExitRoot:  claims[greatestClaimIndex].RollupExitRoot,
					MainnetExitRoot: claims[greatestClaimIndex].MainnetExitRoot,
					Inner: &agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: claim.GlobalExitRoot,
						Timestamp:      timestamp,
						BlockHash:      blockHash,
					},
				},
				ProofLeafLER: &agglayer.MerkleProof{
					Root:  claims[greatestClaimIndex].MainnetExitRoot,
					Proof: claims[greatestClaimIndex].ProofLocalExitRoot,
				},
				ProofLERToRER: &agglayer.MerkleProof{
					Root:  claims[greatestClaimIndex].RollupExitRoot,
					Proof: claims[greatestClaimIndex].ProofRollupExitRoot,
				},
				ProofGERToL1Root: &agglayer.MerkleProof{
					Root:  ger,
					Proof: gerToL1Proof,
				},
			}
		}
	}

	return importedBridgeExits, nil
}

// signCertificate signs a certificate with the sequencer key
func (a *AggSender) signCertificate(certificate *agglayer.Certificate) (*agglayer.SignedCertificate, error) {
	hashToSign := certificate.Hash()

	sig, err := crypto.Sign(hashToSign.Bytes(), a.sequencerKey)
	if err != nil {
		return nil, err
	}

	r, s, isOddParity, err := extractSignatureData(sig)
	if err != nil {
		return nil, err
	}

	return &agglayer.SignedCertificate{
		Certificate: certificate,
		Signature: &agglayer.Signature{
			R:         r,
			S:         s,
			OddParity: isOddParity,
		},
	}, nil
}

// checkIfCertificatesAreSettled checks if certificates are settled
func (a *AggSender) checkIfCertificatesAreSettled(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.CheckSettledInterval.Duration)
	for {
		select {
		case <-ticker.C:
			a.checkPendingCertificatesStatus(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// checkPendingCertificatesStatus checks the status of pending certificates
// and updates in the storage if it changed on agglayer
func (a *AggSender) checkPendingCertificatesStatus(ctx context.Context) {
	pendingCertificates, err := a.storage.GetCertificatesByStatus(ctx, []agglayer.CertificateStatus{agglayer.Pending})
	if err != nil {
		a.log.Errorf("error getting pending certificates: %w", err)
	}

	for _, certificate := range pendingCertificates {
		certificateHeader, err := a.aggLayerClient.GetCertificateHeader(certificate.CertificateID)
		if err != nil {
			a.log.Errorf("error getting header of certificate %s with height: %d from agglayer: %w",
				certificate.CertificateID, certificate.Height, err)
			continue
		}

		if certificateHeader.Status == agglayer.Settled || certificateHeader.Status == agglayer.InError {
			certificate.Status = certificateHeader.Status

			a.log.Infof("certificate %s changed status to %s", certificateHeader.String(), certificate.Status)

			if err := a.storage.UpdateCertificateStatus(ctx, *certificate); err != nil {
				a.log.Errorf("error updating certificate status in storage: %w", err)
				continue
			}
		}
	}
}

// shouldSendCertificate checks if a certificate should be sent at given time
// if we have pending certificates, then we wait until they are settled
func (a *AggSender) shouldSendCertificate(ctx context.Context) (bool, error) {
	pendingCertificates, err := a.storage.GetCertificatesByStatus(ctx, []agglayer.CertificateStatus{agglayer.Pending})
	if err != nil {
		return false, fmt.Errorf("error getting pending certificates: %w", err)
	}

	return len(pendingCertificates) == 0, nil
}

// extractSignatureData extracts the R, S, and V from a 65-byte signature
func extractSignatureData(signature []byte) (r, s common.Hash, isOddParity bool, err error) {
	if len(signature) != signatureSize {
		err = errInvalidSignatureSize
		return
	}

	r = common.BytesToHash(signature[:32])   // First 32 bytes are R
	s = common.BytesToHash(signature[32:64]) // Next 32 bytes are S
	isOddParity = signature[64]%2 == 1       //nolint:mnd // Last byte is V

	return
}
