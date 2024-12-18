package aggsenderrpc

import (
	"fmt"

	zkevm "github.com/0xPolygon/cdk"
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

type aggsenderStatuser interface {
	Status() types.AggsenderStatus
}

type AggsenderRPC struct {
	logger    *log.Logger
	storage   aggsenderStorer
	aggsender aggsenderStatuser
}

func NewAggsenderRPC(
	logger *log.Logger,
	storage aggsenderStorer,
	aggsender aggsenderStatuser,
) *AggsenderRPC {
	return &AggsenderRPC{
		logger:    logger,
		storage:   storage,
		aggsender: aggsender,
	}
}

// curl -X POST http://localhost:5576/ -H "Con -application/json" \
// -d '{"method":"aggsender_status", "params":[], "id":1}'
func (b *AggsenderRPC) Status() (interface{}, rpc.Error) {
	status := b.aggsender.Status()
	res := struct {
		Version       zkevm.FullVersion     `json:"version"`
		Status        types.AggsenderStatus `json:"status"`
		EpochNotifier string                `json:"epoch_notifier"`
	}{
		Version: zkevm.GetVersion(),
		Status:  status,
	}
	return res, nil
}

// latest: curl -X POST http://localhost:5576/ -H "Con -application/json" \
// -d '{"method":"aggsender_getCertificateHeaderPerHeight", "params":[], "id":1}'
func (b *AggsenderRPC) GetCertificateHeaderPerHeight(height *uint64) (interface{}, rpc.Error) {
	var certInfo *types.CertificateInfo
	var err error
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
