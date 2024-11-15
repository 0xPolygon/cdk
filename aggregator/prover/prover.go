package prover

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/0xPolygon/cdk/config/types"
	"github.com/0xPolygon/cdk/log"
	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-crypto/poseidon"
)

const (
	stateRootStartIndex    = 19
	stateRootFinalIndex    = stateRootStartIndex + 8
	accInputHashStartIndex = 27
	accInputHashFinalIndex = accInputHashStartIndex + 8
)

var (
	ErrBadProverResponse    = errors.New("prover returned wrong type for response")  //nolint:revive
	ErrProverInternalError  = errors.New("prover returned INTERNAL_ERROR response")  //nolint:revive
	ErrProverCompletedError = errors.New("prover returned COMPLETED_ERROR response") //nolint:revive
	ErrBadRequest           = errors.New("prover returned ERROR for a bad request")  //nolint:revive
	ErrUnspecified          = errors.New("prover returned an UNSPECIFIED response")  //nolint:revive
	ErrUnknown              = errors.New("prover returned an unknown response")      //nolint:revive
	ErrProofCanceled        = errors.New("proof has been canceled")                  //nolint:revive
)

// Prover abstraction of the grpc prover client.
type Prover struct {
	logger                    *log.Logger
	name                      string
	id                        string
	address                   net.Addr
	proofStatePollingInterval types.Duration
	stream                    AggregatorService_ChannelServer
}

// New returns a new Prover instance.
func New(logger *log.Logger, stream AggregatorService_ChannelServer,
	addr net.Addr, proofStatePollingInterval types.Duration) (*Prover, error) {
	p := &Prover{
		logger:                    logger,
		stream:                    stream,
		address:                   addr,
		proofStatePollingInterval: proofStatePollingInterval,
	}

	status, err := p.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve prover id %w", err)
	}
	p.name = status.ProverName
	p.id = status.ProverId

	return p, nil
}

// Name returns the Prover name.
func (p *Prover) Name() string { return p.name }

// ID returns the Prover ID.
func (p *Prover) ID() string { return p.id }

// Addr returns the prover IP address.
func (p *Prover) Addr() string {
	if p.address == nil {
		return ""
	}

	return p.address.String()
}

// Status gets the prover status.
func (p *Prover) Status() (*GetStatusResponse, error) {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_GetStatusRequest{
			GetStatusRequest: &GetStatusRequest{},
		},
	}
	res, err := p.call(req)
	if err != nil {
		return nil, err
	}
	if msg, ok := res.Response.(*ProverMessage_GetStatusResponse); ok {
		return msg.GetStatusResponse, nil
	}

	return nil, fmt.Errorf("%w, wanted %T, got %T", ErrBadProverResponse, &ProverMessage_GetStatusResponse{}, res.Response)
}

// IsIdle returns true if the prover is idling.
func (p *Prover) IsIdle() (bool, error) {
	status, err := p.Status()
	if err != nil {
		return false, err
	}

	return status.Status == GetStatusResponse_STATUS_IDLE, nil
}

// SupportsForkID returns true if the prover supports the given fork id.
func (p *Prover) SupportsForkID(forkID uint64) bool {
	status, err := p.Status()
	if err != nil {
		p.logger.Warnf("Error asking status for prover ID %s: %v", p.ID(), err)
		return false
	}

	p.logger.Debugf("Prover %s supports fork ID %d", p.ID(), status.ForkId)

	return status.ForkId == forkID
}

// BatchProof instructs the prover to generate a batch proof for the provided
// input. It returns the ID of the proof being computed.
func (p *Prover) BatchProof(input *StatelessInputProver) (*string, error) {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_GenStatelessBatchProofRequest{
			GenStatelessBatchProofRequest: &GenStatelessBatchProofRequest{Input: input},
		},
	}
	res, err := p.call(req)
	if err != nil {
		return nil, err
	}

	if msg, ok := res.Response.(*ProverMessage_GenBatchProofResponse); ok {
		switch msg.GenBatchProofResponse.Result {
		case Result_RESULT_UNSPECIFIED:
			return nil, fmt.Errorf(
				"failed to generate proof %s, %w, input %v",
				msg.GenBatchProofResponse.String(), ErrUnspecified, input,
			)
		case Result_RESULT_OK:
			return &msg.GenBatchProofResponse.Id, nil
		case Result_RESULT_ERROR:
			return nil, fmt.Errorf(
				"failed to generate proof %s, %w, input %v",
				msg.GenBatchProofResponse.String(), ErrBadRequest, input,
			)
		case Result_RESULT_INTERNAL_ERROR:
			return nil, fmt.Errorf(
				"failed to generate proof %s, %w, input %v",
				msg.GenBatchProofResponse.String(), ErrProverInternalError, input,
			)
		default:
			return nil, fmt.Errorf(
				"failed to generate proof %s, %w,input %v",
				msg.GenBatchProofResponse.String(), ErrUnknown, input,
			)
		}
	}

	return nil, fmt.Errorf(
		"%w, wanted %T, got %T",
		ErrBadProverResponse, &ProverMessage_GenBatchProofResponse{}, res.Response,
	)
}

// AggregatedProof instructs the prover to generate an aggregated proof from
// the two inputs provided. It returns the ID of the proof being computed.
func (p *Prover) AggregatedProof(inputProof1, inputProof2 string) (*string, error) {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_GenAggregatedProofRequest{
			GenAggregatedProofRequest: &GenAggregatedProofRequest{
				RecursiveProof_1: inputProof1,
				RecursiveProof_2: inputProof2,
			},
		},
	}
	res, err := p.call(req)
	if err != nil {
		return nil, err
	}

	if msg, ok := res.Response.(*ProverMessage_GenAggregatedProofResponse); ok {
		switch msg.GenAggregatedProofResponse.Result {
		case Result_RESULT_UNSPECIFIED:
			return nil, fmt.Errorf("failed to aggregate proofs %s, %w, input 1 %s, input 2 %s",
				msg.GenAggregatedProofResponse.String(), ErrUnspecified, inputProof1, inputProof2)
		case Result_RESULT_OK:
			return &msg.GenAggregatedProofResponse.Id, nil
		case Result_RESULT_ERROR:
			return nil, fmt.Errorf("failed to aggregate proofs %s, %w, input 1 %s, input 2 %s",
				msg.GenAggregatedProofResponse.String(), ErrBadRequest, inputProof1, inputProof2)
		case Result_RESULT_INTERNAL_ERROR:
			return nil, fmt.Errorf("failed to aggregate proofs %s, %w, input 1 %s, input 2 %s",
				msg.GenAggregatedProofResponse.String(), ErrProverInternalError, inputProof1, inputProof2)
		default:
			return nil, fmt.Errorf("failed to aggregate proofs %s, %w, input 1 %s, input 2 %s",
				msg.GenAggregatedProofResponse.String(), ErrUnknown, inputProof1, inputProof2)
		}
	}

	return nil, fmt.Errorf(
		"%w, wanted %T, got %T",
		ErrBadProverResponse, &ProverMessage_GenAggregatedProofResponse{}, res.Response,
	)
}

// FinalProof instructs the prover to generate a final proof for the given
// input. It returns the ID of the proof being computed.
func (p *Prover) FinalProof(inputProof string, aggregatorAddr string) (*string, error) {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_GenFinalProofRequest{
			GenFinalProofRequest: &GenFinalProofRequest{
				RecursiveProof: inputProof,
				AggregatorAddr: aggregatorAddr,
			},
		},
	}
	res, err := p.call(req)
	if err != nil {
		return nil, err
	}

	if msg, ok := res.Response.(*ProverMessage_GenFinalProofResponse); ok {
		switch msg.GenFinalProofResponse.Result {
		case Result_RESULT_UNSPECIFIED:
			return nil, fmt.Errorf("failed to generate final proof %s, %w, input %s",
				msg.GenFinalProofResponse.String(), ErrUnspecified, inputProof)
		case Result_RESULT_OK:
			return &msg.GenFinalProofResponse.Id, nil
		case Result_RESULT_ERROR:
			return nil, fmt.Errorf("failed to generate final proof %s, %w, input %s",
				msg.GenFinalProofResponse.String(), ErrBadRequest, inputProof)
		case Result_RESULT_INTERNAL_ERROR:
			return nil, fmt.Errorf("failed to generate final proof %s, %w, input %s",
				msg.GenFinalProofResponse.String(), ErrProverInternalError, inputProof)
		default:
			return nil, fmt.Errorf("failed to generate final proof %s, %w, input %s",
				msg.GenFinalProofResponse.String(), ErrUnknown, inputProof)
		}
	}

	return nil, fmt.Errorf(
		"%w, wanted %T, got %T",
		ErrBadProverResponse, &ProverMessage_GenFinalProofResponse{}, res.Response,
	)
}

// CancelProofRequest asks the prover to stop the generation of the proof
// matching the provided proofID.
func (p *Prover) CancelProofRequest(proofID string) error {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_CancelRequest{
			CancelRequest: &CancelRequest{Id: proofID},
		},
	}
	res, err := p.call(req)
	if err != nil {
		return err
	}
	if msg, ok := res.Response.(*ProverMessage_CancelResponse); ok {
		switch msg.CancelResponse.Result {
		case Result_RESULT_UNSPECIFIED:
			return fmt.Errorf("failed to cancel proof id [%s], %w, %s",
				proofID, ErrUnspecified, msg.CancelResponse.String())
		case Result_RESULT_OK:
			return nil
		case Result_RESULT_ERROR:
			return fmt.Errorf("failed to cancel proof id [%s], %w, %s",
				proofID, ErrBadRequest, msg.CancelResponse.String())
		case Result_RESULT_INTERNAL_ERROR:
			return fmt.Errorf("failed to cancel proof id [%s], %w, %s",
				proofID, ErrProverInternalError, msg.CancelResponse.String())
		default:
			return fmt.Errorf("failed to cancel proof id [%s], %w, %s",
				proofID, ErrUnknown, msg.CancelResponse.String())
		}
	}

	return fmt.Errorf("%w, wanted %T, got %T", ErrBadProverResponse, &ProverMessage_CancelResponse{}, res.Response)
}

// WaitRecursiveProof waits for a recursive proof to be generated by the prover
// and returns it.
func (p *Prover) WaitRecursiveProof(ctx context.Context, proofID string) (string, common.Hash, common.Hash, error) {
	res, err := p.waitProof(ctx, proofID)
	if err != nil {
		return "", common.Hash{}, common.Hash{}, err
	}

	resProof, ok := res.Proof.(*GetProofResponse_RecursiveProof)
	if !ok {
		return "", common.Hash{}, common.Hash{}, fmt.Errorf(
			"%w, wanted %T, got %T",
			ErrBadProverResponse, &GetProofResponse_RecursiveProof{}, res.Proof,
		)
	}

	sr, err := GetSanityCheckHashFromProof(p.logger, resProof.RecursiveProof, stateRootStartIndex, stateRootFinalIndex)
	if err != nil && sr != (common.Hash{}) {
		p.logger.Errorf("Error getting state root from proof: %v", err)
	}

	accInputHash, err := GetSanityCheckHashFromProof(p.logger, resProof.RecursiveProof,
		accInputHashStartIndex, accInputHashFinalIndex)
	if err != nil && accInputHash != (common.Hash{}) {
		p.logger.Errorf("Error getting acc input hash from proof: %v", err)
	}

	if sr == (common.Hash{}) {
		p.logger.Info("Recursive proof does not contain state root. Possibly mock prover is in use.")
	}

	return resProof.RecursiveProof, sr, accInputHash, nil
}

// WaitFinalProof waits for the final proof to be generated by the prover and
// returns it.
func (p *Prover) WaitFinalProof(ctx context.Context, proofID string) (*FinalProof, error) {
	res, err := p.waitProof(ctx, proofID)
	if err != nil {
		return nil, err
	}
	resProof, ok := res.Proof.(*GetProofResponse_FinalProof)
	if !ok {
		return nil, fmt.Errorf("%w, wanted %T, got %T", ErrBadProverResponse, &GetProofResponse_FinalProof{}, res.Proof)
	}

	return resProof.FinalProof, nil
}

// waitProof waits for a proof to be generated by the prover and returns the
// prover response.
func (p *Prover) waitProof(ctx context.Context, proofID string) (*GetProofResponse, error) {
	req := &AggregatorMessage{
		Request: &AggregatorMessage_GetProofRequest{
			GetProofRequest: &GetProofRequest{
				// TODO(pg): set Timeout field?
				Id: proofID,
			},
		},
	}

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			res, err := p.call(req)
			if err != nil {
				return nil, err
			}
			if msg, ok := res.Response.(*ProverMessage_GetProofResponse); ok {
				switch msg.GetProofResponse.Result {
				case GetProofResponse_RESULT_PENDING:
					time.Sleep(p.proofStatePollingInterval.Duration)

					continue
				case GetProofResponse_RESULT_UNSPECIFIED:
					return nil, fmt.Errorf("failed to get proof ID: %s, %w, prover response: %s",
						proofID, ErrUnspecified, msg.GetProofResponse.String())
				case GetProofResponse_RESULT_COMPLETED_OK:
					return msg.GetProofResponse, nil
				case GetProofResponse_RESULT_ERROR:
					return nil, fmt.Errorf("failed to get proof with ID %s, %w, prover response: %s",
						proofID, ErrBadRequest, msg.GetProofResponse.String())
				case GetProofResponse_RESULT_COMPLETED_ERROR:
					return nil, fmt.Errorf("failed to get proof with ID %s, %w, prover response: %s",
						proofID, ErrProverCompletedError, msg.GetProofResponse.String())
				case GetProofResponse_RESULT_INTERNAL_ERROR:
					return nil, fmt.Errorf("failed to get proof ID: %s, %w, prover response: %s",
						proofID, ErrProverInternalError, msg.GetProofResponse.String())
				case GetProofResponse_RESULT_CANCEL:
					return nil, fmt.Errorf("proof generation was cancelled for proof ID %s, %w, prover response: %s",
						proofID, ErrProofCanceled, msg.GetProofResponse.String())
				default:
					return nil, fmt.Errorf("failed to get proof ID: %s, %w, prover response: %s",
						proofID, ErrUnknown, msg.GetProofResponse.String())
				}
			}

			return nil, fmt.Errorf(
				"%w, wanted %T, got %T",
				ErrBadProverResponse, &ProverMessage_GetProofResponse{}, res.Response,
			)
		}
	}
}

// call sends a message to the prover and waits to receive the response over
// the connection stream.
func (p *Prover) call(req *AggregatorMessage) (*ProverMessage, error) {
	if err := p.stream.Send(req); err != nil {
		return nil, err
	}
	res, err := p.stream.Recv()
	if err != nil {
		return nil, err
	}

	return res, nil
}

// GetSanityCheckHashFromProof returns info from the proof
func GetSanityCheckHashFromProof(logger *log.Logger, proof string, startIndex, endIndex int) (common.Hash, error) {
	type Publics struct {
		Publics []string `mapstructure:"publics"`
	}

	// Check if the proof contains the SR
	if !strings.Contains(proof, "publics") {
		return common.Hash{}, nil
	}

	var publics Publics
	err := json.Unmarshal([]byte(proof), &publics)
	if err != nil {
		logger.Errorf("Error unmarshalling proof: %v", err)
		return common.Hash{}, err
	}

	var (
		v [8]uint64
		j = 0
	)
	for i := startIndex; i < endIndex; i++ {
		u64, err := strconv.ParseInt(publics.Publics[i], 10, 64)
		if err != nil {
			logger.Fatal(err)
		}
		v[j] = uint64(u64)
		j++
	}
	bigSR := fea2scalar(v[:])
	hexSR := fmt.Sprintf("%x", bigSR)
	if len(hexSR)%2 != 0 {
		hexSR = "0" + hexSR
	}

	return common.HexToHash(hexSR), nil
}

// fea2scalar converts array of uint64 values into one *big.Int.
func fea2scalar(v []uint64) *big.Int {
	if len(v) != poseidon.NROUNDSF {
		return big.NewInt(0)
	}
	res := new(big.Int).SetUint64(v[0])
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[1]), 32))  //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[2]), 64))  //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[3]), 96))  //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[4]), 128)) //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[5]), 160)) //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[6]), 192)) //nolint:mnd
	res.Add(res, new(big.Int).Lsh(new(big.Int).SetUint64(v[7]), 224)) //nolint:mnd

	return res
}
