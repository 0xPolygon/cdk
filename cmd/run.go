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
	"github.com/0xPolygon/cdk/aggregator"
	"github.com/0xPolygon/cdk/aggregator/db"
	cdkcommon "github.com/0xPolygon/cdk/common"
	"github.com/0xPolygon/cdk/config"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/dataavailability/datacommittee"
	"github.com/0xPolygon/cdk/etherman"
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/sequencesender"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/translator"
	ethtxman "github.com/0xPolygon/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygon/zkevm-ethtx-manager/etherman/etherscan"
	aggkitetherman "github.com/agglayer/aggkit/etherman"
	"github.com/agglayer/aggkit/l1infotreesync"
	"github.com/agglayer/aggkit/log"
	"github.com/agglayer/aggkit/reorgdetector"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli/v2"
)

func start(cliCtx *cli.Context) error {
	cfg, err := config.Load(cliCtx)
	if err != nil {
		return err
	}

	log.Init(cfg.Log)

	if cfg.Log.Environment == log.EnvironmentDevelopment {
		zkevm.PrintVersion(os.Stdout)
		log.Info("Starting application")
	} else if cfg.Log.Environment == log.EnvironmentProduction {
		logVersion()
	}

	components := cliCtx.StringSlice(config.FlagComponents)
	l1Client := runL1ClientIfNeeded(components, cfg.Etherman.URL)
	reorgDetectorL1, errChanL1 := runReorgDetectorL1IfNeeded(cliCtx.Context, components, l1Client, &cfg.ReorgDetectorL1)
	go func() {
		if err := <-errChanL1; err != nil {
			log.Fatal("Error from ReorgDetectorL1: ", err)
		}
	}()

	l1InfoTreeSync := runL1InfoTreeSyncerIfNeeded(cliCtx.Context, components, *cfg, l1Client, reorgDetectorL1)
	var rpcServices []jRPC.Service
	for _, component := range components {
		switch component {
		case cdkcommon.SEQUENCE_SENDER:
			cfg.SequenceSender.Log = cfg.Log
			seqSender := createSequenceSender(*cfg, l1Client, l1InfoTreeSync)
			// start sequence sender in a goroutine, checking for errors
			go seqSender.Start(cliCtx.Context)

		case cdkcommon.AGGREGATOR:
			aggregator := createAggregator(cliCtx.Context, *cfg, !cliCtx.Bool(config.FlagMigrations))
			// start aggregator in a goroutine, checking for errors
			go func() {
				if err := aggregator.Start(); err != nil {
					aggregator.Stop()
					log.Fatal(err)
				}
			}()
		}
	}
	if len(rpcServices) > 0 {
		rpcServer := createRPC(cfg.RPC, rpcServices)
		go func() {
			if err := rpcServer.Start(); err != nil {
				log.Fatal(err)
			}
		}()
	}
	waitSignal(nil)

	return nil
}

func createAggregator(ctx context.Context, c config.Config, runMigrations bool) *aggregator.Aggregator {
	logger := log.WithFields("module", cdkcommon.AGGREGATOR)
	// Migrations
	if runMigrations {
		logger.Infof("Running DB migrations. File %s", c.Aggregator.DBPath)
		runAggregatorMigrations(c.Aggregator.DBPath)
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

	aggregator, err := aggregator.New(ctx, c.Aggregator, logger, etherman)
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

func runAggregatorMigrations(dbPath string) {
	runMigrations(dbPath, db.AggregatorMigrationName)
}

func runMigrations(dbPath string, name string) {
	log.Infof("running migrations for %v", name)
	err := db.RunMigrationsUp(dbPath, name)
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

func newReorgDetector(
	cfg *reorgdetector.Config,
	client *ethclient.Client,
	network reorgdetector.Network,
) *reorgdetector.ReorgDetector {
	rd, err := reorgdetector.New(client, *cfg, network)
	if err != nil {
		log.Fatal(err)
	}

	return rd
}

func isNeeded(casesWhereNeeded, actualCases []string) bool {
	for _, actualCase := range actualCases {
		for _, caseWhereNeeded := range casesWhereNeeded {
			if actualCase == caseWhereNeeded {
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
	if !isNeeded([]string{cdkcommon.SEQUENCE_SENDER, cdkcommon.L1INFOTREESYNC}, components) {
		return nil
	}
	l1InfoTreeSync, err := l1infotreesync.New(
		ctx,
		cfg.L1InfoTreeSync.DBPath,
		cfg.L1InfoTreeSync.GlobalExitRootAddr,
		cfg.L1InfoTreeSync.RollupManagerAddr,
		cfg.L1InfoTreeSync.SyncBlockChunkSize,
		aggkitetherman.NewBlockNumberFinality(cfg.L1InfoTreeSync.BlockFinality),
		reorgDetector,
		l1Client,
		cfg.L1InfoTreeSync.WaitForNewBlocksPeriod.Duration,
		cfg.L1InfoTreeSync.InitialBlock,
		cfg.L1InfoTreeSync.RetryAfterErrorPeriod.Duration,
		cfg.L1InfoTreeSync.MaxRetryAttemptsAfterError,
		l1infotreesync.FlagNone,
		aggkitetherman.FinalizedBlock,
	)
	if err != nil {
		log.Fatal(err)
	}
	go l1InfoTreeSync.Start(ctx)

	return l1InfoTreeSync
}

func runL1ClientIfNeeded(components []string, urlRPCL1 string) *ethclient.Client {
	if !isNeeded([]string{
		cdkcommon.SEQUENCE_SENDER, cdkcommon.AGGREGATOR, cdkcommon.L1INFOTREESYNC,
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

func getRollupID(networkConfig ethermanconfig.L1Config,
	l1Client *ethclient.Client) uint32 {
	rollupID, err := etherman.GetRollupID(networkConfig, networkConfig.ZkEVMAddr, l1Client)
	if err != nil {
		log.Fatal(err)
	}
	return rollupID
}

func runReorgDetectorL1IfNeeded(
	ctx context.Context,
	components []string,
	l1Client *ethclient.Client,
	cfg *reorgdetector.Config,
) (*reorgdetector.ReorgDetector, chan error) {
	if !isNeeded([]string{cdkcommon.SEQUENCE_SENDER, cdkcommon.AGGREGATOR, cdkcommon.L1INFOTREESYNC},
		components) {
		return nil, nil
	}
	rd := newReorgDetector(cfg, l1Client, reorgdetector.L1)

	errChan := make(chan error)
	go func() {
		if err := rd.Start(ctx); err != nil {
			errChan <- err
		}
		close(errChan)
	}()

	return rd, errChan
}

func createRPC(cfg jRPC.Config, services []jRPC.Service) *jRPC.Server {
	logger := log.WithFields("module", "RPC")
	return jRPC.NewServer(cfg, services, jRPC.WithLogger(logger.GetSugaredLogger()))
}

func getL2RPCUrl(c *config.Config) string {
	if c.SequenceSender.RPCURL != "" {
		return c.SequenceSender.RPCURL
	}

	return ""
}
