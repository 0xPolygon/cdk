package aggsender

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/bridgesync"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/tree"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ledgerwatch/erigon-lib/kv"
	"github.com/ledgerwatch/erigon-lib/kv/mdbx"
)

const (
	aggSenderDBFolder       = "aggsender"
	sentCertificatesL2Table = "sent_certificates_l2"
)

func tableCfgFunc(defaultBuckets kv.TableCfg) kv.TableCfg {
	return kv.TableCfg{
		sentCertificatesL2Table: {},
	}
}

// AggSender is a component that will send certificates to the aggLayer
type AggSender struct {
	l2Syncer         *bridgesync.BridgeSync
	l2Client         bridgesync.EthClienter
	l1infoTreeSyncer *l1infotreesync.L1InfoTreeSync

	db             kv.RwDB
	aggLayerClient agglayer.AgglayerClientInterface

	sendInterval types.Duration

	sequencerKey *ecdsa.PrivateKey
}

// New returns a new AggSender
func New(
	ctx context.Context,
	cfg Config,
	aggLayerClient agglayer.AgglayerClientInterface,
	l1InfoTreeSyncer *l1infotreesync.L1InfoTreeSync,
	l2Syncer *bridgesync.BridgeSync,
	l2Client bridgesync.EthClienter) (*AggSender, error) {
	db, err := mdbx.NewMDBX(nil).
		Path(filepath.Join(cfg.DBPath, aggSenderDBFolder)).
		WithTableCfg(tableCfgFunc).
		Open()
	if err != nil {
		return nil, err
	}

	sequencerPrivateKey, err := cdkcommon.NewKeyFromKeystore(cfg.SequencerPrivateKey)
	if err != nil {
		return nil, err
	}

	return &AggSender{
		db:               db,
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
	go a.sendCertificates(ctx)
}

// sendCertificates sends certificates to the aggLayer
func (a *AggSender) sendCertificates(ctx context.Context) {
	ticker := time.NewTicker(a.sendInterval.Duration)

	for {
		select {
		case <-ticker.C:
			if err := a.sendCertificatesForNetwork(ctx); err != nil {
				log.Error(err)
			}
		case <-ctx.Done():
			log.Info("AggSender stopped")
			return
		}
	}
}

// buildCertificate builds a certificate from the bridge events
func (a *AggSender) buildCertificate(ctx context.Context,
	bridges []bridgesync.Bridge,
	claims []bridgesync.Claim,
	previousExitRoot common.Hash, lastHeight uint64) (*agglayer.Certificate, error) {
	bridgeExits := make([]*agglayer.BridgeExit, 0, len(bridges))
	importedBridgeExits := make([]*agglayer.ImportedBridgeExit, 0, len(claims))

	for _, bridge := range bridges {
		bridgeExit := convertBridgeToBridgeExit(bridge)
		bridgeExits = append(bridgeExits, bridgeExit)
	}

	for _, claim := range claims {
		importedBridgeExit, err := convertClaimToImportedBridgeExit(claim)
		if err != nil {
			return nil, fmt.Errorf("error converting claim to imported bridge exit: %w", err)
		}

		importedBridgeExits = append(importedBridgeExits, importedBridgeExit)
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

func convertBridgeToBridgeExit(bridge bridgesync.Bridge) *agglayer.BridgeExit {
	return &agglayer.BridgeExit{
		LeafType: agglayer.LeafType(bridge.LeafType),
		TokenInfo: &agglayer.TokenInfo{
			OriginNetwork:      bridge.OriginNetwork,
			OriginTokenAddress: bridge.OriginAddress,
		},
		DestinationNetwork: bridge.DestinationNetwork,
		DestinationAddress: bridge.DestinationAddress,
		Amount:             bridge.Amount,
		Metadata:           bridge.Metadata,
	}
}

func convertClaimToImportedBridgeExit(claim bridgesync.Claim) (*agglayer.ImportedBridgeExit, error) {
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
		ImportedLocalExitRoot: tree.CalculateRoot(bridgeExit.Hash(),
			claim.ProofLocalExitRoot, uint32(claim.GlobalIndex.Uint64())),
		InclusionProof:    claim.ProofLocalExitRoot,
		InclusionProofRER: claim.ProofRollupExitRoot,
		GlobalIndex: &agglayer.GlobalIndex{
			MainnetFlag: mainnetFlag,
			RollupIndex: rollupIndex,
			LeafIndex:   leafIndex,
		},
	}, nil
}

// sendCertificatesForNetwork sends certificates for a network
func (a *AggSender) sendCertificatesForNetwork(ctx context.Context) error {
	lastSentCertificateBlock, lastSentCertificate, err := a.getLastSentCertificate(ctx)
	if err != nil {
		return fmt.Errorf("error getting last sent certificate: %w", err)
	}

	finality := a.l2Syncer.BlockFinality()
	blockFinality, err := finality.ToBlockNum()
	if err != nil {
		return fmt.Errorf("error getting block finality: %w", err)
	}

	lastFinalizedBlock, err := a.l2Client.HeaderByNumber(ctx, blockFinality)
	if err != nil {
		return fmt.Errorf("error getting block number: %w", err)
	}

	bridges, err := a.l2Syncer.GetBridges(ctx, lastSentCertificateBlock+1, lastFinalizedBlock.Nonce.Uint64())
	if err != nil {
		return fmt.Errorf("error getting bridges: %w", err)
	}

	claims, err := a.l2Syncer.GetClaims(ctx, lastSentCertificateBlock+1, lastFinalizedBlock.Nonce.Uint64())
	if err != nil {
		return fmt.Errorf("error getting claims: %w", err)
	}

	if len(bridges) == 0 && len(claims) == 0 {
		// nothing to send
		return nil
	}

	previousExitRoot := common.Hash{}
	lastHeight := uint64(0)
	if lastSentCertificate != nil {
		previousExitRoot = lastSentCertificate.NewLocalExitRoot
		lastHeight = lastSentCertificate.Height
	}

	certificate, err := a.buildCertificate(ctx, bridges, claims, previousExitRoot, lastHeight)
	if err != nil {
		return fmt.Errorf("error building certificate: %w", err)
	}

	signedCertificate, err := a.signCertificate(certificate)
	if err != nil {
		return fmt.Errorf("error signing certificate: %w", err)
	}

	if err := a.aggLayerClient.SendCertificate(signedCertificate); err != nil {
		return fmt.Errorf("error sending certificate: %w", err)
	}

	if err := a.saveLastSentCertificate(ctx, lastFinalizedBlock.Nonce.Uint64(), certificate); err != nil {
		return fmt.Errorf("error saving last sent certificate in db: %w", err)
	}

	return nil
}

// saveLastSentCertificate saves the last sent certificate
func (a *AggSender) saveLastSentCertificate(ctx context.Context, blockNum uint64,
	certificate *agglayer.Certificate) error {
	return a.db.Update(ctx, func(tx kv.RwTx) error {
		raw, err := json.Marshal(certificate)
		if err != nil {
			return err
		}

		return tx.Put(sentCertificatesL2Table, cdkcommon.Uint64ToBytes(blockNum), raw)
	})
}

// getLastSentCertificate returns the last sent certificate
func (a *AggSender) getLastSentCertificate(ctx context.Context) (uint64, *agglayer.Certificate, error) {
	var (
		lastSentCertificateBlock uint64
		lastCertificate          *agglayer.Certificate
	)

	err := a.db.View(ctx, func(tx kv.Tx) error {
		cursor, err := tx.Cursor(sentCertificatesL2Table)
		if err != nil {
			return err
		}

		k, v, err := cursor.Last()
		if err != nil {
			return err
		}

		if k != nil {
			lastSentCertificateBlock = cdkcommon.BytesToUint64(k)
			if err := json.Unmarshal(v, &lastCertificate); err != nil {
				return err
			}
		}

		return nil
	})

	return lastSentCertificateBlock, lastCertificate, err
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
