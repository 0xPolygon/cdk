package datacommittee

import (
	"crypto/ecdsa"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strings"

	"github.com/0xPolygon/cdk-contracts-tooling/contracts/banana/polygondatacommittee"
	"github.com/0xPolygon/cdk-data-availability/client"
	daTypes "github.com/0xPolygon/cdk-data-availability/types"
	"github.com/0xPolygon/cdk/etherman"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygonHermez/zkevm-synchronizer-l1/translator"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/net/context"
)

const (
	unexpectedHashTemplate = "missmatch on transaction data. Expected hash %s, actual hash: %s"
	translateContextName   = "dataCommittee"
)

// DataCommitteeMember represents a member of the Data Committee
type DataCommitteeMember struct {
	Addr common.Address
	URL  string
}

// DataCommittee represents a specific committee
type DataCommittee struct {
	AddressesHash      common.Hash
	Members            []DataCommitteeMember
	RequiredSignatures uint64
}

// Backend implements the DAC integration
type Backend struct {
	logger                     *log.Logger
	dataCommitteeContract      *polygondatacommittee.Polygondatacommittee
	privKey                    *ecdsa.PrivateKey
	dataCommitteeClientFactory client.Factory

	committeeMembers        []DataCommitteeMember
	selectedCommitteeMember int
	ctx                     context.Context
	Translator              translator.Translator
}

// New creates an instance of Backend
func New(
	logger *log.Logger,
	l1RPCURL string,
	dataCommitteeAddr common.Address,
	privKey *ecdsa.PrivateKey,
	dataCommitteeClientFactory client.Factory,
	translator translator.Translator,
) (*Backend, error) {
	ethClient, err := ethclient.Dial(l1RPCURL)
	if err != nil {
		logger.Errorf("error connecting to %s: %+v", l1RPCURL, err)
		return nil, err
	}

	dataCommittee, err := polygondatacommittee.NewPolygondatacommittee(dataCommitteeAddr, ethClient)
	if err != nil {
		return nil, err
	}

	return &Backend{
		logger:                     logger,
		dataCommitteeContract:      dataCommittee,
		privKey:                    privKey,
		dataCommitteeClientFactory: dataCommitteeClientFactory,
		ctx:                        context.Background(),
		Translator:                 translator,
	}, nil
}

// Init loads the DAC to be cached when needed
func (d *Backend) Init() error {
	committee, err := d.getCurrentDataCommittee()
	if err != nil {
		return err
	}
	selectedCommitteeMember := -1
	if committee != nil {
		d.committeeMembers = committee.Members
		if len(committee.Members) > 0 {
			nBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(committee.Members))))
			if err != nil {
				return err
			}
			selectedCommitteeMember = int(nBig.Int64())
		}
	}
	d.selectedCommitteeMember = selectedCommitteeMember

	return nil
}

// GetSequence retrieves backend data by querying committee members for each hash concurrently.
func (d *Backend) GetSequence(_ context.Context, hashes []common.Hash, _ []byte) ([][]byte, error) {
	initialMember := d.selectedCommitteeMember

	var batchData [][]byte
	for retries := 0; retries < len(d.committeeMembers); retries++ {
		member := d.committeeMembers[d.selectedCommitteeMember]
		d.logger.Infof("trying to get data from %s at %s", member.Addr.Hex(), member.URL)

		c := d.dataCommitteeClientFactory.New(member.URL)
		dataMap, err := c.ListOffChainData(d.ctx, hashes)
		if err != nil {
			d.logger.Warnf("error getting data from DAC node %s at %s: %s", member.Addr.Hex(), member.URL, err)
			d.selectedCommitteeMember = (d.selectedCommitteeMember + 1) % len(d.committeeMembers)
			if d.selectedCommitteeMember == initialMember {
				break
			}
			continue
		}

		batchData = make([][]byte, 0, len(hashes))
		for _, hash := range hashes {
			actualTransactionsHash := crypto.Keccak256Hash(dataMap[hash])
			if actualTransactionsHash != hash {
				unexpectedHash := fmt.Errorf(unexpectedHashTemplate, hash, actualTransactionsHash)
				d.logger.Warnf("error getting data from DAC node %s at %s: %s", member.Addr.Hex(), member.URL, unexpectedHash)
				d.selectedCommitteeMember = (d.selectedCommitteeMember + 1) % len(d.committeeMembers)
				if d.selectedCommitteeMember == initialMember {
					break
				}
				continue
			}
			batchData = append(batchData, dataMap[hash])
		}
		return batchData, nil
	}

	if err := d.Init(); err != nil {
		return nil, fmt.Errorf("error loading data committee: %w", err)
	}

	return nil, fmt.Errorf("couldn't get the data from any committee member")
}

type signatureMsg struct {
	addr      common.Address
	signature []byte
	err       error
}

// PostSequenceElderberry submits batches and collects signatures from committee members.
func (d *Backend) PostSequenceElderberry(ctx context.Context, batchesData [][]byte) ([]byte, error) {
	// Get current committee
	committee, err := d.getCurrentDataCommittee()
	if err != nil {
		return nil, err
	}

	// Authenticate as trusted sequencer by signing the sequence
	sequence := make(daTypes.Sequence, 0, len(batchesData))
	for _, batchData := range batchesData {
		sequence = append(sequence, batchData)
	}
	signedSequence, err := sequence.Sign(d.privKey)
	if err != nil {
		return nil, err
	}

	// Request signatures to all members in parallel
	ch := make(chan signatureMsg, len(committee.Members))
	signatureCtx, cancelSignatureCollection := context.WithCancel(ctx)
	for _, member := range committee.Members {
		signedSequenceElderberry := daTypes.SignedSequence{
			Sequence:  sequence,
			Signature: signedSequence,
		}
		go d.requestSignatureFromMember(signatureCtx, &signedSequenceElderberry,
			func(c client.Client) ([]byte, error) { return c.SignSequence(ctx, signedSequenceElderberry) }, member, ch)
	}
	return d.collectSignatures(committee, ch, cancelSignatureCollection)
}

// PostSequenceBanana submits a sequence to the data committee and collects the signed response from them.
func (d *Backend) PostSequenceBanana(ctx context.Context, sequence etherman.SequenceBanana) ([]byte, error) {
	// Get current committee
	committee, err := d.getCurrentDataCommittee()
	if err != nil {
		return nil, err
	}

	sequenceBatches := make([]daTypes.Batch, 0, len(sequence.Batches))
	for _, batch := range sequence.Batches {
		sequenceBatches = append(sequenceBatches, daTypes.Batch{
			L2Data:            batch.L2Data,
			Coinbase:          batch.LastCoinbase,
			ForcedBlockHashL1: batch.ForcedBlockHashL1,
			ForcedGER:         batch.ForcedGlobalExitRoot,
			ForcedTimestamp:   daTypes.ArgUint64(batch.ForcedBatchTimestamp),
		})
	}

	sequenceBanana := daTypes.SequenceBanana{
		Batches:              sequenceBatches,
		OldAccInputHash:      sequence.OldAccInputHash,
		L1InfoRoot:           sequence.L1InfoRoot,
		MaxSequenceTimestamp: daTypes.ArgUint64(sequence.MaxSequenceTimestamp),
	}
	hashToSign := common.BytesToHash(sequenceBanana.HashToSign())
	if hashToSign != sequence.AccInputHash {
		return nil, fmt.Errorf(
			"calculated accInputHash diverges: DA = %s vs Seq = %s",
			hashToSign, sequence.AccInputHash,
		)
	}

	signature, err := sequenceBanana.Sign(d.privKey)
	if err != nil {
		return nil, err
	}

	// Request signatures to all members in parallel
	ch := make(chan signatureMsg, len(committee.Members))
	signatureCtx, cancelSignatureCollection := context.WithCancel(ctx)
	for _, member := range committee.Members {
		signedSequenceBanana := daTypes.SignedSequenceBanana{
			Sequence:  sequenceBanana,
			Signature: signature,
		}
		go d.requestSignatureFromMember(signatureCtx,
			&signedSequenceBanana,
			func(c client.Client) ([]byte, error) { return c.SignSequenceBanana(ctx, signedSequenceBanana) },
			member, ch)
	}

	return d.collectSignatures(committee, ch, cancelSignatureCollection)
}

func (d *Backend) collectSignatures(
	committee *DataCommittee, ch chan signatureMsg, cancelSignatureCollection context.CancelFunc) ([]byte, error) {
	// Collect signatures
	// Stop requesting as soon as we have N valid signatures
	var (
		msgs                = make(signatureMsgs, 0, len(committee.Members))
		collectedSignatures uint64
		failedToCollect     uint64
	)
	for collectedSignatures < committee.RequiredSignatures {
		msg := <-ch
		if msg.err != nil {
			d.logger.Errorf("error when trying to get signature from %s: %s", msg.addr, msg.err)
			failedToCollect++
			if len(committee.Members)-int(failedToCollect) < int(committee.RequiredSignatures) {
				cancelSignatureCollection()

				return nil, errors.New("too many members failed to send their signature")
			}
		} else {
			d.logger.Infof("received signature from %s", msg.addr)
			collectedSignatures++
		}
		msgs = append(msgs, msg)
	}

	cancelSignatureCollection()

	return d.buildSignaturesAndAddrs(msgs, committee.Members), nil
}

type funcSignType func(c client.Client) ([]byte, error)

// funcSetSignatureType: is not possible to define a SetSignature function because
// the type daTypes.SequenceBanana and daTypes.Sequence belong to different packages
// So a future refactor is define a common interface for both
func (d *Backend) requestSignatureFromMember(ctx context.Context, signedSequence daTypes.SignedSequenceInterface,
	funcSign funcSignType,
	member DataCommitteeMember, ch chan signatureMsg) {
	select {
	case <-ctx.Done():
		return
	default:
	}

	// request
	c := client.New(member.URL)
	d.logger.Infof("sending request to sign the sequence to %s at %s", member.Addr.Hex(), member.URL)
	// funcSign must call something like that  c.SignSequenceBanana(ctx, signedSequence)
	signature, err := funcSign(c)

	if err != nil {
		ch <- signatureMsg{
			addr: member.Addr,
			err:  err,
		}

		return
	}
	// verify returned signature
	signedSequence.SetSignature(signature)
	signer, err := signedSequence.Signer()
	if err != nil {
		ch <- signatureMsg{
			addr: member.Addr,
			err:  err,
		}

		return
	}
	if signer != member.Addr {
		ch <- signatureMsg{
			addr: member.Addr,
			err:  fmt.Errorf("invalid signer. Expected %s, actual %s", member.Addr.Hex(), signer.Hex()),
		}

		return
	}
	ch <- signatureMsg{
		addr:      member.Addr,
		signature: signature,
	}
}

func (d *Backend) buildSignaturesAndAddrs(sigs signatureMsgs, members []DataCommitteeMember) []byte {
	const (
		sigLen = 65
	)
	res := make([]byte, 0, len(sigs)*sigLen+len(members)*common.AddressLength)
	sort.Sort(sigs)
	for _, msg := range sigs {
		d.logger.Debugf("adding signature %s from %s", common.Bytes2Hex(msg.signature), msg.addr.Hex())
		res = append(res, msg.signature...)
	}
	for _, member := range members {
		d.logger.Debugf("adding addr %s", common.Bytes2Hex(member.Addr.Bytes()))
		res = append(res, member.Addr.Bytes()...)
	}
	d.logger.Debugf("full res %s", common.Bytes2Hex(res))
	return res
}

type signatureMsgs []signatureMsg

func (s signatureMsgs) Len() int { return len(s) }
func (s signatureMsgs) Less(i, j int) bool {
	return strings.ToUpper(s[i].addr.Hex()) < strings.ToUpper(s[j].addr.Hex())
}
func (s signatureMsgs) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// getCurrentDataCommittee return the currently registered data committee
func (d *Backend) getCurrentDataCommittee() (*DataCommittee, error) {
	addrsHash, err := d.dataCommitteeContract.CommitteeHash(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting CommitteeHash from L1 SC: %w", err)
	}

	reqSign, err := d.dataCommitteeContract.RequiredAmountOfSignatures(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting RequiredAmountOfSignatures from L1 SC: %w", err)
	}

	members, err := d.getCurrentDataCommitteeMembers()
	if err != nil {
		return nil, err
	}

	return &DataCommittee{
		AddressesHash:      addrsHash,
		RequiredSignatures: reqSign.Uint64(),
		Members:            members,
	}, nil
}

// getCurrentDataCommitteeMembers return the currently registered data committee members
func (d *Backend) getCurrentDataCommitteeMembers() ([]DataCommitteeMember, error) {
	nMembers, err := d.dataCommitteeContract.GetAmountOfMembers(&bind.CallOpts{Pending: false})
	if err != nil {
		return nil, fmt.Errorf("error getting GetAmountOfMembers from L1 SC: %w", err)
	}
	members := make([]DataCommitteeMember, 0, nMembers.Int64())
	for i := int64(0); i < nMembers.Int64(); i++ {
		member, err := d.dataCommitteeContract.Members(&bind.CallOpts{Pending: false}, big.NewInt(i))
		if err != nil {
			return nil, fmt.Errorf("error getting Members %d from L1 SC: %w", i, err)
		}
		if d.Translator != nil {
			member.Url = d.Translator.Translate(translateContextName, member.Url)
		}
		members = append(members, DataCommitteeMember{
			Addr: member.Addr,
			URL:  member.Url,
		})
	}

	return members, nil
}
