package common

const (
	// SEQUENCE_SENDER name to identify the sequence-sender component
	SEQUENCE_SENDER = "sequence-sender" //nolint:stylecheck
	// AGGREGATOR name to identify the aggregator component
	AGGREGATOR = "aggregator"
	// AGGORACLE name to identify the aggoracle component
	AGGORACLE = "aggoracle"
	// RPC name to identify the rpc component (implides implies bridgesyncL1, bridgesyncL2)
	RPC = "rpc"
	// CLAIM_SPONSOR name to identify the claim sponsor component
	CLAIM_SPONSOR = "claim-sponsor" //nolint:stylecheck
	// PROVER name to identify the prover component
	PROVER = "prover"
	// AGGSENDER name to identify the aggsender component (implies bridgesyncL2)
	AGGSENDER = "aggsender"
	// BRIDGE_SYNC_L1 name to identify the bridgesyncL1 component
	BRIDGE_SYNC_L1 = "bridgesyncL1" //nolint:stylecheck
	// BRIDGE_SYNC_L2 name to identify the bridgesyncL2 component
	BRIDGE_SYNC_L2 = "bridgesyncL2" //nolint:stylecheck
)
