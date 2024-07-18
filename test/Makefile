.PHONY: generate-mocks
generate-mocks: 
	$(MAKE) generate-mocks-localbridgesync
	$(MAKE) generate-mocks-reorgdetector

.PHONY: generate-mocks-localbridgesync
generate-mocks-localbridgesync: ## Generates mocks for localbridgesync, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=L2Mock --filename=mock_l2_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=downloaderFull --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=DownloaderMock --filename=mock_downloader_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=processorInterface --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ProcessorMock --filename=mock_processor_test.go
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=ReorgDetector --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=ReorgDetectorMock --filename=mock_reorgdetector_test.go

.PHONY: generate-mocks-reorgdetector
generate-mocks-reorgdetector: ## Generates mocks for reorgdetector, using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClient --dir=../reorgdetector --output=../reorgdetector --outpkg=reorgdetector --inpackage --structname=EthClientMock --filename=mock_eth_client.go