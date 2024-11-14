package config

// This values doesnt have a default value because depend on the
// environment / deployment
const DefaultMandatoryVars = `
L1URL = "http://localhost:8545"
L2URL = "http://localhost:8123"
AggLayerURL = "https://agglayer-dev.polygon.technology"


ForkId = 9
ContractVersions = "elderberry"
IsValidiumMode = false

L2Coinbase = "0xfa3b44587990f97ba8b6ba7e230a5f0e95d14b3d"
SequencerPrivateKeyPath = "/app/sequencer.keystore"
SequencerPrivateKeyPassword = "test"
WitnessURL = "http://localhost:8123"

AggregatorPrivateKeyPath = "/app/keystore/aggregator.keystore"
AggregatorPrivateKeyPassword = "testonly"
# Who send Proof to L1? AggLayer addr, or aggregator addr?
SenderProofToL1Addr = "0x0000000000000000000000000000000000000000"
polygonBridgeAddr = "0x0000000000000000000000000000000000000000"


# This values can be override directly from genesis.json
rollupCreationBlockNumber = 0
rollupManagerCreationBlockNumber = 0
genesisBlockNumber = 0
[L1Config]
	chainId = 0
	polygonZkEVMGlobalExitRootAddress = "0x0000000000000000000000000000000000000000"
	polygonRollupManagerAddress = "0x0000000000000000000000000000000000000000"
	polTokenAddress = "0x0000000000000000000000000000000000000000"
	polygonZkEVMAddress = "0x0000000000000000000000000000000000000000"


[L2Config]
	GlobalExitRootAddr = "0x0000000000000000000000000000000000000000"

`

// This doesnt below to config, but are the vars used
// to avoid repetition in config-files
const DefaultVars = `
PathRWData = "/tmp/cdk"
L1URLSyncChunkSize = 100

`

// DefaultValues is the default configuration
const DefaultValues = `
ForkUpgradeBatchNumber = 0
ForkUpgradeNewForkId = 0


[Log]
Environment = "development" # "production" or "development"
Level = "info"
Outputs = ["stderr"]

[Etherman]
	URL="{{L1URL}}"
	ForkIDChunkSize={{L1URLSyncChunkSize}}
	[Etherman.EthermanConfig]
		URL="{{L1URL}}"
		MultiGasProvider=false
		L1ChainID={{NetworkConfig.L1.L1ChainID}}
		HTTPHeaders=[]
		[Etherman.EthermanConfig.Etherscan]
			ApiKey=""
			Url="https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey="

[Common]
NetworkID = 1
IsValidiumMode = {{IsValidiumMode}}
ContractVersions = "{{ContractVersions}}"

[SequenceSender]
WaitPeriodSendSequence = "15s"
LastBatchVirtualizationTimeMaxWaitPeriod = "10s"
L1BlockTimestampMargin = "30s"
MaxTxSizeForL1 = 131072
L2Coinbase = "{{L2Coinbase}}"
PrivateKey = { Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"}
SequencesTxFileName = "sequencesender.json"
GasOffset = 80000
WaitPeriodPurgeTxFile = "15m"
MaxPendingTx = 1
MaxBatchesForL1 = 300
BlockFinality = "FinalizedBlock"
RPCURL = "{{L2URL}}"
GetBatchWaitInterval = "10s"
	[SequenceSender.EthTxManager]
		FrequencyToMonitorTxs = "1s"
		WaitTxToBeMined = "2m"
		GetReceiptMaxTime = "250ms"
		GetReceiptWaitInterval = "1s"
		PrivateKeys = [
			{Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"},
		]
		ForcedGas = 0
		GasPriceMarginFactor = 1
		MaxGasPriceLimit = 0
		StoragePath = "ethtxmanager.sqlite"
		ReadPendingL1Txs = false
		SafeStatusL1NumberOfBlocks = 0
		FinalizedStatusL1NumberOfBlocks = 0
			[SequenceSender.EthTxManager.Etherman]
				URL = "{{L1URL}}"
				MultiGasProvider = false
				L1ChainID = {{NetworkConfig.L1.L1ChainID}}
[Aggregator]
# GRPC server host
Host = "0.0.0.0"
# GRPC server port
Port = 50081
RetryTime = "5s"
VerifyProofInterval = "10s"
ProofStatePollingInterval = "5s"
TxProfitabilityCheckerType = "acceptall"
TxProfitabilityMinReward = "1.1"
IntervalAfterWhichBatchConsolidateAnyway="0s"
BatchProofSanityCheckEnabled = true
#  ChainID is L2ChainID. Is populated on runtimme
ChainID = 0
ForkId = {{ForkId}}
SenderAddress = "{{SenderProofToL1Addr}}"
CleanupLockedProofsInterval = "2m"
GeneratingProofCleanupThreshold = "10m"
GasOffset = 0
RPCURL = "{{L2URL}}"
WitnessURL = "{{WitnessURL}}"
UseFullWitness = false
SettlementBackend = "l1"
AggLayerTxTimeout = "5m"
AggLayerURL = "{{AggLayerURL}}"
SyncModeOnlyEnabled = false
	[Aggregator.SequencerPrivateKey]
		Path = "{{SequencerPrivateKeyPath}}"
		Password = "{{SequencerPrivateKeyPassword}}"
	[Aggregator.DB]
		Name = "aggregator_db"
		User = "aggregator_user"
		Password = "aggregator_password"
		Host = "cdk-aggregator-db"
		Port = "5432"
		EnableLog = false	
		MaxConns = 200
	[Aggregator.Log]
		Environment ="{{Log.Environment}}" # "production" or "development"
		Level = "{{Log.Level}}"
		Outputs = ["stderr"]
	[Aggregator.EthTxManager]
		FrequencyToMonitorTxs = "1s"
		WaitTxToBeMined = "2m"
		GetReceiptMaxTime = "250ms"
		GetReceiptWaitInterval = "1s"
		PrivateKeys = [
			{Path = "{{AggregatorPrivateKeyPath}}", Password = "{{AggregatorPrivateKeyPassword}}"},
		]
		ForcedGas = 0
		GasPriceMarginFactor = 1
		MaxGasPriceLimit = 0
		StoragePath = ""
		ReadPendingL1Txs = false
		SafeStatusL1NumberOfBlocks = 0
		FinalizedStatusL1NumberOfBlocks = 0
			[Aggregator.EthTxManager.Etherman]
				URL = "{{L1URL}}"
				L1ChainID = {{NetworkConfig.L1.L1ChainID}}
				HTTPHeaders = []
	[Aggregator.Synchronizer]
		[Aggregator.Synchronizer.Log]
			Environment = "{{Log.Environment}}" # "production" or "development"
			Level = "{{Log.Level}}"
			Outputs = ["stderr"]
		[Aggregator.Synchronizer.SQLDB]
			DriverName = "sqlite3"
			DataSource = "file:{{PathRWData}}/aggregator_sync_db.sqlite"
		[Aggregator.Synchronizer.Synchronizer]
			SyncInterval = "10s"
			SyncChunkSize = 1000
			GenesisBlockNumber = {{genesisBlockNumber}}
			SyncUpToBlock = "finalized"
			BlockFinality = "finalized"
			OverrideStorageCheck = false
		[Aggregator.Synchronizer.Etherman]
			L1URL = "{{L1URL}}"
			ForkIDChunkSize = 100
			L1ChainID = {{NetworkConfig.L1.L1ChainID}}
			PararellBlockRequest = false
			[Aggregator.Synchronizer.Etherman.Contracts]
				GlobalExitRootManagerAddr = "{{NetworkConfig.L1.GlobalExitRootManagerAddr}}"
				RollupManagerAddr = "{{NetworkConfig.L1.RollupManagerAddr}}"
				ZkEVMAddr = "{{NetworkConfig.L1.ZkEVMAddr}}"
			[Aggregator.Synchronizer.Etherman.Validium]
				Enabled = {{IsValidiumMode}}
				# L2URL, empty ask to contract
				TrustedSequencerURL = ""
				RetryOnDACErrorInterval = "1m"
				DataSourcePriority = ["trusted", "external"]
			[Aggregator.Synchronizer.Etherman.Validium.Translator]
				FullMatchRules = []
			[Aggregator.Synchronizer.Etherman.Validium.RateLimit]
				NumRequests = 1000
				Interval = "1s"
[ReorgDetectorL1]
DBPath = "{{PathRWData}}/reorgdetectorl1.sqlite"

[ReorgDetectorL2]
DBPath = "{{PathRWData}}/reorgdetectorl2.sqlite"

[L1InfoTreeSync]
DBPath = "{{PathRWData}}/L1InfoTreeSync.sqlite"
GlobalExitRootAddr="{{NetworkConfig.L1.GlobalExitRootManagerAddr}}"
RollupManagerAddr = "{{NetworkConfig.L1.RollupManagerAddr}}"
SyncBlockChunkSize=100
BlockFinality="LatestBlock"
URLRPCL1="{{L1URL}}"
WaitForNewBlocksPeriod="100ms"
InitialBlock=0

[AggOracle]
TargetChainType="EVM"
URLRPCL1="{{L1URL}}"
BlockFinality="FinalizedBlock"
WaitPeriodNextGER="100ms"
	[AggOracle.EVMSender]
		GlobalExitRootL2="{{L2Config.GlobalExitRootAddr}}"
		URLRPCL2="{{L2URL}}"
		ChainIDL2=1337
		GasOffset=0
		WaitPeriodMonitorTx="100ms"
		SenderAddr="0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
		[AggOracle.EVMSender.EthTxManager]
				FrequencyToMonitorTxs = "1s"
				WaitTxToBeMined = "2s"
				GetReceiptMaxTime = "250ms"
				GetReceiptWaitInterval = "1s"
				PrivateKeys = [
					{Path = "/app/keystore/aggoracle.keystore", Password = "testonly"},
				]
				ForcedGas = 0
				GasPriceMarginFactor = 1
				MaxGasPriceLimit = 0
				StoragePath = "/{{PathRWData}}/ethtxmanager-sequencesender.sqlite"
				ReadPendingL1Txs = false
				SafeStatusL1NumberOfBlocks = 5
				FinalizedStatusL1NumberOfBlocks = 10
					[AggOracle.EVMSender.EthTxManager.Etherman]
						URL = "{{L2URL}}"
						MultiGasProvider = false
						L1ChainID = {{NetworkConfig.L1.L1ChainID}}
						HTTPHeaders = []

[RPC]
Host = "0.0.0.0"
Port = 5576
ReadTimeout = "2s"
WriteTimeout = "2s"
MaxRequestsPerIPAndSecond = 10

[ClaimSponsor]
DBPath = "/{{PathRWData}}/claimsopnsor.sqlite"
Enabled = true
SenderAddr = "0xfa3b44587990f97ba8b6ba7e230a5f0e95d14b3d"
BridgeAddrL2 = "0xB7098a13a48EcE087d3DA15b2D28eCE0f89819B8"
MaxGas = 200000
RetryAfterErrorPeriod = "1s"
MaxRetryAttemptsAfterError = -1
WaitTxToBeMinedPeriod = "3s"
WaitOnEmptyQueue = "3s"
GasOffset = 0
	[ClaimSponsor.EthTxManager]
		FrequencyToMonitorTxs = "1s"
		WaitTxToBeMined = "2s"
		GetReceiptMaxTime = "250ms"
		GetReceiptWaitInterval = "1s"
		PrivateKeys = [
			{Path = "/app/keystore/claimsopnsor.keystore", Password = "testonly"},
		]
		ForcedGas = 0
		GasPriceMarginFactor = 1
		MaxGasPriceLimit = 0
		StoragePath = "/{{PathRWData}}/ethtxmanager-claimsponsor.sqlite"
		ReadPendingL1Txs = false
		SafeStatusL1NumberOfBlocks = 5
		FinalizedStatusL1NumberOfBlocks = 10
			[ClaimSponsor.EthTxManager.Etherman]
				URL = "{{L2URL}}"
				MultiGasProvider = false
				L1ChainID = {{NetworkConfig.L1.L1ChainID}}
				HTTPHeaders = []

[BridgeL1Sync]
DBPath = "{{PathRWData}}/bridgel1sync.sqlite"
BlockFinality = "LatestBlock"
InitialBlockNum = 0
BridgeAddr = "{{polygonBridgeAddr}}"
SyncBlockChunkSize = 100
RetryAfterErrorPeriod = "1s"
MaxRetryAttemptsAfterError = -1
WaitForNewBlocksPeriod = "3s"
OriginNetwork=0

[BridgeL2Sync]
DBPath = "{{PathRWData}}/bridgel2sync.sqlite"
BlockFinality = "LatestBlock"
InitialBlockNum = 0
BridgeAddr = "{{polygonBridgeAddr}}"
SyncBlockChunkSize = 100
RetryAfterErrorPeriod = "1s"
MaxRetryAttemptsAfterError = -1
WaitForNewBlocksPeriod = "3s"
OriginNetwork=1

[LastGERSync]
DBPath = "{{PathRWData}}/lastgersync.sqlite"
BlockFinality = "LatestBlock"
InitialBlockNum = 0
GlobalExitRootL2Addr = "{{L2Config.GlobalExitRootAddr}}"
RetryAfterErrorPeriod = "1s"
MaxRetryAttemptsAfterError = -1
WaitForNewBlocksPeriod = "1s"
DownloadBufferSize = 100

[NetworkConfig.L1]
L1ChainID = {{L1Config.chainId}}
PolAddr = "{{L1Config.polTokenAddress}}"
ZkEVMAddr = "{{L1Config.polygonZkEVMAddress}}"
RollupManagerAddr = "{{L1Config.polygonRollupManagerAddress}}"
GlobalExitRootManagerAddr = "{{L1Config.polygonZkEVMGlobalExitRootAddress}}"


[AggSender]
StoragePath = "{{PathRWData}}/aggsender.sqlite"
AggLayerURL = "{{AggLayerURL}}"
AggsenderPrivateKey = {Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"}
BlockGetInterval = "2s"
URLRPCL2="{{L2URL}}"
CheckSettledInterval = "2s"
BlockFinality = "LatestBlock"
EpochNotificationPercentage = 50
SaveCertificatesToFilesPath = ""
`
