package aggsender

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	treeTypes "github.com/0xPolygon/cdk/tree/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// L1InfoTreeSyncer is an interface defining functions that an L1InfoTreeSyncer should implement
type L1InfoTreeSyncer interface {
	GetInfoByGlobalExitRoot(globalExitRoot common.Hash) (*l1infotreesync.L1InfoTreeLeaf, error)
	GetL1InfoTreeMerkleProofFromIndexToRoot(ctx context.Context,
		index uint32, root common.Hash) (treeTypes.Proof, error)
}

// AggSender is a component that will send certificates to the aggLayer
type AggSender struct {
	log *log.Logger

	l2Syncer         *bridgesync.BridgeSync
	l2Client         bridgesync.EthClienter
	l1infoTreeSyncer L1InfoTreeSyncer

	storage        db.AggSenderStorage
	aggLayerClient agglayer.AgglayerClientInterface

	sendInterval types.Duration

	sequencerKey *ecdsa.PrivateKey
}

// New returns a new AggSender
func New(
	ctx context.Context,
	logger *log.Logger,
	cfg Config,
	aggLayerClient agglayer.AgglayerClientInterface,
	l1InfoTreeSyncer *l1infotreesync.L1InfoTreeSync,
	l2Syncer *bridgesync.BridgeSync,
	l2Client bridgesync.EthClienter) (*AggSender, error) {
	storage, err := db.NewAggSenderSQLStorage(logger, cfg.DBPath)
	if err != nil {
		return nil, err
	}

	sequencerPrivateKey, err := cdkcommon.NewKeyFromKeystore(cfg.SequencerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &AggSender{
		log:              logger,
		storage:          storage,
		l2Syncer:         l2Syncer,
		l2Client:         l2Client,
		aggLayerClient:   aggLayerClient,
		l1infoTreeSyncer: l1InfoTreeSyncer,
		sequencerKey:     sequencerPrivateKey,
		sendInterval:     cfg.CertificateSendInterval,
	}, nil
}

// Start starts the AggSender
func (a *AggSender) Start(ctx context.Context) {
	a.sendCertificates(ctx)
}

// sendCertificates sends certificates to the aggLayer
func (a *AggSender) sendCertificates(ctx context.Context) {
	ticker := time.NewTicker(a.sendInterval.Duration)

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
	a.log.Info("trying to send a new certificate...")

	lastSentCertificate, err := a.storage.GetLastSentCertificate(ctx)
	if err != nil {
		return fmt.Errorf("error getting last sent certificate: %w", err)
	}

	a.log.Info("last sent certificate: %s", lastSentCertificate.CertificateID)

	finality := a.l2Syncer.BlockFinality()
	blockFinality, err := finality.ToBlockNum()
	if err != nil {
		return fmt.Errorf("error getting block finality: %w", err)
	}

	lastL2Block, err := a.l2Client.HeaderByNumber(ctx, blockFinality)
	if err != nil {
		return fmt.Errorf("error getting block from l2: %w", err)
	}

	var (
		previousLocalExitRoot common.Hash
		previousHeight        uint64
		lastCertificateBlock  uint64
	)

	if lastSentCertificate.CertificateID != (common.Hash{}) {
		// we have sent a certificate before, get the last certificate header
		lastSentCertificateHeader, err := a.aggLayerClient.GetCertificateHeader(lastSentCertificate.CertificateID)
		if err != nil {
			return fmt.Errorf("error getting certificate %s header: %w", lastSentCertificate.CertificateID, err)
		}

		if lastSentCertificateHeader.Status == agglayer.InError {
			// last sent certificate had errors, we need to remove it from the db
			// and build a new certificate from that block
			if err := a.storage.DeleteCertificate(ctx, lastSentCertificateHeader.CertificateID); err != nil {
				return fmt.Errorf("error deleting certificate %s: %w", lastSentCertificate.CertificateID, err)
			}

			lastValidCertificate, err := a.storage.GetCertificateByHeight(ctx, lastSentCertificateHeader.Height)
			if err != nil {
				return fmt.Errorf("error getting certificate by height %d: %w", lastSentCertificateHeader.Height, err)
			}

			previousLocalExitRoot = lastValidCertificate.NewLocalExitRoot
			previousHeight = lastValidCertificate.Height
		} else {
			previousLocalExitRoot = lastSentCertificateHeader.NewLocalExitRoot
			previousHeight = lastSentCertificateHeader.Height
		}

		lastCertificateBlock, err = a.l2Syncer.GetBlockByLER(ctx, lastSentCertificateHeader.NewLocalExitRoot)
		if err != nil {
			return fmt.Errorf("error getting block by LER %s: %w", lastSentCertificate.CertificateID, err)
		}
	}

	if lastL2Block.Number.Uint64() <= lastCertificateBlock {
		a.log.Info("no new blocks to send a certificate")
		return nil
	}

	fromBlock := lastCertificateBlock + 1
	toBlock := lastL2Block.Number.Uint64()

	bridges, err := a.l2Syncer.GetBridges(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("error getting bridges: %w", err)
	}

	if len(bridges) == 0 {
		a.log.Info("no bridges consumed, no need to send a certificate from block: %d to block: %d", fromBlock, toBlock)
		return nil
	}

	claims, err := a.l2Syncer.GetClaims(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("error getting claims: %w", err)
	}

	a.log.Info("building certificate for block: %d to block: %d", fromBlock, toBlock)

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

	a.log.Info("certificate: %s sent successfully for block: %d to block: %d", certificateHash, fromBlock, toBlock)

	return nil
}

// buildCertificate builds a certificate from the bridge events
func (a *AggSender) buildCertificate(ctx context.Context,
	bridges []bridgesync.Bridge,
	claims []bridgesync.Claim,
	previousExitRoot common.Hash, lastHeight uint64) (*agglayer.Certificate, error) {
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
