package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	zkevm "github.com/0xPolygon/cdk"
	dataCommitteeClient "github.com/0xPolygon/cdk-data-availability/client"
	"github.com/0xPolygon/cdk/aggregator"
	"github.com/0xPolygon/cdk/aggregator/db"
	"github.com/0xPolygon/cdk/config"
	"github.com/0xPolygon/cdk/dataavailability"
	"github.com/0xPolygon/cdk/dataavailability/datacommittee"
	"github.com/0xPolygon/cdk/etherman"
	ethermanconfig "github.com/0xPolygon/cdk/etherman/config"
	"github.com/0xPolygon/cdk/etherman/contracts"
	"github.com/0xPolygon/cdk/log"
	"github.com/0xPolygon/cdk/sequencesender"
	"github.com/0xPolygon/cdk/sequencesender/txbuilder"
	"github.com/0xPolygon/cdk/state"
	"github.com/0xPolygon/cdk/state/pgstatestorage"
	"github.com/0xPolygon/cdk/translator"
	ethtxman "github.com/0xPolygonHermez/zkevm-ethtx-manager/etherman"
	"github.com/0xPolygonHermez/zkevm-ethtx-manager/etherman/etherscan"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/urfave/cli/v2"
)

func start(cliCtx *cli.Context) error {
	c, err := config.Load(cliCtx, true)
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

	for _, component := range components {
		switch component {
		case SEQUENCE_SENDER:
			c.SequenceSender.Log = c.Log
			seqSender := createSequenceSender(*c)
			// start sequence sender in a goroutine, checking for errors
			go seqSender.Start(cliCtx.Context)

		case AGGREGATOR:
			aggregator := createAggregator(cliCtx.Context, *c, !cliCtx.Bool(config.FlagMigrations))
			// start aggregator in a goroutine, checking for errors
			go func() {
				if err := aggregator.Start(cliCtx.Context); err != nil {
					log.Fatal(err)
				}
			}()
		}
	}

	waitSignal(nil)

	return nil
}

func createAggregator(ctx context.Context, c config.Config, runMigrations bool) *aggregator.Aggregator {
	// Migrations
	if runMigrations {
		log.Infof("Running DB migrations host: %s:%s db:%s user:%s", c.Aggregator.DB.Host, c.Aggregator.DB.Port, c.Aggregator.DB.Name, c.Aggregator.DB.User)
		runAggregatorMigrations(c.Aggregator.DB)
	}

	// DB
	stateSqlDB, err := db.NewSQLDB(c.Aggregator.DB)
	if err != nil {
		log.Fatal(err)
	}

	etherman, err := newEtherman(c)
	if err != nil {
		log.Fatal(err)
	}

	// READ CHAIN ID FROM POE SC
	l2ChainID, err := etherman.GetL2ChainID()
	if err != nil {
		log.Fatal(err)
	}

	st := newState(&c, l2ChainID, stateSqlDB)

	c.Aggregator.ChainID = l2ChainID

	checkAggregatorMigrations(c.Aggregator.DB)

	// Populate Network config
	c.Aggregator.Synchronizer.Etherman.Contracts.GlobalExitRootManagerAddr = c.NetworkConfig.L1Config.GlobalExitRootManagerAddr
	c.Aggregator.Synchronizer.Etherman.Contracts.RollupManagerAddr = c.NetworkConfig.L1Config.RollupManagerAddr
	c.Aggregator.Synchronizer.Etherman.Contracts.ZkEVMAddr = c.NetworkConfig.L1Config.ZkEVMAddr

	aggregator, err := aggregator.New(ctx, c.Aggregator, st, etherman)
	if err != nil {
		log.Fatal(err)
	}

	return aggregator
}

func createSequenceSender(cfg config.Config) *sequencesender.SequenceSender {
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
		log.Fatalf("Failed to create etherman. Err: %w, ", err)
	}

	auth, _, err := ethman.LoadAuthFromKeyStore(cfg.SequenceSender.PrivateKey.Path, cfg.SequenceSender.PrivateKey.Password)
	if err != nil {
		log.Fatal(err)
	}
	cfg.SequenceSender.SenderAddress = auth.From

	txBuilder, err := newTxBuilder(cfg, ethman)
	if err != nil {
		log.Fatal(err)
	}
	seqSender, err := sequencesender.New(cfg.SequenceSender, ethman, txBuilder)
	if err != nil {
		log.Fatal(err)
	}

	return seqSender
}

func newTxBuilder(cfg config.Config, ethman *etherman.Client) (txbuilder.TxBuilder, error) {
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
			txBuilder = txbuilder.NewTxBuilderBananaValidium(ethman.Contracts.Banana.Rollup, ethman.Contracts.Banana.GlobalExitRoot, da, *auth, auth.From, cfg.SequenceSender.MaxBatchesForL1)
		} else {
			txBuilder = txbuilder.NewTxBuilderBananaZKEVM(ethman.Contracts.Banana.Rollup, ethman.Contracts.Banana.GlobalExitRoot, *auth, auth.From, cfg.SequenceSender.MaxTxSizeForL1)
		}
	case contracts.VersionElderberry:
		if cfg.Common.IsValidiumMode {
			txBuilder = txbuilder.NewTxBuilderElderberryValidium(ethman.Contracts.Elderberry.Rollup, da, *auth, auth.From, cfg.SequenceSender.MaxBatchesForL1)
		} else {
			txBuilder = txbuilder.NewTxBuilderElderberryZKEVM(ethman.Contracts.Elderberry.Rollup, *auth, auth.From, cfg.SequenceSender.MaxTxSizeForL1)
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
	translator := translator.NewTranslatorImpl()
	log.Infof("Translator rules: %v", c.Common.Translator)
	translator.AddConfigRules(c.Common.Translator)

	// Backend specific config
	daProtocolName, err := etherman.GetDAProtocolName()
	if err != nil {
		return nil, fmt.Errorf("error getting data availability protocol name: %v", err)
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
			return nil, fmt.Errorf("error getting trusted sequencer URI. Error: %v", err)
		}

		daBackend, err = datacommittee.New(
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

func checkAggregatorMigrations(c db.Config) {
	err := db.CheckMigrations(c, db.AggregatorMigrationName)
	if err != nil {
		log.Fatal(err)
	}
}

func runMigrations(c db.Config, name string) {
	log.Infof("running migrations for %v", name)
	err := db.RunMigrationsUp(c, name)
	if err != nil {
		log.Fatal(err)
	}
}

func newEtherman(c config.Config) (*etherman.Client, error) {
	config := ethermanconfig.Config{
		URL: c.Aggregator.EthTxManager.Etherman.URL,
	}
	return etherman.NewClient(config, c.NetworkConfig.L1Config, c.Common)
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

	stateDb := pgstatestorage.NewPostgresStorage(stateCfg, sqlDB)

	st := state.NewState(stateCfg, stateDb)
	return st
}
