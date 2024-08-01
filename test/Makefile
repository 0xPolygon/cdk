.PHONY: generate-mocks
generate-mocks: generate-mocks-localbridgesync generate-mocks-reorgdetector generate-mocks-sequencesender generate-mocks-da generate-mocks-l1infotreesync generate-mocks-aggoracle generate-mocks-sync


.PHONY: generate-mocks-localbridgesync
generate-mocks-localbridgesync: ## Generates mocks for localbridgesync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=downloaderFull --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=DownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go

.PHONY: generate-mocks-reorgdetector
generate-mocks-reorgdetector: ## Generates mocks for reorgdetector, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClient --dir=../reorgdetector --output=../reorgdetector --outpkg=reorgdetector --inpackage --structname=EthClientMock --filename=mock_eth_client.go

COMMON_MOCKERY_PARAMS=--disable-version-string --with-expecter --exported
.PHONY: generate-mocks-sequencesender
generate-mocks-sequencesender: ## Generates mocks for sequencesender, using mockery tool
	rm -Rf ../sequencesender/txbuilder/mocks_txbuilder
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../sequencesender/txbuilder --output ../sequencesender/txbuilder/mocks_txbuilder --outpkg mocks_txbuilder ${COMMON_MOCKERY_PARAMS}

.PHONY: generate-mocks-da
generate-mocks-da: ## Generates mocks for dataavailability, using mockery tool
	rm -Rf ../dataavailability/mocks_da
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --all --case snake --dir ../dataavailability --output ../dataavailability/mocks_da --outpkg mocks_da ${COMMON_MOCKERY_PARAMS}
	

.PHONY: test-e2e-elderberry-validium
test-e2e-elderberry-validium: stop ## Runs e2e tests checking elderberry/validium
	./run-e2e-seq_sender.sh cdk-validium

.PHONY: test-e2e-elderberry-rollup
test-e2e-elderberry-rollup: stop ## Runs e2e tests checking elderberry/rollup
	./run-e2e-seq_sender.sh rollup

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
.PHONY: generate-mocks-l1infotreesync
generate-mocks-l1infotreesync: ## Generates mocks for l1infotreesync , using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../sync --output=../l1infotreesync --outpkg=l1infotreesync --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go

.PHONY: generate-mocks-aggoracle
generate-mocks-aggoracle: ## Generates mocks for aggoracle , using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthTxManager --dir=../aggoracle/chaingersender --output=../aggoracle --outpkg=aggoracle --structname=EthTxManagerMock --filename=mock_ethtxmanager_test.go

.PHONY: generate-mocks-sync
generate-mocks-sync: ## Generates mocks for sync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=evmDownloaderFull --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=EVMDownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../sync --output=../sync --outpkg=sync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go
