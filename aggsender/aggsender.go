package aggsender

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var errNoBridgesAndClaims = errors.New("no bridges and claims to build certificate")

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

	previousLocalExitRoot, previousHeight, lastCertificateBlock, err := a.getLastSentCertificateData(ctx)
	if err != nil {
		return err
	}

	if lastCertificateBlock >= lasL2BlockSynced {
		a.log.Infof("no new blocks to send a certificate, last certificate block: %d, last L2 block: %d",
			lastCertificateBlock, lasL2BlockSynced)
		return nil
	}

	fromBlock := lastCertificateBlock + 1
	toBlock := lasL2BlockSynced

	bridges, err := a.l2Syncer.GetBridges(ctx, fromBlock, toBlock)
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

	certificate, err := a.buildCertificate(ctx, bridges, claims, previousLocalExitRoot, previousHeight)
	if err != nil {
		return fmt.Errorf("error building certificate: %w", err)
	}

	signedCertificate, err := a.signCertificate(certificate)
	if err != nil {
		return fmt.Errorf("error signing certificate: %w", err)
	}

	certificateHash, err := a.aggLayerClient.SendCertificate(signedCertificate)
	if err != nil {
		return fmt.Errorf("error sending certificate: %w", err)
	}

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

// buildCertificate builds a certificate from the bridge events
func (a *AggSender) buildCertificate(ctx context.Context,
	bridges []bridgesync.Bridge,
	claims []bridgesync.Claim,
	previousExitRoot common.Hash, lastHeight uint64) (*agglayer.Certificate, error) {
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

	return &agglayer.Certificate{
		NetworkID:           a.l2Syncer.OriginNetwork(),
		PrevLocalExitRoot:   previousExitRoot,
		NewLocalExitRoot:    exitRoot.Hash,
		BridgeExits:         bridgeExits,
		ImportedBridgeExits: importedBridgeExits,
		Height:              lastHeight + 1,
	}, nil
}

// getLastSentCertificateData gets the previous local exit root, previous certificate height
// and last certificate block sent (gotten from the last sent certificate in db)
func (a *AggSender) getLastSentCertificateData(ctx context.Context) (common.Hash, uint64, uint64, error) {
	lastSentCertificate, err := a.storage.GetLastSentCertificate(ctx)
	if err != nil {
		return common.Hash{}, 0, 0, fmt.Errorf("error getting last sent certificate: %w", err)
	}

	a.log.Infof("last sent certificate: %s", lastSentCertificate.String())

	var (
		previousLocalExitRoot common.Hash
		previousHeight        uint64
		lastCertificateBlock  uint64
	)

	if lastSentCertificate.CertificateID != (common.Hash{}) {
		// we have sent a certificate before, get the last certificate header
		lastSentCertificateHeader, err := a.aggLayerClient.GetCertificateHeader(lastSentCertificate.CertificateID)
		if err != nil {
			return common.Hash{}, 0, 0, fmt.Errorf("error getting certificate %s header: %w",
				lastSentCertificate.CertificateID, err)
		}

		if lastSentCertificateHeader.Status == agglayer.InError {
			// last sent certificate had errors, we need to remove it from the db
			// and build a new certificate from that block
			if err := a.storage.DeleteCertificate(ctx, lastSentCertificateHeader.CertificateID); err != nil {
				return common.Hash{}, 0, 0, fmt.Errorf("error deleting certificate %s: %w",
					lastSentCertificate.CertificateID, err)
			}

			lastValidCertificate, err := a.storage.GetCertificateByHeight(ctx, lastSentCertificateHeader.Height-1)
			if err != nil {
				return common.Hash{}, 0, 0, fmt.Errorf("error getting certificate by height %d: %w",
					lastSentCertificateHeader.Height, err)
			}

			previousLocalExitRoot = lastValidCertificate.NewLocalExitRoot
			previousHeight = lastValidCertificate.Height
			lastCertificateBlock = lastValidCertificate.ToBlock
		} else {
			previousLocalExitRoot = lastSentCertificateHeader.NewLocalExitRoot
			previousHeight = lastSentCertificateHeader.Height
			lastCertificateBlock = lastSentCertificate.ToBlock
		}
	}

	return previousLocalExitRoot, previousHeight, lastCertificateBlock, nil
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
	)

	for _, claim := range claims {
		info, err := a.l1infoTreeSyncer.GetInfoByGlobalExitRoot(claim.GlobalExitRoot)
		if err != nil {
			return nil, fmt.Errorf("error getting info by global exit root: %w", err)
		}

		if greatestL1InfoTreeIndex < info.L1InfoTreeIndex {
			greatestL1InfoTreeIndex = info.L1InfoTreeIndex
			ger = claim.GlobalExitRoot
			timestamp = info.Timestamp
			blockHash = info.PreviousBlockHash
		}

		importedBridgeExit, err := a.convertClaimToImportedBridgeExit(claim)
		if err != nil {
			return nil, fmt.Errorf("error converting claim to imported bridge exit: %w", err)
		}

		importedBridgeExits = append(importedBridgeExits, importedBridgeExit)
	}

	for i, ibe := range importedBridgeExits {
		gerToL1Proof, err := a.l1infoTreeSyncer.GetL1InfoTreeMerkleProofFromIndexToRoot(ctx, ibe.GlobalIndex.LeafIndex, ger)
		if err != nil {
			return nil, fmt.Errorf("error getting L1 Info tree merkle proof: %w", err)
		}

		claim := claims[i]
		if ibe.GlobalIndex.MainnetFlag {
			ibe.ClaimData = &agglayer.ClaimFromMainnnet{
				L1Leaf: agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: ibe.GlobalIndex.LeafIndex,
					RollupExitRoot:  claims[i].RollupExitRoot,
					MainnetExitRoot: claims[i].MainnetExitRoot,
					Inner: agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: ger,
						Timestamp:      timestamp,
						BlockHash:      blockHash,
					},
				},
				ProofLeafMER: agglayer.MerkleProof{
					Root:  claim.MainnetExitRoot,
					Proof: claim.ProofLocalExitRoot,
				},
				ProofGERToL1Root: agglayer.MerkleProof{
					Root:  ger,
					Proof: gerToL1Proof,
				},
			}
		} else {
			ibe.ClaimData = &agglayer.ClaimFromRollup{
				L1Leaf: agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: ibe.GlobalIndex.LeafIndex,
					RollupExitRoot:  claim.RollupExitRoot,
					MainnetExitRoot: claim.MainnetExitRoot,
					Inner: agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: ger,
						Timestamp:      timestamp,
						BlockHash:      blockHash,
					},
				},
				ProofLeafLER: agglayer.MerkleProof{
					Root:  claim.MainnetExitRoot,
					Proof: claim.ProofLocalExitRoot,
				},
				ProofLERToRER: agglayer.MerkleProof{},
				ProofGERToL1Root: agglayer.MerkleProof{
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

	return &agglayer.SignedCertificate{
		Certificate: certificate,
		Signature:   sig,
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
