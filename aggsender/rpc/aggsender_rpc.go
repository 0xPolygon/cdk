package aggsenderrpc

import (
	"fmt"

	"github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/aggsender/types"
	"github.com/0xPolygon/cdk/log"
)

const (
	base10 = 10
)

type aggsenderStorer interface {
	GetCertificateByHeight(height uint64) (*types.CertificateInfo, error)
	GetLastSentCertificate() (*types.CertificateInfo, error)
}

type aggsenderInterface interface {
	Info() types.AggsenderInfo
}

// AggsenderRPC is the RPC interface for the aggsender
type AggsenderRPC struct {
	logger    *log.Logger
	storage   aggsenderStorer
	aggsender aggsenderInterface
}

func NewAggsenderRPC(
	logger *log.Logger,
	storage aggsenderStorer,
	aggsender aggsenderInterface,
) *AggsenderRPC {
	return &AggsenderRPC{
		logger:    logger,
		storage:   storage,
		aggsender: aggsender,
	}
}

// Status returns the status of the aggsender
// curl -X POST http://localhost:5576/ -H "Con -application/json" \
// -d '{"method":"aggsender_status", "params":[], "id":1}'
func (b *AggsenderRPC) Status() (interface{}, rpc.Error) {
	info := b.aggsender.Info()
	return info, nil
}

// GetCertificateHeaderPerHeight returns the certificate header for the given height
// if param is `nil` it returns the last sent certificate
// latest:
//
//	curl -X POST http://localhost:5576/ -H "Con -application/json" \
//	 -d '{"method":"aggsender_getCertificateHeaderPerHeight", "params":[], "id":1}'
//
// specific height:
//
// curl -X POST http://localhost:5576/ -H "Con -application/json" \
// -d '{"method":"aggsender_getCertificateHeaderPerHeight", "params":[$height], "id":1}'
func (b *AggsenderRPC) GetCertificateHeaderPerHeight(height *uint64) (interface{}, rpc.Error) {
	var (
		certInfo *types.CertificateInfo
		err      error
	)
	if height == nil {
		certInfo, err = b.storage.GetLastSentCertificate()
	} else {
		certInfo, err = b.storage.GetCertificateByHeight(*height)
	}
	if err != nil {
		return nil, rpc.NewRPCError(rpc.DefaultErrorCode, fmt.Sprintf("error getting certificate by height: %v", err))
	}
	if certInfo == nil {
		return nil, rpc.NewRPCError(rpc.NotFoundErrorCode, "certificate not found")
	}
	return certInfo, nil
}
