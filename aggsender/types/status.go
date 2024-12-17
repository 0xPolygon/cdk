package types

import "time"

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

func (a *AggsenderStatus) SetLastError(err error) {
	if err == nil {
		a.LastError = ""
	} else {
		a.LastError = err.Error()
	}
}
