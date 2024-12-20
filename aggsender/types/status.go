package types

import (
	"time"

	zkevm "github.com/0xPolygon/cdk"
)

type AggsenderStatusType string

const (
	StatusNone                 AggsenderStatusType = "none"
	StatusCheckingInitialStage AggsenderStatusType = "checking_initial_stage"
	StatusCertificateStage     AggsenderStatusType = "certificate_stage"
)

type AggsenderStatus struct {
	Running   bool                `json:"running"`
	StartTime time.Time           `json:"start_time"`
	Status    AggsenderStatusType `json:"status"`
	LastError string              `json:"last_error"`
}

type AggsenderInfo struct {
	AggsenderStatus          AggsenderStatus `json:"aggsender_status"`
	Version                  zkevm.FullVersion
	EpochNotifierDescription string `json:"epoch_notifier_description"`
	NetworkID                uint32 `json:"network_id"`
}

func (a *AggsenderStatus) Start(startTime time.Time) {
	a.Running = true
	a.StartTime = startTime
}

func (a *AggsenderStatus) SetLastError(err error) {
	if err == nil {
		a.LastError = ""
	} else {
		a.LastError = err.Error()
	}
}
