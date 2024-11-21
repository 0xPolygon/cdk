package config

// This values doesnt have a default value because depend on the
// environment / deployment
const DefaultMandatoryVars = `
# Layer 1 (Ethereum) RPC provider URL
L1URL = "http://localhost:8545"

# Layer 2 Sequencer RPC URL
# Usually running on the same machine as cdk-node
L2URL = "http://localhost:8123"

# AggLayer URL
AggLayerURL = "https://agglayer-dev.polygon.technology"

# ForkId is the fork id of the chain
ForkId = 12

# ContractVersions is the version of the contracts used in the chain
ContractVersions = "banana"
# Set if it's a Validium or a Rollup chain
IsValidiumMode = false

# L2Coinbase is the address that will receive the fees
L2Coinbase = "0xfa3b44587990f97ba8b6ba7e230a5f0e95d14b3d"
# SequencerPrivateKeyPath is the path to the sequencer private key
SequencerPrivateKeyPath = "/app/sequencer.keystore"
# SequencerPrivateKeyPassword is the password to the sequencer private key
SequencerPrivateKeyPassword = "test"
# WitnessURL is the URL of the RPC node that will be used to get the witness
WitnessURL = "http://localhost:8123"

# AggregatorPrivateKeyPath is the path to the aggregator private key
AggregatorPrivateKeyPath = "/app/keystore/aggregator.keystore"
# AggregatorPrivateKeyPassword is the password to the aggregator private key
AggregatorPrivateKeyPassword = "testonly"
# SenderProofToL1Addr is the address that sends the proof to L1
SenderProofToL1Addr = "0x0000000000000000000000000000000000000000"
# polygonBridgeAddr is the address of the bridge contract
polygonBridgeAddr = "0x0000000000000000000000000000000000000000"

##################################################
# This values can be override directly from genesis.json
# rollupCreationBlockNumber is the block number where the rollup contract was created
rollupCreationBlockNumber = 0
# rollupManagerCreationBlockNumber is the block number where the rollup manager contract was created
rollupManagerCreationBlockNumber = 0
# genesisBlockNumber is the block number of the genesis block
genesisBlockNumber = 0

##################################################
# This values are used to configure the L1 network
[L1Config]
  # chainId is the chain id of the L1 network
  chainId = 0
  # polygonZkEVMGlobalExitRootAddress is the address of the global exit root contract
  polygonZkEVMGlobalExitRootAddress = "0x0000000000000000000000000000000000000000"
  # polygonRollupManagerAddress is the address of the rollup manager contract
  polygonRollupManagerAddress = "0x0000000000000000000000000000000000000000"
  # polTokenAddress is the address of the POL token
  polTokenAddress = "0x0000000000000000000000000000000000000000"
  # polygonZkEVMAddress is the address of the ZkEVM contract
  polygonZkEVMAddress = "0x0000000000000000000000000000000000000000"

##################################################
# This values are used to configure the L2 network
[L2Config]
  # GlobalExitRootAddr is the address of the global exit root contract
  GlobalExitRootAddr = "0x0000000000000000000000000000000000000000"

`

// This doesn't belong to config, but are the vars used
// to avoid repetition in config-files
const DefaultVars = `
PathRWData = "/tmp/cdk"
L1URLSyncChunkSize = 100

`

// DefaultValues is the default configuration
const DefaultValues = `
# This is the default configuration for the cdk-node

# ForkUpgradeBatchNumber is the batch number of the fork upgrade
# when the batch number is reached the node will stop to allow the fork upgrade
ForkUpgradeBatchNumber = 0
# ForkUpgradeNewForkId is the new fork id that will be used after the fork upgrade
ForkUpgradeNewForkId = 0

# Log configuration
[Log]
  # Environment is the environment where the node is running
  Environment = "development" # "production" or "development"
  # Level is the log level
  Level = "info"
  # Outputs are the outputs where the logs will be written
  Outputs = ["stderr"]

# Etherman configuration
[Etherman]
  # URL is the URL of the L1 RPC node, usually an Ethereum node or and RPC provider
  URL="{{L1URL}}"
  # ForkIDChunkSize is the chunk size used to get the fork id.
  # This depends on the RPC provider, some providers don't support big chunk sizes
  # so this should be adjusted to what the provider supports
  ForkIDChunkSize={{L1URLSyncChunkSize}}
  
  # EthermanConfig is the configuration of the Etherman
  [Etherman.EthermanConfig]
    # URL is the URL of the L1 RPC node, usually an Ethereum node or and RPC provider
    URL="{{L1URL}}"
    # MultiGasProvider is used to get the gas price from multiple providers
    MultiGasProvider=false
    # L1ChainID is the chain id of the L1 network
    L1ChainID={{NetworkConfig.L1.L1ChainID}}
    # HTTPHeaders additional headers that will be used in the HTTP requests
    HTTPHeaders=[]
    [Etherman.EthermanConfig.Etherscan]
	  ApiKey=""
	  Url="https://api.etherscan.io/api?module=gastracker&action=gasoracle&apikey="

# Common configuration
[Common]
  # NetworkID is the network id of the CDK chain
  NetworkID = 1
  # IsValidiumMode is true if the chain is a Validium chain false if it's a Rollup chain
  IsValidiumMode = {{IsValidiumMode}}
  # Versions of the contracts used in the chain
  ContractVersions = "{{ContractVersions}}"

# SequenceSender configuration
[SequenceSender]
  # WaitPeriodSendSequence is the wait period between sending sequences
  WaitPeriodSendSequence = "15s"
  # LastBatchVirtualizationTimeMaxWaitPeriod is the max wait period to virtualize the last batch
  LastBatchVirtualizationTimeMaxWaitPeriod = "10s"
  # L1BlockTimestampMargin is the margin of the L1 block timestamp
  L1BlockTimestampMargin = "30s"
  # MaxTxSizeForL1 is the max size of the transaction that will be sent to L1
  MaxTxSizeForL1 = 131072
  # L2Coinbase is the address that will receive the fees
  L2Coinbase = "{{L2Coinbase}}"
  # PrivateKey is the private key of the squence-sender
  PrivateKey = { Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"}
  # SequencesTxFileName is the file name to store sequences sent to L1
  SequencesTxFileName = "sequencesender.json"
  # GasOffset is the amount of gas to be added to the gas estimation in order
  # to provide an amount that is higher than the estimated one. This is used
  # to avoid the TX getting reverted in case something has changed in the network
  # state after the estimation which can cause the TX to require more gas to be
  # executed.
  #
  #  ex:
  #  gas estimation: 1000
  #  gas offset: 100
  #  final gas: 1100
  GasOffset = 80000
  # WaitPeriodPurgeTxFile is the wait period to purge the tx file
  WaitPeriodPurgeTxFile = "15m"
  # MaxPendingTx is the max number of pending transactions
  MaxPendingTx = 1
  # MaxBatchesForL1 is the maximum amount of batches to be sequenced in a single L1 tx
  MaxBatchesForL1 = 300
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality = "FinalizedBlock"
  # RPCURL is the URL of the L2 RPC node
  RPCURL = "{{L2URL}}"
  # GetBatchWaitInterval is the time to wait to query for a new batch when there are no more batches available
  GetBatchWaitInterval = "10s"
  [SequenceSender.EthTxManager]
    # FrequencyToMonitorTxs frequency of the resending failed txs
    FrequencyToMonitorTxs = "1s"
    # WaitTxToBeMined time to wait after transaction was sent to the ethereum
    WaitTxToBeMined = "2m"
    # GetReceiptMaxTime is the max time to wait to get the receipt of the mined transaction
    GetReceiptMaxTime = "250ms"
    # GetReceiptWaitInterval is the time to sleep before trying to get the receipt of the mined transaction
    GetReceiptWaitInterval = "1s"
    # PrivateKeys defines all the key store files that are going
    # to be read in order to provide the private keys to sign the L1 txs
    PrivateKeys = [
	    {Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"},
    ]
    # ForcedGas is the amount of gas to be forced in case of gas estimation error
    ForcedGas = 0
    # GasPriceMarginFactor is used to multiply the suggested gas price provided by the network
    # in order to allow a different gas price to be set for all the transactions and making it
    # easier to have the txs prioritized in the pool, default value is 1.
    #
    # ex:
    # suggested gas price: 100
    # GasPriceMarginFactor: 1
    # gas price = 100
    #
    # suggested gas price: 100
    # GasPriceMarginFactor: 1.1
    # gas price = 110
    GasPriceMarginFactor = 1
    # MaxGasPriceLimit helps avoiding transactions to be sent over an specified
    # gas price amount, default value is 0, which means no limit.
    # If the gas price provided by the network and adjusted by the GasPriceMarginFactor
    # is greater than this configuration, transaction will have its gas price set to
    # the value configured in this config as the limit.
    #
    # ex:
    #
    # suggested gas price: 100
    # gas price margin factor: 20%
    # max gas price limit: 150
    # tx gas price = 120
    #
    # suggested gas price: 100
    # gas price margin factor: 20%
    # max gas price limit: 110
    # tx gas price = 110
    MaxGasPriceLimit = 0
    # StoragePath is the path of the internal storage
    StoragePath = "ethtxmanager.sqlite"
    # ReadPendingL1Txs is a flag to enable the reading of pending L1 txs
    # It can only be enabled if DBPath is empty
    ReadPendingL1Txs = false
    # SafeStatusL1NumberOfBlocks overwrites the number of blocks to consider a tx as safe
    # overwriting the default value provided by the network
    # 0 means that the default value will be used
    SafeStatusL1NumberOfBlocks = 0
    # FinalizedStatusL1NumberOfBlocks overwrites the number of blocks to consider a tx as finalized
    # overwriting the default value provided by the network
    # 0 means that the default value will be used
    FinalizedStatusL1NumberOfBlocks = 0
    [SequenceSender.EthTxManager.Etherman]
      # URL is the URL of the Ethereum node for L1
      URL = "{{L1URL}}"
      # allow that L1 gas price calculation use multiples sources
      MultiGasProvider = false
      # L1ChainID is the chain ID of the L1
      L1ChainID = {{NetworkConfig.L1.L1ChainID}}
[Aggregator]
  # GRPC server host
  Host = "0.0.0.0"
  # GRPC server port
  Port = 50081
  # RetryTime is the time the aggregator main loop sleeps if there are no proofs to aggregate
	# or batches to generate proofs. It is also used in the isSynced loop
  RetryTime = "5s"
  # VerifyProofInterval is the interval of time to verify/send an proof in L1
  VerifyProofInterval = "10s"
  # ProofStatePollingInterval is the interval time to polling the prover about the generation state of a proof
  ProofStatePollingInterval = "5s"
  # TxProfitabilityCheckerType type for checking is it profitable for aggregator to validate batch
  # possible values: base/acceptall
  TxProfitabilityCheckerType = "acceptall"
  # TxProfitabilityMinReward min reward for base tx profitability checker when aggregator will validate batch
  # this parameter is used for the base tx profitability checker
  TxProfitabilityMinReward = "1.1"
  # IntervalAfterWhichBatchConsolidateAnyway is the interval duration for the main sequencer to check
  # if there are no transactions. If there are no transactions in this interval, the sequencer will
  # consolidate the batch anyway.
  IntervalAfterWhichBatchConsolidateAnyway="0s"
  # BatchProofSanityCheckEnabled is a flag to enable the sanity check of the batch proof
  BatchProofSanityCheckEnabled = true
  # ChainID is the L2 ChainID provided by the Network Config
  ChainID = 0
  # ForkID is the L2 ForkID provided by the Network Config
  ForkId = {{ForkId}}
  # SenderAddress defines which private key the eth tx manager needs to use
	# to sign the L1 txs
  SenderAddress = "{{SenderProofToL1Addr}}"
  # CleanupLockedProofsInterval is the interval of time to clean up locked proofs.
  CleanupLockedProofsInterval = "2m"
  # GeneratingProofCleanupThreshold represents the time interval after
	# which a proof in generating state is considered to be stuck and
	# allowed to be cleared.
  GeneratingProofCleanupThreshold = "10m"
  # GasOffset is the amount of gas to be added to the gas estimation in order
	# to provide an amount that is higher than the estimated one. This is used
	# to avoid the TX getting reverted in case something has changed in the network
	# state after the estimation which can cause the TX to require more gas to be
	# executed.
	#
	# ex:
	# gas estimation: 1000
	# gas offset: 100
	# final gas: 1100
  GasOffset = 0
  # RPCURL is the URL of the RPC server
  RPCURL = "{{L2URL}}"
  # WitnessURL is the URL of the witness server
  WitnessURL = "{{WitnessURL}}"
  # UseFullWitness is a flag to enable the use of full witness in the aggregator
  UseFullWitness = false
  # SettlementBackend configuration defines how a final ZKP should be settled.
	# It can be settled directly to L1 or over Agglayer.
  SettlementBackend = "l1"
  # AggLayerTxTimeout is the interval time to wait for a tx to be mined from the agglayer
  AggLayerTxTimeout = "5m"
  # AggLayerURL url of the agglayer service
  AggLayerURL = "{{AggLayerURL}}"
  # SyncModeOnlyEnabled is a flag that activates sync mode exclusively.
  # When enabled, the aggregator will sync data only from L1 and will not generate or read the data stream.
  SyncModeOnlyEnabled = false
  # SequencerPrivateKey Private key of the trusted sequencer
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
  # DBPath is the path of the database
  DBPath = "{{PathRWData}}/reorgdetectorl1.sqlite"

[ReorgDetectorL2]
  # DBPath is the path of the database
  DBPath = "{{PathRWData}}/reorgdetectorl2.sqlite"

[L1InfoTreeSync]
  # DBPath is the path of the database
  DBPath = "{{PathRWData}}/L1InfoTreeSync.sqlite"
  # GlobalExitRootAddr is the address of the global exit root manager contract
  GlobalExitRootAddr="{{NetworkConfig.L1.GlobalExitRootManagerAddr}}"
  # RollupManagerAddr is the address of the rollup manager contract
  RollupManagerAddr = "{{NetworkConfig.L1.RollupManagerAddr}}"
  # SyncBlockChunkSize is the chunk size used to sync the blocks
  SyncBlockChunkSize=100
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality="LatestBlock"
  # URLRPCL1 is the URL of the L1 RPC node
  URLRPCL1="{{L1URL}}"
  # WaitForNewBlocksPeriod is the wait period to get new blocks
  WaitForNewBlocksPeriod="100ms"
  # InitialBlock is the initial block to start the sync
  InitialBlock={{rollupManagerCreationBlockNumber}}

[AggOracle]
  # TargetChainType is the target chain type
  TargetChainType="EVM"
  # URLRPCL1 is the URL of the L1 RPC node
  URLRPCL1="{{L1URL}}"
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality="FinalizedBlock"
  # WaitPeriodNextGER is the wait period to get the next GER
  WaitPeriodNextGER="100ms"
  [AggOracle.EVMSender]
    # GlobalExitRootL2 is the address of the global exit root contract
    GlobalExitRootL2="{{L2Config.GlobalExitRootAddr}}"
    # URLRPCL2 is the URL of the L2 RPC node
    URLRPCL2="{{L2URL}}"
    # ChainIDL2 is the chain id of the L2 network
    ChainIDL2=1337
    # GasOffset is the gas offset to use when sending the txs
    GasOffset=0
    # WaitPeriodMonitorTx is the wait period to monitor the txs
    WaitPeriodMonitorTx="100ms"
    # SenderAddr is the address of the sender
    SenderAddr="0x70997970c51812dc3a010c7d01b50e0d17dc79c8"
    # EthTxManager is the configuration of the EthTxManager
    [AggOracle.EVMSender.EthTxManager]
      # FrequencyToMonitorTxs is the frequency to monitor the txs
      FrequencyToMonitorTxs = "1s"
      # WaitTxToBeMined is the wait period to wait for the tx to be mined
      WaitTxToBeMined = "2s"
      # GetReceiptMaxTime is the max time to get the receipt of the tx
      GetReceiptMaxTime = "250ms"
      # GetReceiptWaitInterval is the wait interval to get the receipt
      GetReceiptWaitInterval = "1s"
      # PrivateKeys is the private keys to use to sign the txs
      PrivateKeys = [
	      {Path = "/app/keystore/aggoracle.keystore", Password = "testonly"},
      ]
      # ForcedGas is the forced gas to use in case of gas estimation error
      ForcedGas = 0
      # GasPriceMarginFactor is the gas price margin factor
      GasPriceMarginFactor = 1
      # MaxGasPriceLimit is the max gas price limit
      MaxGasPriceLimit = 0
      # StoragePath is the path of the storage
      StoragePath = "/{{PathRWData}}/ethtxmanager-sequencesender.sqlite"
      # ReadPendingL1Txs is a flag to enable the reading of pending L1 txs
      # It can only be enabled if PersistenceFilename is empty
      ReadPendingL1Txs = false
      # SafeStatusL1NumberOfBlocks overwrites the number of blocks to consider a tx as safe
      # overwriting the default value provided by the network
      # 0 means that the default value will be used
      SafeStatusL1NumberOfBlocks = 5
      # FinalizedStatusL1NumberOfBlocks overwrites the number of blocks to consider a tx as finalized
      # overwriting the default value provided by the network
      # 0 means that the default value will be used
      FinalizedStatusL1NumberOfBlocks = 10
      [AggOracle.EVMSender.EthTxManager.Etherman]
        URL = "{{L2URL}}"
        MultiGasProvider = false
        L1ChainID = {{NetworkConfig.L1.L1ChainID}}
        HTTPHeaders = []

[RPC]
  # Host defines the network adapter that will be used to serve the HTTP requests
  Host = "0.0.0.0"
  # Port defines the port to serve the endpoints via HTTP
  Port = 5576
  # ReadTimeout is the HTTP server read timeout
  # check net/http.server.ReadTimeout and net/http.server.ReadHeaderTimeout
  ReadTimeout = "2s"
  # WriteTimeout is the HTTP server write timeout
  # check net/http.server.WriteTimeout
  WriteTimeout = "2s"
  # MaxRequestsPerIPAndSecond defines how much requests a single IP can
  # send within a single second
  MaxRequestsPerIPAndSecond = 10

[ClaimSponsor]
  # DBPath is the path of the database
  DBPath = "/{{PathRWData}}/claimsopnsor.sqlite"
  # Enabled indicates if the sponsor should be run or not
  Enabled = true
  # SenderAddr is the address that will be used to send the claim txs
  SenderAddr = "0xfa3b44587990f97ba8b6ba7e230a5f0e95d14b3d"
  # BridgeAddrL2 is the address of the bridge smart contract on L2
  BridgeAddrL2 = "0xB7098a13a48EcE087d3DA15b2D28eCE0f89819B8"
  # MaxGas is the max gas (limit) allowed for a claim to be sponsored
  MaxGas = 200000
  # RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
  RetryAfterErrorPeriod = "1s"
  # MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
  # Any number smaller than zero will be considered as unlimited retries
  MaxRetryAttemptsAfterError = -1
  # WaitTxToBeMinedPeriod is the period that will be used to ask if a given tx has been mined (or failed)
  WaitTxToBeMinedPeriod = "3s"
  # WaitOnEmptyQueue is the time that will be waited before trying to send the next claim of the queue
  # if the queue is empty
  WaitOnEmptyQueue = "3s"
  # GasOffset is the gas to add on top of the estimated gas when sending the claim txs
  GasOffset = 0
  # EthTxManager is the configuration of the EthTxManager to be used by the claim sponsor
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
  # DBPath path of the DB
  DBPath = "{{PathRWData}}/bridgel1sync.sqlite"
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality = "LatestBlock"
  # InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
  # It should be a number equal or bellow the creation of the bridge contract
  InitialBlockNum = 0
  # BridgeAddr is the address of the bridge smart contract
  BridgeAddr = "{{polygonBridgeAddr}}"
  # SyncBlockChunkSize is the amount of blocks that will be queried to the client on each request
  SyncBlockChunkSize = 100
  # RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
  RetryAfterErrorPeriod = "1s"
  # MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
  # Any number smaller than zero will be considered as unlimited retries
  MaxRetryAttemptsAfterError = -1
  # WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
  WaitForNewBlocksPeriod = "3s"
  # OriginNetwork is the id of the network where the bridge is deployed
  OriginNetwork=0

[BridgeL2Sync]
  # DBPath path of the DB
  DBPath = "{{PathRWData}}/bridgel2sync.sqlite"
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality = "LatestBlock"
  # InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
  InitialBlockNum = 0
  # BridgeAddr is the address of the bridge smart contract
  BridgeAddr = "{{polygonBridgeAddr}}"
  # SyncBlockChunkSize is the amount of blocks that will be queried to the client on each request
  SyncBlockChunkSize = 100
  # RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
  RetryAfterErrorPeriod = "1s"
  # MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
  MaxRetryAttemptsAfterError = -1
  # WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
  WaitForNewBlocksPeriod = "3s"
  # OriginNetwork is the id of the network where the bridge is deployed
  OriginNetwork=1

[LastGERSync]
  # DBPath path of the DB
  DBPath = "{{PathRWData}}/lastgersync.sqlite"
  # BlockFinality indicates the status of the blocks that will be queried in order to sync
  BlockFinality = "LatestBlock"
  #InitialBlockNum is the first block that will be queried when starting the synchronization from scratch.
  # It should be a number equal or bellow the creation of the bridge contract
  InitialBlockNum = 0
  # GlobalExitRootL2Addr is the address of the GER smart contract on L2
  GlobalExitRootL2Addr = "{{L2Config.GlobalExitRootAddr}}"
  # RetryAfterErrorPeriod is the time that will be waited when an unexpected error happens before retry
  RetryAfterErrorPeriod = "1s"
  # MaxRetryAttemptsAfterError is the maximum number of consecutive attempts that will happen before panicing.
  # Any number smaller than zero will be considered as unlimited retries
  MaxRetryAttemptsAfterError = -1
  # WaitForNewBlocksPeriod time that will be waited when the synchronizer has reached the latest block
  WaitForNewBlocksPeriod = "1s"
  # DownloadBufferSize buffer of events to be porcessed. When the buffer limit is reached,
  # downloading will stop until the processing catches up.
  DownloadBufferSize = 100

[NetworkConfig.L1]
  # L1ChainID is the chain id of the L1 network
  L1ChainID = {{L1Config.chainId}}
  # PolAddr is the address of the POL token contract
  PolAddr = "{{L1Config.polTokenAddress}}"
  # ZkEVMAddr is the address of the ZkEVM contract
  ZkEVMAddr = "{{L1Config.polygonZkEVMAddress}}"
  # RollupManagerAddr is the address of the rollup manager contract
  RollupManagerAddr = "{{L1Config.polygonRollupManagerAddress}}"
  # GlobalExitRootManagerAddr is the address of the global exit root manager contract
  GlobalExitRootManagerAddr = "{{L1Config.polygonZkEVMGlobalExitRootAddress}}"

[AggSender]
  # StoragePath is the path of the sqlite db on which the AggSender will store the data
  StoragePath = "{{PathRWData}}/aggsender.sqlite"
  # AggLayerURL is the URL of the AggLayer
  AggLayerURL = "{{AggLayerURL}}"
  # BlockGetInterval is the interval at which the AggSender will get the blocks from L1
  BlockGetInterval = "2s"
  # CheckSettledInterval is the interval at which the AggSender will check if the blocks are settled
  CheckSettledInterval = "2s"
  # AggsenderPrivateKey is the private key which is used to sign certificates
  AggsenderPrivateKey = {Path = "{{SequencerPrivateKeyPath}}", Password = "{{SequencerPrivateKeyPassword}}"}
  # URLRPCL2 is the URL of the L2 RPC node
  URLRPCL2="{{L2URL}}"
  # BlockFinality indicates which finality follows AggLayer
  BlockFinality = "LatestBlock"
  # EpochNotificationPercentage indicates the percentage of the epoch
  # the AggSender should send the certificate
  # 0 -> Begin
  # 50 -> Middle
  EpochNotificationPercentage = 50
  # SaveCertificatesToFilesPath if != "" tells  the AggSender to save the certificates to a file in this path
  SaveCertificatesToFilesPath = ""
`
