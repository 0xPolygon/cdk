package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"runtime"

	zkevm "github.com/0xPolygon/cdk"
	dataCommitteeClient "github.com/0xPolygon/cdk-data-availability/client"
	jRPC "github.com/0xPolygon/cdk-rpc/rpc"
	"github.com/0xPolygon/cdk/agglayer"
	"github.com/0xPolygon/cdk/aggoracle"
	"github.com/0xPolygon/cdk/aggoracle/chaingersender"
	"github.com/0xPolygon/cdk/aggregator"
	"github.com/0xPolygon/cdk/aggregator/db"
	"github.com/0xPolygon/cdk/aggsender"
	"github.com/0xPolygon/cdk/bridgesync"
	"github.com/0xPolygon/cdk/claimsponsor"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/dataavailability/datacommittee"
	"github.com/0xPolygon/cdk/etherman"
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/l1infotreesync"
	"github.com/0xPolygon/cdk/lastgersync"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/reorgdetector"
	"github.com/0xPolygon/cdk/rpc"
	"github.com/0xPolygon/cdk/sequencesender"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/pgstatestorage"
	"github.com/0xPolygon/cdk/translator"
	ethtxman "github.com/0xPolygon/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman/etherscan"
	"github.com/0xPolygon/zkevm-ethtx-manager/ethtxmanager"
	ethtxlog "github.com/0xPolygon/zkevm-ethtx-manager/log"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/urfave/cli/v2"
)

func start(cliCtx *cli.Context) error {
	c, err := config.Load(cliCtx)
	if err != nil {
		return err
	}

	log.Init(c.Log)

	if c.Log.Environment == log.EnvironmentDevelopment {
		zkevm.PrintVersion(os.Stdout)
		log.Info("Starting application")
	} else if c.Log.Environment == log.EnvironmentProduction {
		logVersion()
	}

	components := cliCtx.StringSlice(config.FlagComponents)
	l1Client := runL1ClientIfNeeded(components, c.Etherman.URL)
	l2Client := runL2ClientIfNeeded(components, getL2RPCUrl(c))
	reorgDetectorL1, errChanL1 := runReorgDetectorL1IfNeeded(cliCtx.Context, components, l1Client, &c.ReorgDetectorL1)
	go func() {
		if err := <-errChanL1; err != nil {
			log.Fatal("Error from ReorgDetectorL1: ", err)
		}
	}()

	reorgDetectorL2, errChanL2 := runReorgDetectorL2IfNeeded(cliCtx.Context, components, l2Client, &c.ReorgDetectorL2)
	go func() {
		if err := <-errChanL2; err != nil {
			log.Fatal("Error from ReorgDetectorL2: ", err)
		}
	}()

	l1InfoTreeSync := runL1InfoTreeSyncerIfNeeded(cliCtx.Context, components, *c, l1Client, reorgDetectorL1)
	claimSponsor := runClaimSponsorIfNeeded(cliCtx.Context, components, l2Client, c.ClaimSponsor)
	l1BridgeSync := runBridgeSyncL1IfNeeded(cliCtx.Context, components, c.BridgeL1Sync, reorgDetectorL1, l1Client)
	l2BridgeSync := runBridgeSyncL2IfNeeded(cliCtx.Context, components, c.BridgeL2Sync, reorgDetectorL2, l2Client)
	lastGERSync := runLastGERSyncIfNeeded(
		cliCtx.Context, components, c.LastGERSync, reorgDetectorL2, l2Client, l1InfoTreeSync,
	)

	for _, component := range components {
		switch component {
		case cdkcommon.SEQUENCE_SENDER:
			c.SequenceSender.Log = c.Log
			seqSender := createSequenceSender(*c, l1Client, l1InfoTreeSync)
			// start sequence sender in a goroutine, checking for errors
			go seqSender.Start(cliCtx.Context)

		case cdkcommon.AGGREGATOR:
			aggregator := createAggregator(cliCtx.Context, *c, !cliCtx.Bool(config.FlagMigrations))
			// start aggregator in a goroutine, checking for errors
			go func() {
				if err := aggregator.Start(); err != nil {
					aggregator.Stop()
					log.Fatal(err)
				}
			}()
		case cdkcommon.AGGORACLE:
			aggOracle := createAggoracle(*c, l1Client, l2Client, l1InfoTreeSync)
			go aggOracle.Start(cliCtx.Context)
		case cdkcommon.RPC:
			server := createRPC(
				c.RPC,
				c.Common.NetworkID,
				claimSponsor,
				l1InfoTreeSync,
				lastGERSync,
				l1BridgeSync,
				l2BridgeSync,
			)
			go func() {
				if err := server.Start(); err != nil {
					log.Fatal(err)
				}
			}()
		case cdkcommon.AGGSENDER:
			aggsender, err := createAggSender(
				cliCtx.Context,
				c.AggSender,
				l1Client,
				l1InfoTreeSync,
				l2BridgeSync,
			)
			if err != nil {
				log.Fatal(err)
			}

			go aggsender.Start(cliCtx.Context)
		}
	}

	waitSignal(nil)

	return nil
}

func createAggSender(
	ctx context.Context,
	cfg aggsender.Config,
	l1EthClient *ethclient.Client,
	l1InfoTreeSync *l1infotreesync.L1InfoTreeSync,
	l2Syncer *bridgesync.BridgeSync) (*aggsender.AggSender, error) {
	logger := log.WithFields("module", cdkcommon.AGGSENDER)
	agglayerClient := agglayer.NewAggLayerClient(cfg.AggLayerURL)
	blockNotifier, err := aggsender.NewBlockNotifierPolling(l1EthClient, aggsender.ConfigBlockNotifierPolling{
		BlockFinalityType:     etherman.BlockNumberFinality(cfg.BlockFinality),
		CheckNewBlockInterval: aggsender.AutomaticBlockInterval,
	}, logger, nil)
	if err != nil {
		return nil, err
	}

	notifierCfg, err := aggsender.NewConfigEpochNotifierPerBlock(agglayerClient, cfg.EpochNotificationPercentage)
	if err != nil {
		return nil, fmt.Errorf("cant generate config for Epoch Notifier because: %w", err)
	}
	epochNotifier, err := aggsender.NewEpochNotifierPerBlock(
		blockNotifier,
		logger,
		*notifierCfg, nil)
	if err != nil {
		return nil, err
	}
	log.Infof("Starting blockNotifier: %s", blockNotifier.String())
	go blockNotifier.Start(ctx)
	log.Infof("Starting epochNotifier: %s", epochNotifier.String())
	go epochNotifier.Start(ctx)
	return aggsender.New(ctx, logger, cfg, agglayerClient, l1InfoTreeSync, l2Syncer, epochNotifier)
}

func createAggregator(ctx context.Context, c config.Config, runMigrations bool) *aggregator.Aggregator {
	logger := log.WithFields("module", cdkcommon.AGGREGATOR)
	// Migrations
	if runMigrations {
		logger.Infof(
			"Running DB migrations host: %s:%s db:%s user:%s",
			c.Aggregator.DB.Host, c.Aggregator.DB.Port, c.Aggregator.DB.Name, c.Aggregator.DB.User,
		)
		runAggregatorMigrations(c.Aggregator.DB)
	}

	// DB
	stateSQLDB, err := db.NewSQLDB(logger, c.Aggregator.DB)
	if err != nil {
		logger.Fatal(err)
	}

	etherman, err := newEtherman(c)
	if err != nil {
		logger.Fatal(err)
	}

	// READ CHAIN ID FROM POE SC

	if c.Aggregator.ChainID == 0 {
		l2ChainID, err := etherman.GetL2ChainID()
		if err != nil {
			logger.Fatal(err)
		}
		log.Infof("Autodiscover L2ChainID: %d", l2ChainID)
		c.Aggregator.ChainID = l2ChainID
	}

	st := newState(&c, c.Aggregator.ChainID, stateSQLDB)

	aggregator, err := aggregator.New(ctx, c.Aggregator, logger, st, etherman)
	if err != nil {
		logger.Fatal(err)
	}

	return aggregator
}

func createSequenceSender(
	cfg config.Config,
	l1Client *ethclient.Client,
	l1InfoTreeSync *l1infotreesync.L1InfoTreeSync,
) *sequencesender.SequenceSender {
	logger := log.WithFields("module", cdkcommon.SEQUENCE_SENDER)

	// Check config
	if cfg.SequenceSender.RPCURL == "" {
		logger.Fatal("Required field RPCURL is empty in sequence sender config")
	}

	ethman, err := etherman.NewClient(ethermanconfig.Config{
		EthermanConfig: ethtxman.Config{
			URL:              cfg.SequenceSender.EthTxManager.Etherman.URL,
			MultiGasProvider: cfg.SequenceSender.EthTxManager.Etherman.MultiGasProvider,
			L1ChainID:        cfg.SequenceSender.EthTxManager.Etherman.L1ChainID,
			Etherscan: etherscan.Config{
				ApiKey: cfg.SequenceSender.EthTxManager.Etherman.Etherscan.ApiKey,
				Url:    cfg.SequenceSender.EthTxManager.Etherman.Etherscan.Url,
			},
			HTTPHeaders: cfg.SequenceSender.EthTxManager.Etherman.HTTPHeaders,
		},
	}, cfg.NetworkConfig.L1Config, cfg.Common)
	if err != nil {
		logger.Fatalf("Failed to create etherman. Err: %w, ", err)
	}

	auth, _, err := ethman.LoadAuthFromKeyStore(cfg.SequenceSender.PrivateKey.Path, cfg.SequenceSender.PrivateKey.Password)
	if err != nil {
		logger.Fatal(err)
	}
	cfg.SequenceSender.SenderAddress = auth.From
	blockFinalityType := etherman.BlockNumberFinality(cfg.SequenceSender.BlockFinality)

	blockFinality, err := blockFinalityType.ToBlockNum()
	if err != nil {
		logger.Fatalf("Failed to create block finality. Err: %w, ", err)
	}
	txBuilder, err := newTxBuilder(cfg, logger, ethman, l1Client, l1InfoTreeSync, blockFinality)
	if err != nil {
		logger.Fatal(err)
	}
	seqSender, err := sequencesender.New(cfg.SequenceSender, logger, ethman, txBuilder)
	if err != nil {
		logger.Fatal(err)
	}

	return seqSender
}

func newTxBuilder(
	cfg config.Config,
	logger *log.Logger,
	ethman *etherman.Client,
	l1Client *ethclient.Client,
	l1InfoTreeSync *l1infotreesync.L1InfoTreeSync,
	blockFinality *big.Int,
) (txbuilder.TxBuilder, error) {
	auth, _, err := ethman.LoadAuthFromKeyStore(cfg.SequenceSender.PrivateKey.Path, cfg.SequenceSender.PrivateKey.Password)
	if err != nil {
		log.Fatal(err)
	}
	da, err := newDataAvailability(cfg, ethman)
	if err != nil {
		log.Fatal(err)
	}
	var txBuilder txbuilder.TxBuilder

	switch contracts.VersionType(cfg.Common.ContractVersions) {
	case contracts.VersionBanana:
		if cfg.Common.IsValidiumMode {
			txBuilder = txbuilder.NewTxBuilderBananaValidium(
				logger,
				ethman.Contracts.Banana.Rollup,
				ethman.Contracts.Banana.GlobalExitRoot,
				da,
				*auth,
				cfg.SequenceSender.MaxBatchesForL1,
				l1InfoTreeSync,
				l1Client,
				blockFinality,
			)
		} else {
			txBuilder = txbuilder.NewTxBuilderBananaZKEVM(
				logger,
				ethman.Contracts.Banana.Rollup,
				ethman.Contracts.Banana.GlobalExitRoot,
				*auth,
				cfg.SequenceSender.MaxTxSizeForL1,
				l1InfoTreeSync,
				l1Client,
				blockFinality,
			)
		}
	case contracts.VersionElderberry:
		if cfg.Common.IsValidiumMode {
			txBuilder = txbuilder.NewTxBuilderElderberryValidium(
				logger, ethman.Contracts.Elderberry.Rollup, da, *auth, cfg.SequenceSender.MaxBatchesForL1,
			)
		} else {
			txBuilder = txbuilder.NewTxBuilderElderberryZKEVM(
				logger, ethman.Contracts.Elderberry.Rollup, *auth, cfg.SequenceSender.MaxTxSizeForL1,
			)
		}
	default:
		err = fmt.Errorf("unknown contract version: %s", cfg.Common.ContractVersions)
	}

	return txBuilder, err
}

func createAggoracle(
	cfg config.Config,
	l1Client,
	l2Client *ethclient.Client,
	syncer *l1infotreesync.L1InfoTreeSync,
) *aggoracle.AggOracle {
	logger := log.WithFields("module", cdkcommon.AGGORACLE)
	var sender aggoracle.ChainSender
	switch cfg.AggOracle.TargetChainType {
	case aggoracle.EVMChain:
		cfg.AggOracle.EVMSender.EthTxManager.Log = ethtxlog.Config{
			Environment: ethtxlog.LogEnvironment(cfg.Log.Environment),
			Level:       cfg.Log.Level,
			Outputs:     cfg.Log.Outputs,
		}
		ethTxManager, err := ethtxmanager.New(cfg.AggOracle.EVMSender.EthTxManager)
		if err != nil {
			log.Fatal(err)
		}
		go ethTxManager.Start()
		sender, err = chaingersender.NewEVMChainGERSender(
			logger,
			cfg.AggOracle.EVMSender.GlobalExitRootL2Addr,
			cfg.AggOracle.EVMSender.SenderAddr,
			l2Client,
			ethTxManager,
			cfg.AggOracle.EVMSender.GasOffset,
			cfg.AggOracle.EVMSender.WaitPeriodMonitorTx.Duration,
		)
		if err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf(
			"Unsupported chaintype %s. Supported values: %v",
			cfg.AggOracle.TargetChainType, aggoracle.SupportedChainTypes,
		)
	}
	aggOracle, err := aggoracle.New(
		logger,
		sender,
		l1Client,
		syncer,
		etherman.BlockNumberFinality(cfg.AggOracle.BlockFinality),
		cfg.AggOracle.WaitPeriodNextGER.Duration,
	)
	if err != nil {
		logger.Fatal(err)
	}

	return aggOracle
}

func newDataAvailability(c config.Config, etherman *etherman.Client) (*dataavailability.DataAvailability, error) {
	if !c.Common.IsValidiumMode {
		return nil, nil
	}
	logger := log.WithFields("module", "da-committee")
	translator := translator.NewTranslatorImpl(logger)
	logger.Infof("Translator rules: %v", c.Common.Translator)
	translator.AddConfigRules(c.Common.Translator)

	// Backend specific config
	daProtocolName, err := etherman.GetDAProtocolName()
	if err != nil {
		return nil, fmt.Errorf("error getting data availability protocol name: %w", err)
	}
	var daBackend dataavailability.DABackender
	switch daProtocolName {
	case string(dataavailability.DataAvailabilityCommittee):
		var (
			pk  *ecdsa.PrivateKey
			err error
		)
		_, pk, err = etherman.LoadAuthFromKeyStore(c.SequenceSender.PrivateKey.Path, c.SequenceSender.PrivateKey.Password)
		if err != nil {
			return nil, err
		}
		dacAddr, err := etherman.GetDAProtocolAddr()
		if err != nil {
			return nil, fmt.Errorf("error getting trusted sequencer URI. Error: %w", err)
		}

		daBackend, err = datacommittee.New(
			logger,
			c.SequenceSender.EthTxManager.Etherman.URL,
			dacAddr,
			pk,
			dataCommitteeClient.NewFactory(),
			translator,
		)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unexpected / unsupported DA protocol: %s", daProtocolName)
	}

	return dataavailability.New(daBackend)
}

func runAggregatorMigrations(c db.Config) {
	runMigrations(c, db.AggregatorMigrationName)
}

func runMigrations(c db.Config, name string) {
	log.Infof("running migrations for %v", name)
	err := db.RunMigrationsUp(c, name)
	if err != nil {
		log.Fatal(err)
	}
}

func newEtherman(c config.Config) (*etherman.Client, error) {
	return etherman.NewClient(ethermanconfig.Config{
		EthermanConfig: ethtxman.Config{
			URL:              c.Aggregator.EthTxManager.Etherman.URL,
			MultiGasProvider: c.Aggregator.EthTxManager.Etherman.MultiGasProvider,
			L1ChainID:        c.Aggregator.EthTxManager.Etherman.L1ChainID,
			HTTPHeaders:      c.Aggregator.EthTxManager.Etherman.HTTPHeaders,
		},
	}, c.NetworkConfig.L1Config, c.Common)
}

func logVersion() {
	log.Infow("Starting application",
		// version is already logged by default
		"gitRevision", zkevm.GitRev,
		"gitBranch", zkevm.GitBranch,
		"goVersion", runtime.Version(),
		"built", zkevm.BuildDate,
		"os/arch", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	)
}

func waitSignal(cancelFuncs []context.CancelFunc) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for sig := range signals {
		switch sig {
		case os.Interrupt, os.Kill:
			log.Info("terminating application gracefully...")

			exitStatus := 0
			for _, cancel := range cancelFuncs {
				cancel()
			}
			os.Exit(exitStatus)
		}
	}
}

func newState(c *config.Config, l2ChainID uint64, sqlDB *pgxpool.Pool) *state.State {
	stateCfg := state.Config{
		DB:      c.Aggregator.DB,
		ChainID: l2ChainID,
	}

	stateDB := pgstatestorage.NewPostgresStorage(stateCfg, sqlDB)

	st := state.NewState(stateCfg, stateDB)

	return st
}

func newReorgDetector(
	cfg *reorgdetector.Config,
	client *ethclient.Client,
) *reorgdetector.ReorgDetector {
	rd, err := reorgdetector.New(client, *cfg)
	if err != nil {
		log.Fatal(err)
	}

	return rd
}

func isNeeded(casesWhereNeeded, actualCases []string) bool {
	for _, actaulCase := range actualCases {
		for _, caseWhereNeeded := range casesWhereNeeded {
			if actaulCase == caseWhereNeeded {
				return true
			}
		}
	}

	return false
}

func runL1InfoTreeSyncerIfNeeded(
	ctx context.Context,
	components []string,
	cfg config.Config,
	l1Client *ethclient.Client,
	reorgDetector *reorgdetector.ReorgDetector,
) *l1infotreesync.L1InfoTreeSync {
	if !isNeeded([]string{cdkcommon.AGGORACLE, cdkcommon.RPC,
		cdkcommon.SEQUENCE_SENDER, cdkcommon.AGGSENDER, cdkcommon.L1INFOTREESYNC}, components) {
		return nil
	}
	l1InfoTreeSync, err := l1infotreesync.New(
		ctx,
		cfg.L1InfoTreeSync.DBPath,
		cfg.L1InfoTreeSync.GlobalExitRootAddr,
		cfg.L1InfoTreeSync.RollupManagerAddr,
		cfg.L1InfoTreeSync.SyncBlockChunkSize,
		etherman.BlockNumberFinality(cfg.L1InfoTreeSync.BlockFinality),
		reorgDetector,
		l1Client,
		cfg.L1InfoTreeSync.WaitForNewBlocksPeriod.Duration,
		cfg.L1InfoTreeSync.InitialBlock,
		cfg.L1InfoTreeSync.RetryAfterErrorPeriod.Duration,
		cfg.L1InfoTreeSync.MaxRetryAttemptsAfterError,
		l1infotreesync.FlagNone,
	)
	if err != nil {
		log.Fatal(err)
	}
	go l1InfoTreeSync.Start(ctx)

	return l1InfoTreeSync
}

func runL1ClientIfNeeded(components []string, urlRPCL1 string) *ethclient.Client {
	if !isNeeded([]string{
		cdkcommon.SEQUENCE_SENDER, cdkcommon.AGGREGATOR,
		cdkcommon.AGGORACLE, cdkcommon.RPC,
		cdkcommon.AGGSENDER,
		cdkcommon.L1INFOTREESYNC,
	}, components) {
		return nil
	}
	log.Debugf("dialing L1 client at: %s", urlRPCL1)
	l1CLient, err := ethclient.Dial(urlRPCL1)
	if err != nil {
		log.Fatalf("failed to create client for L1 using URL: %s. Err:%v", urlRPCL1, err)
	}

	return l1CLient
}

func runL2ClientIfNeeded(components []string, urlRPCL2 string) *ethclient.Client {
	if !isNeeded([]string{cdkcommon.AGGORACLE, cdkcommon.RPC, cdkcommon.AGGSENDER}, components) {
		return nil
	}

	log.Infof("dialing L2 client at: %s", urlRPCL2)
	l2CLient, err := ethclient.Dial(urlRPCL2)
	if err != nil {
		log.Fatal(err)
	}

	return l2CLient
}

func runReorgDetectorL1IfNeeded(
	ctx context.Context,
	components []string,
	l1Client *ethclient.Client,
	cfg *reorgdetector.Config,
) (*reorgdetector.ReorgDetector, chan error) {
	if !isNeeded([]string{
		cdkcommon.SEQUENCE_SENDER, cdkcommon.AGGREGATOR,
		cdkcommon.AGGORACLE, cdkcommon.RPC, cdkcommon.AGGSENDER,
		cdkcommon.L1INFOTREESYNC},
		components) {
		return nil, nil
	}
	rd := newReorgDetector(cfg, l1Client)

	errChan := make(chan error)
	go func() {
		if err := rd.Start(ctx); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	return rd, errChan
}

func runReorgDetectorL2IfNeeded(
	ctx context.Context,
	components []string,
	l2Client *ethclient.Client,
	cfg *reorgdetector.Config,
) (*reorgdetector.ReorgDetector, chan error) {
	if !isNeeded([]string{cdkcommon.AGGORACLE, cdkcommon.RPC, cdkcommon.AGGSENDER}, components) {
		return nil, nil
	}
	rd := newReorgDetector(cfg, l2Client)

	errChan := make(chan error)
	go func() {
		if err := rd.Start(ctx); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	return rd, errChan
}

func runClaimSponsorIfNeeded(
	ctx context.Context,
	components []string,
	l2Client *ethclient.Client,
	cfg claimsponsor.EVMClaimSponsorConfig,
) *claimsponsor.ClaimSponsor {
	if !isNeeded([]string{cdkcommon.RPC}, components) || !cfg.Enabled {
		return nil
	}

	logger := log.WithFields("module", cdkcommon.CLAIM_SPONSOR)
	// In the future there may support different backends other than EVM, and this will require different config.
	// But today only EVM is supported
	ethTxManagerL2, err := ethtxmanager.New(cfg.EthTxManager)
	if err != nil {
		logger.Fatal(err)
	}
	go ethTxManagerL2.Start()
	cs, err := claimsponsor.NewEVMClaimSponsor(
		logger,
		cfg.DBPath,
		l2Client,
		cfg.BridgeAddrL2,
		cfg.SenderAddr,
		cfg.MaxGas,
		cfg.GasOffset,
		ethTxManagerL2,
		cfg.RetryAfterErrorPeriod.Duration,
		cfg.MaxRetryAttemptsAfterError,
		cfg.WaitTxToBeMinedPeriod.Duration,
		cfg.WaitTxToBeMinedPeriod.Duration,
	)
	if err != nil {
		logger.Fatalf("error creating claim sponsor: %s", err)
	}
	go cs.Start(ctx)

	return cs
}

func runLastGERSyncIfNeeded(
	ctx context.Context,
	components []string,
	cfg lastgersync.Config,
	reorgDetectorL2 *reorgdetector.ReorgDetector,
	l2Client *ethclient.Client,
	l1InfoTreeSync *l1infotreesync.L1InfoTreeSync,
) *lastgersync.LastGERSync {
	if !isNeeded([]string{cdkcommon.RPC}, components) {
		return nil
	}
	lastGERSync, err := lastgersync.New(
		ctx,
		cfg.DBPath,
		reorgDetectorL2,
		l2Client,
		cfg.GlobalExitRootL2Addr,
		l1InfoTreeSync,
		cfg.RetryAfterErrorPeriod.Duration,
		cfg.MaxRetryAttemptsAfterError,
		etherman.BlockNumberFinality(cfg.BlockFinality),
		cfg.WaitForNewBlocksPeriod.Duration,
		cfg.DownloadBufferSize,
	)
	if err != nil {
		log.Fatalf("error creating lastGERSync: %s", err)
	}
	go lastGERSync.Start(ctx)

	return lastGERSync
}

func runBridgeSyncL1IfNeeded(
	ctx context.Context,
	components []string,
	cfg bridgesync.Config,
	reorgDetectorL1 *reorgdetector.ReorgDetector,
	l1Client *ethclient.Client,
) *bridgesync.BridgeSync {
	if !isNeeded([]string{cdkcommon.RPC}, components) {
		return nil
	}
	bridgeSyncL1, err := bridgesync.NewL1(
		ctx,
		cfg.DBPath,
		cfg.BridgeAddr,
		cfg.SyncBlockChunkSize,
		etherman.BlockNumberFinality(cfg.BlockFinality),
		reorgDetectorL1,
		l1Client,
		cfg.InitialBlockNum,
		cfg.WaitForNewBlocksPeriod.Duration,
		cfg.RetryAfterErrorPeriod.Duration,
		cfg.MaxRetryAttemptsAfterError,
		cfg.OriginNetwork,
	)
	if err != nil {
		log.Fatalf("error creating bridgeSyncL1: %s", err)
	}
	go bridgeSyncL1.Start(ctx)

	return bridgeSyncL1
}

func runBridgeSyncL2IfNeeded(
	ctx context.Context,
	components []string,
	cfg bridgesync.Config,
	reorgDetectorL2 *reorgdetector.ReorgDetector,
	l2Client *ethclient.Client,
) *bridgesync.BridgeSync {
	if !isNeeded([]string{cdkcommon.RPC, cdkcommon.AGGSENDER}, components) {
		return nil
	}

	bridgeSyncL2, err := bridgesync.NewL2(
		ctx,
		cfg.DBPath,
		cfg.BridgeAddr,
		cfg.SyncBlockChunkSize,
		etherman.BlockNumberFinality(cfg.BlockFinality),
		reorgDetectorL2,
		l2Client,
		cfg.InitialBlockNum,
		cfg.WaitForNewBlocksPeriod.Duration,
		cfg.RetryAfterErrorPeriod.Duration,
		cfg.MaxRetryAttemptsAfterError,
		cfg.OriginNetwork,
	)
	if err != nil {
		log.Fatalf("error creating bridgeSyncL2: %s", err)
	}
	go bridgeSyncL2.Start(ctx)

	return bridgeSyncL2
}

func createRPC(
	cfg jRPC.Config,
	cdkNetworkID uint32,
	sponsor *claimsponsor.ClaimSponsor,
	l1InfoTree *l1infotreesync.L1InfoTreeSync,
	injectedGERs *lastgersync.LastGERSync,
	bridgeL1 *bridgesync.BridgeSync,
	bridgeL2 *bridgesync.BridgeSync,
) *jRPC.Server {
	logger := log.WithFields("module", cdkcommon.RPC)
	services := []jRPC.Service{
		{
			Name: rpc.BRIDGE,
			Service: rpc.NewBridgeEndpoints(
				logger,
				cfg.WriteTimeout.Duration,
				cfg.ReadTimeout.Duration,
				cdkNetworkID,
				sponsor,
				l1InfoTree,
				injectedGERs,
				bridgeL1,
				bridgeL2,
			),
		},
	}

	return jRPC.NewServer(cfg, services, jRPC.WithLogger(logger.GetSugaredLogger()))
}

func getL2RPCUrl(c *config.Config) string {
	if c.AggSender.URLRPCL2 != "" {
		return c.AggSender.URLRPCL2
	}

	return c.AggOracle.EVMSender.URLRPCL2
}
