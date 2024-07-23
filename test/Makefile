.PHONY: generate-mocks
generate-mocks: 
	$(MAKE) generate-mocks-localbridgesync
	$(MAKE) generate-mocks-reorgdetector
	$(MAKE) generate-mocks-l1infotreesync

.PHONY: generate-mocks-localbridgesync
generate-mocks-localbridgesync: ## Generates mocks for localbridgesync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=downloaderFull --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=DownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go

.PHONY: generate-mocks-reorgdetector
generate-mocks-reorgdetector: ## Generates mocks for reorgdetector, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClient --dir=../reorgdetector --output=../reorgdetector --outpkg=reorgdetector --inpackage --structname=EthClientMock --filename=mock_eth_client.go

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
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../l1infotreesync --output=../l1infotreesync --outpkg=l1infotreesync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=downloaderFull --dir=../l1infotreesync --output=../l1infotreesync --outpkg=l1infotreesync --inpackage --structname=DownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../l1infotreesync --output=../l1infotreesync --outpkg=l1infotreesync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../l1infotreesync --output=../l1infotreesync --outpkg=l1infotreesync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go
