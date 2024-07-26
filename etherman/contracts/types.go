package contracts

import (
	"errors"
)

type NameType string

const (
	ContractNameDAProtocol     NameType = "daprotocol"
	ContractNameRollupManager  NameType = "rollupmanager"
	ContractNameGlobalExitRoot NameType = "globalexitroot"
	ContractNameRollup         NameType = "rollup"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type VersionType string

const (
	VersionBanana     VersionType = "banana"
	VersionElderberry VersionType = "elderberry"
)
