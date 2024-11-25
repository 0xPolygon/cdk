package aggsender

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggsender/db"
	"github.com/0xPolygon/cdk/aggsender/types"
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
	log types.Logger

	l2Syncer         types.L2BridgeSyncer
	l1infoTreeSyncer types.L1InfoTreeSyncer
	epochNotifier    types.EpochNotifier

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
	l2Syncer types.L2BridgeSyncer,
	epochNotifier types.EpochNotifier) (*AggSender, error) {
	storage, err := db.NewAggSenderSQLStorage(logger, cfg.StoragePath)
	if err != nil {
		return nil, err
	}

	sequencerPrivateKey, err := cdkcommon.NewKeyFromKeystore(cfg.AggsenderPrivateKey)
	if err != nil {
		return nil, err
	}

	logger.Infof("Aggsender Config: %s.", cfg.String())

	return &AggSender{
		cfg:              cfg,
		log:              logger,
		storage:          storage,
		l2Syncer:         l2Syncer,
		aggLayerClient:   aggLayerClient,
		l1infoTreeSyncer: l1InfoTreeSyncer,
		sequencerKey:     sequencerPrivateKey,
		epochNotifier:    epochNotifier,
	}, nil
}

// Start starts the AggSender
func (a *AggSender) Start(ctx context.Context) {
	a.log.Info("AggSender started")
	a.checkInitialStatus(ctx)
	a.sendCertificates(ctx)
}

// checkInitialStatus check local status vs agglayer status
func (a *AggSender) checkInitialStatus(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.DelayBeetweenRetries.Duration)
	defer ticker.Stop()

	for {
		if err := a.checkLastCertificateFromAgglayer(ctx); err != nil {
			a.log.Errorf("error checking initial status: %w, retrying in %s", err, a.cfg.DelayBeetweenRetries.String())
		} else {
			a.log.Info("Initial status checked successfully")
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// sendCertificates sends certificates to the aggLayer
func (a *AggSender) sendCertificates(ctx context.Context) {
	chEpoch := a.epochNotifier.Subscribe("aggsender")
	for {
		select {
		case epoch := <-chEpoch:
			a.log.Infof("Epoch received: %s", epoch.String())
			thereArePendingCerts := a.checkPendingCertificatesStatus(ctx)
			if !thereArePendingCerts {
				if _, err := a.sendCertificate(ctx); err != nil {
					a.log.Error(err)
				}
			} else {
				log.Infof("Skipping epoch %s because there are pending certificates",
					epoch.String())
			}
		case <-ctx.Done():
			a.log.Info("AggSender stopped")
			return
		}
	}
}

// sendCertificate sends certificate for a network
func (a *AggSender) sendCertificate(ctx context.Context) (*agglayer.SignedCertificate, error) {
	a.log.Infof("trying to send a new certificate...")

	shouldSend, err := a.shouldSendCertificate()
	if err != nil {
		return nil, err
	}

	if !shouldSend {
		a.log.Infof("waiting for pending certificates to be settled")
		return nil, nil
	}

	lasL2BlockSynced, err := a.l2Syncer.GetLastProcessedBlock(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting last processed block from l2: %w", err)
	}

	lastSentCertificateInfo, err := a.storage.GetLastSentCertificate()
	if err != nil {
		return nil, err
	}
	if lastSentCertificateInfo == nil {
		// There are no certificates, so we set that to a empty one
		lastSentCertificateInfo = &types.CertificateInfo{}
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
		return nil, nil
	}

	fromBlock := previousToBlock + 1
	toBlock := lasL2BlockSynced

	bridges, err := a.l2Syncer.GetBridgesPublished(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("error getting bridges: %w", err)
	}

	if len(bridges) == 0 {
		a.log.Infof("no bridges consumed, no need to send a certificate from block: %d to block: %d", fromBlock, toBlock)
		return nil, nil
	}

	claims, err := a.l2Syncer.GetClaims(ctx, fromBlock, toBlock)
	if err != nil {
		return nil, fmt.Errorf("error getting claims: %w", err)
	}

	a.log.Infof("building certificate for block: %d to block: %d", fromBlock, toBlock)

	certificate, err := a.buildCertificate(ctx, bridges, claims, *lastSentCertificateInfo, toBlock)
	if err != nil {
		return nil, fmt.Errorf("error building certificate: %w", err)
	}

	signedCertificate, err := a.signCertificate(certificate)
	if err != nil {
		return nil, fmt.Errorf("error signing certificate: %w", err)
	}

	a.saveCertificateToFile(signedCertificate)
	a.log.Infof("certificate ready to be send to AggLayer: %s", signedCertificate.String())
	// a.log.Fatalf("certificate ready to be send to AggLayer: %s", signedCertificate.String())
	certificateHash := common.Hash{}
	//	certificateHash, err := a.aggLayerClient.SendCertificate(signedCertificate)
	if err != nil {
		return nil, fmt.Errorf("error sending certificate: %w", err)
	}

	a.log.Debugf("certificate send: Height: %d hash: %s", signedCertificate.Height, certificateHash.String())

	raw, err := json.Marshal(signedCertificate)
	if err != nil {
		return nil, fmt.Errorf("error marshalling signed certificate: %w", err)
	}

	createdTime := time.Now().UTC().UnixMilli()
	certInfo := types.CertificateInfo{
		Height:            certificate.Height,
		CertificateID:     certificateHash,
		NewLocalExitRoot:  certificate.NewLocalExitRoot,
		FromBlock:         fromBlock,
		ToBlock:           toBlock,
		CreatedAt:         createdTime,
		UpdatedAt:         createdTime,
		SignedCertificate: string(raw),
	}
	// TODO: Improve this case, if a cert is not save in the storage, we are going to settle a unknown certificate
	err = a.saveCertificateToStorage(ctx, certInfo, a.cfg.MaxRetriesStoreCertificate)
	if err != nil {
		a.log.Errorf("error saving certificate to storage: %w", err)
		return nil, fmt.Errorf("error saving last sent certificate %s in db: %w", certInfo.String(), err)
	}

	a.log.Infof("certificate: %s sent successfully for range of l2 blocks (from block: %d, to block: %d) cert:%s",
		certificateHash, fromBlock, toBlock, signedCertificate.String())

	return signedCertificate, nil
}

// saveCertificateToStorage saves the certificate to the storage
// it retries if it fails. if param retries == 0 it retries indefinitely
func (a *AggSender) saveCertificateToStorage(ctx context.Context, cert types.CertificateInfo, maxRetries int) error {
	retries := 1
	err := fmt.Errorf("initial_error")
	for err != nil {
		if err = a.storage.SaveLastSentCertificate(ctx, cert); err != nil {
			// If this happens we can't work as normal, because local DB is outdated, we have to retry
			a.log.Errorf("error saving last sent certificate %s in db: %w", cert.String(), err)
			if retries == maxRetries {
				return fmt.Errorf("error saving last sent certificate %s in db: %w", cert.String(), err)
			} else {
				retries++
				time.Sleep(a.cfg.DelayBeetweenRetries.Duration)
			}
		}
	}
	return nil
}

// saveCertificate saves the certificate to a tmp file
func (a *AggSender) saveCertificateToFile(signedCertificate *agglayer.SignedCertificate) {
	if signedCertificate == nil || a.cfg.SaveCertificatesToFilesPath == "" {
		return
	}
	fn := fmt.Sprintf("%s/certificate_%04d-%07d.json",
		a.cfg.SaveCertificatesToFilesPath, signedCertificate.Height, time.Now().Unix())
	a.log.Infof("saving certificate to file: %s", fn)
	jsonData, err := json.MarshalIndent(signedCertificate, "", "  ")
	if err != nil {
		a.log.Errorf("error marshalling certificate: %w", err)
	}

	if err = os.WriteFile(fn, jsonData, 0644); err != nil { //nolint:gosec,mnd // we are writing to a tmp file
		a.log.Errorf("error writing certificate to file: %w", err)
	}
}

// getNextHeightAndPreviousLER returns the height and previous LER for the new certificate
func (a *AggSender) getNextHeightAndPreviousLER(
	lastSentCertificateInfo *types.CertificateInfo) (uint64, common.Hash) {
	height := lastSentCertificateInfo.Height + 1
	if lastSentCertificateInfo.Status.IsInError() {
		// previous certificate was in error, so we need to resend it
		a.log.Debugf("Last certificate %s failed so reusing height %d",
			lastSentCertificateInfo.CertificateID, lastSentCertificateInfo.Height)
		height = lastSentCertificateInfo.Height
	}

	previousLER := lastSentCertificateInfo.NewLocalExitRoot
	if lastSentCertificateInfo.NewLocalExitRoot == (common.Hash{}) ||
		lastSentCertificateInfo.Height == 0 && lastSentCertificateInfo.Status.IsInError() {
		// meaning this is the first certificate
		height = 0
		previousLER = zeroLER
	}

	return height, previousLER
}

// buildCertificate builds a certificate from the bridge events
func (a *AggSender) buildCertificate(ctx context.Context,
	bridges []bridgesync.Bridge,
	claims []bridgesync.Claim,
	lastSentCertificateInfo types.CertificateInfo,
	toBlock uint64) (*agglayer.Certificate, error) {
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

	height, previousLER := a.getNextHeightAndPreviousLER(&lastSentCertificateInfo)

	return &agglayer.Certificate{
		NetworkID:           a.l2Syncer.OriginNetwork(),
		PrevLocalExitRoot:   previousLER,
		NewLocalExitRoot:    exitRoot.Hash,
		BridgeExits:         bridgeExits,
		ImportedBridgeExits: importedBridgeExits,
		Height:              height,
		Metadata:            createCertificateMetadata(toBlock),
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
func (a *AggSender) getImportedBridgeExits(
	ctx context.Context, claims []bridgesync.Claim,
) ([]*agglayer.ImportedBridgeExit, error) {
	if len(claims) == 0 {
		// no claims to convert
		return []*agglayer.ImportedBridgeExit{}, nil
	}

	var (
		greatestL1InfoTreeIndexUsed uint32
		importedBridgeExits         = make([]*agglayer.ImportedBridgeExit, 0, len(claims))
		claimL1Info                 = make([]*l1infotreesync.L1InfoTreeLeaf, 0, len(claims))
	)

	for _, claim := range claims {
		info, err := a.l1infoTreeSyncer.GetInfoByGlobalExitRoot(claim.GlobalExitRoot)
		if err != nil {
			return nil, fmt.Errorf("error getting info by global exit root: %w", err)
		}

		claimL1Info = append(claimL1Info, info)

		if info.L1InfoTreeIndex > greatestL1InfoTreeIndexUsed {
			greatestL1InfoTreeIndexUsed = info.L1InfoTreeIndex
		}
	}

	rootToProve, err := a.l1infoTreeSyncer.GetL1InfoTreeRootByIndex(ctx, greatestL1InfoTreeIndexUsed)
	if err != nil {
		return nil, fmt.Errorf("error getting L1 Info tree root by index: %d. Error: %w", greatestL1InfoTreeIndexUsed, err)
	}

	for i, claim := range claims {
		l1Info := claimL1Info[i]

		a.log.Debugf("claim[%d]: destAddr: %s GER:%s", i, claim.DestinationAddress.String(), claim.GlobalExitRoot.String())
		ibe, err := a.convertClaimToImportedBridgeExit(claim)
		if err != nil {
			return nil, fmt.Errorf("error converting claim to imported bridge exit: %w", err)
		}

		importedBridgeExits = append(importedBridgeExits, ibe)

		gerToL1Proof, err := a.l1infoTreeSyncer.GetL1InfoTreeMerkleProofFromIndexToRoot(
			ctx, l1Info.L1InfoTreeIndex, rootToProve.Hash,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"error getting L1 Info tree merkle proof for leaf index: %d and root: %s. Error: %w",
				l1Info.L1InfoTreeIndex, rootToProve.Hash, err,
			)
		}

		claim := claims[i]
		if ibe.GlobalIndex.MainnetFlag {
			ibe.ClaimData = &agglayer.ClaimFromMainnnet{
				L1Leaf: &agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: l1Info.L1InfoTreeIndex,
					RollupExitRoot:  claim.RollupExitRoot,
					MainnetExitRoot: claim.MainnetExitRoot,
					Inner: &agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: l1Info.GlobalExitRoot,
						Timestamp:      l1Info.Timestamp,
						BlockHash:      l1Info.PreviousBlockHash,
					},
				},
				ProofLeafMER: &agglayer.MerkleProof{
					Root:  claim.MainnetExitRoot,
					Proof: claim.ProofLocalExitRoot,
				},
				ProofGERToL1Root: &agglayer.MerkleProof{
					Root:  rootToProve.Hash,
					Proof: gerToL1Proof,
				},
			}
		} else {
			ibe.ClaimData = &agglayer.ClaimFromRollup{
				L1Leaf: &agglayer.L1InfoTreeLeaf{
					L1InfoTreeIndex: l1Info.L1InfoTreeIndex,
					RollupExitRoot:  claim.RollupExitRoot,
					MainnetExitRoot: claim.MainnetExitRoot,
					Inner: &agglayer.L1InfoTreeLeafInner{
						GlobalExitRoot: l1Info.GlobalExitRoot,
						Timestamp:      l1Info.Timestamp,
						BlockHash:      l1Info.PreviousBlockHash,
					},
				},
				ProofLeafLER: &agglayer.MerkleProof{
					Root:  claim.MainnetExitRoot,
					Proof: claim.ProofLocalExitRoot,
				},
				ProofLERToRER: &agglayer.MerkleProof{
					Root:  claim.RollupExitRoot,
					Proof: claim.ProofRollupExitRoot,
				},
				ProofGERToL1Root: &agglayer.MerkleProof{
					Root:  rootToProve.Hash,
					Proof: gerToL1Proof,
				},
			}
		}
	}

	return importedBridgeExits, nil
}

// signCertificate signs a certificate with the sequencer key
func (a *AggSender) signCertificate(certificate *agglayer.Certificate) (*agglayer.SignedCertificate, error) {
	hashToSign := certificate.HashToSign()

	sig, err := crypto.Sign(hashToSign.Bytes(), a.sequencerKey)
	if err != nil {
		return nil, err
	}

	a.log.Infof("Signed certificate. sequencer address: %s. New local exit root: %s Hash signed: %s",
		crypto.PubkeyToAddress(a.sequencerKey.PublicKey).String(),
		common.BytesToHash(certificate.NewLocalExitRoot[:]).String(),
		hashToSign.String(),
	)

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

// checkPendingCertificatesStatus checks the status of pending certificates
// and updates in the storage if it changed on agglayer
// It returns:
// bool -> if there are pending certificates
func (a *AggSender) checkPendingCertificatesStatus(ctx context.Context) bool {
	pendingCertificates, err := a.storage.GetCertificatesByStatus(agglayer.NonSettledStatuses)
	if err != nil {
		a.log.Errorf("error getting pending certificates: %w", err)
		return true
	}

	a.log.Debugf("checkPendingCertificatesStatus num of pendingCertificates: %d", len(pendingCertificates))
	thereArePendingCerts := false

	for _, certificate := range pendingCertificates {
		certificateHeader, err := a.aggLayerClient.GetCertificateHeader(certificate.CertificateID)
		if err != nil {
			a.log.Errorf("error getting certificate header of %s from agglayer: %w",
				certificate.ID(), err)
			return true
		}

		a.log.Debugf("aggLayerClient.GetCertificateHeader status [%s] of certificate %s  elapsed time:%s",
			certificateHeader.Status,
			certificateHeader.ID(),
			certificate.ElapsedTimeSinceCreation())

		if err := a.updateCertificateStatus(ctx, certificate, certificateHeader); err != nil {
			a.log.Errorf("error updating certificate %s status in storage: %w", certificateHeader.String(), err)
			return true
		}

		if !certificate.IsClosed() {
			a.log.Infof("certificate %s is still pending, elapsed time:%s ",
				certificateHeader.ID(), certificate.ElapsedTimeSinceCreation())
			thereArePendingCerts = true
		}
	}
	return thereArePendingCerts
}

// updateCertificate updates the certificate status in the storage
func (a *AggSender) updateCertificateStatus(ctx context.Context,
	localCert *types.CertificateInfo,
	agglayerCert *agglayer.CertificateHeader) error {
	if localCert.Status == agglayerCert.Status {
		return nil
	}
	a.log.Infof("certificate %s changed status from [%s] to [%s] elapsed time: %s full_cert: %s",
		localCert.ID(), localCert.Status, agglayerCert.Status, localCert.ElapsedTimeSinceCreation(),
		localCert.String())

	// That is a strange situation
	if agglayerCert.Status.IsOpen() && localCert.Status.IsClosed() {
		a.log.Warnf("certificate %s is reopen! from [%s] to [%s]",
			localCert.ID(), localCert.Status, agglayerCert.Status)
	}

	localCert.Status = agglayerCert.Status
	localCert.UpdatedAt = time.Now().UTC().UnixMilli()
	if err := a.storage.UpdateCertificate(ctx, *localCert); err != nil {
		a.log.Errorf("error updating certificate %s status in storage: %w", agglayerCert.ID(), err)
		return fmt.Errorf("error updating certificate. Err: %w", err)
	}
	return nil
}

// shouldSendCertificate checks if a certificate should be sent at given time
// if we have pending certificates, then we wait until they are settled
func (a *AggSender) shouldSendCertificate() (bool, error) {
	pendingCertificates, err := a.storage.GetCertificatesByStatus(agglayer.NonSettledStatuses)
	if err != nil {
		return false, fmt.Errorf("error getting pending certificates: %w", err)
	}

	return len(pendingCertificates) == 0, nil
}

// checkLastCertificateFromAgglayer checks the last certificate from agglayer
func (a *AggSender) checkLastCertificateFromAgglayer(ctx context.Context) error {
	networkID := a.l2Syncer.OriginNetwork()
	a.log.Infof("recovery: checking last certificate from AggLayer for network %d", networkID)
	aggLayerLastCert, err := a.aggLayerClient.GetLatestKnownCertificateHeader(networkID)
	if err != nil {
		return fmt.Errorf("recovery: error getting latest known certificate header from agglayer: %w", err)
	}
	a.log.Infof("recovery: last certificate from AggLayer: %s", aggLayerLastCert.String())
	localLastCert, err := a.storage.GetLastSentCertificate()
	if err != nil {
		return fmt.Errorf("recovery: error getting last sent certificate from local storage: %w", err)
	}
	a.log.Infof("recovery: last certificate in storage: %s", localLastCert.String())

	// CASE 1: No certificates in local storage and agglayer
	if localLastCert == nil && aggLayerLastCert == nil {
		a.log.Info("recovery: No certificates in local storage and agglayer: initial state")
		return nil
	}
	// CASE 2: No certificates in local storage but agglayer has one
	if localLastCert == nil && aggLayerLastCert != nil {
		a.log.Info("recovery: No certificates in local storage but agglayer have one: recovery aggSender cert: %s",
			aggLayerLastCert.String())
		if _, err := a.updateLocalStorageWithAggLayerCert(ctx, aggLayerLastCert); err != nil {
			return fmt.Errorf("recovery: error updating local storage with agglayer certificate: %w", err)
		}
		return nil
	}
	// CASE 3: aggsender stopped between sending to agglayer and storing on DB
	if aggLayerLastCert.Height == localLastCert.Height+1 {
		a.log.Infof("recovery: AggLayer have next cert (height:%d), so is a recovery case: storing cert: %s",
			aggLayerLastCert.Height, localLastCert.String())
		// we need to store the certificate in the local storage.
		localLastCert, err = a.updateLocalStorageWithAggLayerCert(ctx, aggLayerLastCert)
		if err != nil {
			log.Errorf("recovery: error updating status certificate: %s status: %w", aggLayerLastCert.String(), err)
			return fmt.Errorf("recovery: error updating certificate status: %w", err)
		}
	}
	// CASE 4: AggSender and AggLayer are not on the same page
	// note: we don't need to check individual fields of the certificate
	// because CertificateID is a hash of all the fields
	if localLastCert.CertificateID != aggLayerLastCert.CertificateID {
		a.log.Errorf("recovery: Local certificate:\n %s \n is different from agglayer certificate:\n %s",
			localLastCert.String(), aggLayerLastCert.String())
		return fmt.Errorf("recovery: mismatch between local and agglayer certificates")
	}
	// CASE 5: AggSender and AggLayer are at same page
	// just update status
	err = a.updateCertificateStatus(ctx, localLastCert, aggLayerLastCert)
	if err != nil {
		a.log.Errorf("recovery: error updating status certificate: %s status: %w", aggLayerLastCert.String(), err)
		return fmt.Errorf("recovery: error updating certificate status: %w", err)
	}

	a.log.Infof("recovery: successfully checked last certificate from AggLayer for network %d", networkID)
	return nil
}

// updateLocalStorageWithAggLayerCert updates the local storage with the certificate from the AggLayer
func (a *AggSender) updateLocalStorageWithAggLayerCert(ctx context.Context,
	aggLayerCert *agglayer.CertificateHeader) (*types.CertificateInfo, error) {
	certInfo := NewCertificateInfoFromAgglayerCertHeader(aggLayerCert)
	a.log.Infof("setting initial certificate from AggLayer: %s", certInfo.String())
	return certInfo, a.storage.SaveLastSentCertificate(ctx, *certInfo)
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

// createCertificateMetadata creates a certificate metadata from given input
func createCertificateMetadata(toBlock uint64) common.Hash {
	return common.BigToHash(new(big.Int).SetUint64(toBlock))
}

func extractFromCertificateMetadataToBlock(metadata common.Hash) uint64 {
	return metadata.Big().Uint64()
}

func NewCertificateInfoFromAgglayerCertHeader(c *agglayer.CertificateHeader) *types.CertificateInfo {
	if c == nil {
		return nil
	}
	now := time.Now().UTC().UnixMilli()
	return &types.CertificateInfo{
		Height:            c.Height,
		CertificateID:     c.CertificateID,
		NewLocalExitRoot:  c.NewLocalExitRoot,
		FromBlock:         0,
		ToBlock:           extractFromCertificateMetadataToBlock(c.Metadata),
		Status:            c.Status,
		CreatedAt:         now,
		UpdatedAt:         now,
		SignedCertificate: "na/agglayer header",
	}
}
