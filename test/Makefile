.PHONY: generate-mocks
generate-mocks: generate-mocks-bridgesync generate-mocks-reorgdetector generate-mocks-sequencesender \
	generate-mocks-da generate-mocks-l1infotreesync generate-mocks-helpers \
	generate-mocks-sync generate-mocks-l1infotreesync generate-mocks-aggregator \
	generate-mocks-aggsender generate-mocks-agglayer generate-mocks-bridgesync
	generate-mocks-sync generate-mocks-l1infotreesync generate-mocks-aggregator \
	generate-mocks-aggsender generate-mocks-agglayer generate-mocks-bridgesync

.PHONY: generate-mocks-bridgesync
generate-mocks-bridgesync: ## Generates mocks for bridgesync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../bridgesync --output=../bridgesync --outpkg=bridgesync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	
.PHONY: generate-mocks-reorgdetector
generate-mocks-reorgdetector: ## Generates mocks for reorgdetector, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClient --dir=../reorgdetector --output=../reorgdetector --outpkg=reorgdetector --inpackage --structname=EthClientMock --filename=mock_eth_client.go

COMMON_MOCKERY_PARAMS=--disable-version-string --with-expecter --exported
.PHONY: generate-mocks-sequencesender
generate-mocks-sequencesender: ## Generates mocks for sequencesender, using mockery tool
	rm -Rf ../sequencesender/txbuilder/mocks_txbuilder
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../sequencesender/txbuilder --output ../sequencesender/txbuilder/mocks_txbuilder --outpkg mocks_txbuilder ${COMMON_MOCKERY_PARAMS}
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthTxManager --dir=../sequencesender --output=../sequencesender/mocks --outpkg=mocks --structname=EthTxManagerMock --filename=mock_ethtxmanager.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=Etherman --dir=../sequencesender --output=../sequencesender/mocks --outpkg=mocks --structname=EthermanMock --filename=mock_etherman.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=RPCInterface --dir=../sequencesender --output=../sequencesender/mocks --outpkg=mocks --structname=RPCInterfaceMock --filename=mock_rpc.go

.PHONY: generate-mocks-da
generate-mocks-da: ## Generates mocks for dataavailability, using mockery tool
	rm -Rf ../dataavailability/mocks_da
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../dataavailability --output ../dataavailability/mocks_da --outpkg mocks_da ${COMMON_MOCKERY_PARAMS}
	
.PHONY: generate-mocks-rpc
generate-mocks-rpc: ## Generates mocks for rpc, using mockery tool
	rm -Rf ../rpc/mocks
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../rpc --output ../rpc/mocks --outpkg mocks ${COMMON_MOCKERY_PARAMS}

.PHONY: generate-mocks-l1infotreesync
generate-mocks-l1infotreesync: ## Generates mocks for l1infotreesync, using mockery tool
	rm -Rf ../l1infotreesync/mocks
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../l1infotreesync --output ../l1infotreesync/mocks --outpkg mocks_l1infotreesync ${COMMON_MOCKERY_PARAMS}
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../sync --output=../l1infotreesync/mocks --outpkg=mocks_l1infotreesync --structname=ReorgDetectorMock --filename=mock_reorgdetector.go
	
.PHONY: generate-mocks-aggoracle
generate-mocks-helpers: ## Generates mocks for helpers , using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthTxManager --dir=../aggoracle/chaingersender --output=./helpers --outpkg=helpers --structname=EthTxManagerMock --filename=mock_ethtxmanager.go

.PHONY: generate-mocks-sync
generate-mocks-sync: ## Generates mocks for sync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=evmDownloaderFull --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=EVMDownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go

.PHONY: generate-mocks-aggregator	
generate-mocks-aggregator: ## Generates mocks for aggregator, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ProverInterface --dir=../aggregator --output=../aggregator/mocks --outpkg=mocks --structname=ProverInterfaceMock --filename=mock_prover.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=Etherman --dir=../aggregator --output=../aggregator/mocks --outpkg=mocks --structname=EthermanMock --filename=mock_etherman.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=StateInterface --dir=../aggregator --output=../aggregator/mocks --outpkg=mocks --structname=StateInterfaceMock --filename=mock_state.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=Synchronizer --srcpkg=github.com/0xPolygonHermez/zkevm-synchronizer-l1/synchronizer --output=../aggregator/mocks --outpkg=mocks --structname=SynchronizerInterfaceMock --filename=mock_synchronizer.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthTxManagerClient --dir=../aggregator --output=../aggregator/mocks --outpkg=mocks --structname=EthTxManagerClientMock --filename=mock_eth_tx_manager.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=Tx --srcpkg=github.com/jackc/pgx/v4 --output=../aggregator/mocks --outpkg=mocks --structname=DbTxMock --filename=mock_dbtx.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=RPCInterface --dir=../aggregator --output=../aggregator/mocks --outpkg=mocks --structname=RPCInterfaceMock --filename=mock_rpc.go


.PHONY: generate-mocks-aggsender
generate-mocks-aggsender: ## Generates mocks for aggsender, using mockery tool
	rm -Rf ../aggsender/mocks
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../aggsender --output ../aggsender/mocks --outpkg mocks ${COMMON_MOCKERY_PARAMS}

.PHONY: generate-mocks-agglayer
generate-mocks-agglayer: ## Generates mocks for agglayer, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=AgglayerClientInterface --dir=../agglayer --output=../agglayer --outpkg=agglayer --inpackage --structname=AgglayerClientMock --filename=mock_agglayer_client.go ${COMMON_MOCKERY_PARAMS}

.PHONY: generate-mocks-bridgesync
generate-mocks-bridgesync: ## Generates mocks for bridgesync, using mockery tool
	rm -Rf ../bridgesync/mocks
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../bridgesync --output ../bridgesync/mocks --outpkg mocks_bridgesync ${COMMON_MOCKERY_PARAMS}
	

.PHONY: test-e2e-fork9-validium
test-e2e-fork9-validium: stop
	./run-e2e.sh fork9 cdk-validium
	bats .

.PHONY: test-e2e-fork11-rollup
test-e2e-fork11-rollup: stop
	./run-e2e.sh fork11 rollup
	bats .

.PHONY: test-e2e-fork12-validium
test-e2e-fork12-validium: stop
	./run-e2e.sh fork12 cdk-validium
	bats .

.PHONY: test-e2e-fork12-rollup
test-e2e-fork12-rollup: stop
	./run-e2e.sh fork12 rollup
	bats .

.PHONY: stop
stop:
	kurtosis clean --all


## Help display.
## Pulls comments from beside commands and prints a nicely formatted
## display with the commands and their usage information.
.DEFAULT_GOAL := help

.PHONY: help
help: ## Prints this help
		@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
