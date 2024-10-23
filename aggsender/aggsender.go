package aggsender

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	aggsendertypes "github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var errNoBridgesAndClaims = errors.New("no bridges and claims to build certificate")

// AggSender is a component that will send certificates to the aggLayer
type AggSender struct {
	log aggsendertypes.Logger

	l2Syncer         aggsendertypes.L2BridgeSyncer
	l2Client         aggsendertypes.EthClient
	l1infoTreeSyncer aggsendertypes.L1InfoTreeSyncer
	l1Client         aggsendertypes.EthClient

	storage        db.AggSenderStorage
	aggLayerClient agglayer.AgglayerClientInterface

	cfg Config

	sequencerKey *ecdsa.PrivateKey

	lock                   sync.Mutex
	lastL1CertificateBlock uint64
}

// New returns a new AggSender
func New(
	ctx context.Context,
	logger *log.Logger,
	cfg Config,
	aggLayerClient agglayer.AgglayerClientInterface,
	l1Client aggsendertypes.EthClient,
	l1InfoTreeSyncer *l1infotreesync.L1InfoTreeSync,
	l2Syncer *bridgesync.BridgeSync,
	l2Client aggsendertypes.EthClient) (*AggSender, error) {
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
		l2Client:         l2Client,
		l1Client:         l1Client,
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

	l1Block, err := a.l1Client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("error getting l1 block number: %w", err)
	}

	if !a.shouldSendCertificate(l1Block) {
		a.log.Infof("block %d on L1 not near epoch ending, so we don't send a certificate", l1Block)
		return nil
	}

	lastL2Block, err := a.getLastL2Block(ctx)
	if err != nil {
		return fmt.Errorf("error getting block from l2: %w", err)
	}
	lasL2BlockSynced, err := a.l2Syncer.GetLastProcessedBlock(ctx)
	if err != nil {
		return fmt.Errorf("error getting last processed block from l2: %w", err)
	}
	log.Debugf("lastL2Block: %d, lasL2BlockSynced: %d", lastL2Block.Number.Uint64(), lasL2BlockSynced)

	previousLocalExitRoot, previousHeight, lastCertificateBlock, err := a.getLastSentCertificateData(ctx)
	if err != nil {
		return err
	}
	fromBlock := lastCertificateBlock + 1
	toBlock, err := chooseToL2Block(lastL2Block.Number.Uint64(), lasL2BlockSynced, lastCertificateBlock)
	if err != nil {
		a.log.Errorf("no new blocks to send a certificate: %s", err.Error())
		return nil
	}
	a.log.Infof("lastL2Block: %d, lasL2BlockSynced: %d  choose range:[%d,%d]",
		lastL2Block.Number.Uint64(), lasL2BlockSynced, fromBlock, toBlock)

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

	a.lock.Lock()
	a.lastL1CertificateBlock = l1Block
	a.lock.Unlock()

	a.log.Infof("certificate: %s sent successfully for range of l2 blocks (from block: %d, to block: %d)",
		certificateHash, fromBlock, toBlock)

	return nil
}

func chooseToL2Block(lastL2Block, lasL2BlockSynced, lastCertificateBlock uint64) (uint64, error) {
	if lastL2Block <= lastCertificateBlock {
		return lastCertificateBlock, fmt.Errorf("lastL2Block: %d (into RPC) is <= to lastCertificateBlock: %d",
			lastL2Block, lastCertificateBlock)
	}
	if lasL2BlockSynced <= lastCertificateBlock {
		return lastCertificateBlock, fmt.Errorf("lasL2BlockSynced: %d (into L2 Syncer) is <= to lastCertificateBlock: %d",
			lastL2Block, lastCertificateBlock)
	}
	if lasL2BlockSynced <= lastL2Block {
		// If lasL2BlockSynced is less than or equal to lastL2Block, means that syncer is not on top of block but
		// we can use it to send a new certificate
		return lasL2BlockSynced, nil
	}
	return lastL2Block, nil
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
		} else {
			previousLocalExitRoot = lastSentCertificateHeader.NewLocalExitRoot
			previousHeight = lastSentCertificateHeader.Height
		}

		lastCertificateBlock, err = a.l2Syncer.GetBlockByLER(ctx, previousLocalExitRoot)
		if err != nil {
			return common.Hash{}, 0, 0, fmt.Errorf("error getting block by LER %s: %w",
				lastSentCertificate.CertificateID, err)
		}
	}

	return previousLocalExitRoot, previousHeight, lastCertificateBlock, nil
}

// getLastL2Block gets the last L2 block based on the finality configured for l2 bridge syncer
func (a *AggSender) getLastL2Block(ctx context.Context) (*types.Header, error) {
	finality := a.l2Syncer.BlockFinality()
	blockFinality, err := finality.ToBlockNum()
	if err != nil {
		return nil, fmt.Errorf("error getting block finality: %w", err)
	}

	return a.l2Client.HeaderByNumber(ctx, blockFinality)
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

	for i, claim := range claims {
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

// shouldSendCertificate checks if a certificate should be sent at given L1 block
// we send certificates at two blocks before the epoch ending so we get most of the
// bridges and claims in that epoch
func (a *AggSender) shouldSendCertificate(block uint64) bool {
	if block == 0 {
		return false
	}

	a.lock.Lock()
	lastL1BlockSeen := a.lastL1CertificateBlock
	a.lock.Unlock()

	shouldSend := lastL1BlockSeen+a.cfg.EpochSize-a.cfg.BlocksBeforeEpochEnding <= block

	return shouldSend
}
