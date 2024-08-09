package config

// DefaultValues is the default configuration
const DefaultValues = `
ForkUpgradeBatchNumber = 0
ForkUpgradeNewForkId = 0

[Common]
IsValidiumMode = false
ContractVersions = "banana"

[Log]
Environment = "development" # "production" or "development"
Level = "info"
Outputs = ["stderr"]

[SequenceSender]
WaitPeriodSendSequence = "15s"
LastBatchVirtualizationTimeMaxWaitPeriod = "10s"
L1BlockTimestampMargin = "30s"
MaxTxSizeForL1 = 131072
L2Coinbase = "0xfa3b44587990f97ba8b6ba7e230a5f0e95d14b3d"
PrivateKey = {Path = "./test/sequencer.keystore", Password = "testonly"}
SequencesTxFileName = "sequencesender.json"
GasOffset = 80000
WaitPeriodPurgeTxFile = "15m"
MaxPendingTx = 1
	[SequenceSender.StreamClient]
		Server = "127.0.0.1:6900"
	[SequenceSender.EthTxManager]
		FrequencyToMonitorTxs = "1s"
		WaitTxToBeMined = "2m"
		GetReceiptMaxTime = "250ms"
		GetReceiptWaitInterval = "1s"
		PrivateKeys = [
			{Path = "./test/sequencer.keystore", Password = "testonly"},
		]
		ForcedGas = 0
		GasPriceMarginFactor = 1
		MaxGasPriceLimit = 0
		PersistenceFilename = "ethtxmanager.json"
		ReadPendingL1Txs = false
		SafeStatusL1NumberOfBlocks = 0
		FinalizedStatusL1NumberOfBlocks = 0
			[SequenceSender.EthTxManager.Etherman]
				URL = "http://127.0.0.1:8545"
				MultiGasProvider = false
				L1ChainID = 1337
[Aggregator]
Host = "0.0.0.0"
Port = 50081
RetryTime = "5s"
VerifyProofInterval = "10s"
TxProfitabilityCheckerType = "acceptall"
TxProfitabilityMinReward = "1.1"
ProofStatePollingInterval = "5s"
SenderAddress = ""
CleanupLockedProofsInterval = "2m"
GeneratingProofCleanupThreshold = "10m"
ForkId = 9
GasOffset = 0
WitnessURL = "localhost:8123"
UseL1BatchData = true
UseFullWitness = false
SettlementBackend = "l1"
AggLayerTxTimeout = "5m"
AggLayerURL = ""
SequencerPrivateKey = {}
	[Aggregator.DB]
		Name = "aggregator_db"
		User = "aggregator_user"
		Password = "aggregator_password"
		Host = "cdk-aggregator-db"
		Port = "5432"
		EnableLog = false	
		MaxConns = 200
	[Aggregator.Log]
		Environment = "development" # "production" or "development"
		Level = "info"
		Outputs = ["stderr"]
	[Aggregator.StreamClient]
		Server = "localhost:6900"
	[Aggregator.EthTxManager]
		FrequencyToMonitorTxs = "1s"
		WaitTxToBeMined = "2m"
		GetReceiptMaxTime = "250ms"
		GetReceiptWaitInterval = "1s"
		PrivateKeys = [
			{Path = "/pk/aggregator.keystore", Password = "testonly"},
		]
		ForcedGas = 0
		GasPriceMarginFactor = 1
		MaxGasPriceLimit = 0
		PersistenceFilename = ""
		ReadPendingL1Txs = false
		SafeStatusL1NumberOfBlocks = 0
		FinalizedStatusL1NumberOfBlocks = 0
			[Aggregator.EthTxManager.Etherman]
				URL = ""
				L1ChainID = 11155111
				HTTPHeaders = []
	[Aggregator.Synchronizer]
		[Aggregator.Synchronizer.DB]
			Name = "sync_db"
			User = "sync_user"
			Password = "sync_password"
			Host = "cdk-l1-sync-db"
			Port = "5432"
			EnableLog = false
			MaxConns = 10
		[Aggregator.Synchronizer.Synchronizer]
			SyncInterval = "10s"
			SyncChunkSize = 1000
			GenesisBlockNumber = 5511080
			SyncUpToBlock = "finalized"
			BlockFinality = "finalized"
			OverrideStorageCheck = false
		[Aggregator.Synchronizer.Etherman]
			[Aggregator.Synchronizer.Etherman.Validium]
				Enabled = false

[ReorgDetectorL1]
DBPath = "/tmp/reorgdetector"

[L1InfoTreeSync]
DBPath = "/tmp/L1InfoTreeSync"
GlobalExitRootAddr="0x8464135c8F25Da09e49BC8782676a84730C318bC"
SyncBlockChunkSize=10
BlockFinality="LatestBlock"
URLRPCL1="http://test-aggoracle-l1:8545"
WaitForNewBlocksPeriod="100ms"
InitialBlock=0

[AggOracle]
TargetChainType="EVM"
URLRPCL1="http://test-aggoracle-l1:8545"
BlockFinality="LatestBlock"
WaitPeriodNextGER="100ms"
	[AggOracle.EVMSender]
		GlobalExitRootL2="0x8464135c8F25Da09e49BC8782676a84730C318bC"
		URLRPCL2="http://test-aggoracle-l2:8545"
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
				PersistenceFilename = "/tmp/ethtxmanager.json"
				ReadPendingL1Txs = false
				SafeStatusL1NumberOfBlocks = 5
				FinalizedStatusL1NumberOfBlocks = 10
					[AggOracle.EVMSender.EthTxManager.Etherman]
						URL = "http://test-aggoracle-l2"
						MultiGasProvider = false
						L1ChainID = 1337
						HTTPHeaders = []
`
