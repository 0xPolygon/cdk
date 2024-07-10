.PHONY: generate-mocks
generate-mocks: generate-mocks-localbridgesync

.PHONY: generate-mocks-localbridgesync
generate-mocks-localbridgesync: ## Generates mocks for localbridgesync , using mockery tool
	export "GOROOT=$$(go env GOROOT)" && $$(go env GOPATH)/bin/mockery --name=EthClienter --dir=../localbridgesync --output=../localbridgesync --outpkg=localbridgesync --inpackage --structname=L2Mock --filename=mockl2_test.go
